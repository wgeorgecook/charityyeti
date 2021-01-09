package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dghubble/go-twitter/twitter"
)

// processInvocation parses an incoming tweet from the tweetQueue, pulls out the user who sent it, the ID of the
// originating tweet, and passes it to respondToInvocation to send to Twitter
func processInvocation() {

	// loop forever to listen for incoming tweets
	for {
		// when a tweet gets received from the queue, start processing
		incomingTweet := <-tweetQueue

		honorary := getInReplyToTwitterUser(incomingTweet.InReplyToUserID)

		yeti := yetiInvokedData{
			invoker:         incomingTweet.User,
			honorary:        honorary,
			invokerTweetID:  incomingTweet.ID,
			originalTweetID: incomingTweet.InReplyToStatusID,
		}

		err := respondToInvocation(yeti)
		if err != nil {
			log.Error(err)
		}
	}

}

// generateResponseTweetText mixes and matches response words to generate some different, human sounding phrases
// to make us look less spammy
func generateResponseTweetText(link string) string {
	// our message strings we're gonna mix and match
	greetings := []string{
		"Hey there!",
		"Hello!",
		"Hi!",
		"Glad you reached out!",
		"Howdy!",
		"*Excited Yeti Noises*",
	}
	thanks := []string{
		"Thanks for reaching out.",
		"Glad you tagged us.",
		"We're stoked for this awesome tweet.",
		"Thanks for wanting to help out!",
		"You clearly did not forget to be awesome today.",
	}
	callToAction := []string{
		"Here's a personalized link:",
		"One hot and fresh donation link coming up:",
		"Here's a unique link on Charity Yeti just for you:",
		"You can find your personal Charity Yeti here:",
	}

	// grab random index from each
	source := rand.NewSource(time.Now().Unix())
	randomizer := rand.New(source) // initialize local pseudorandom generator
	greetingsIdx := randomizer.Intn(len(greetings))
	thanksIdx := randomizer.Intn(len(thanks))
	callToActionIdx := randomizer.Intn(len(callToAction))

	// now start sticking them together
	return fmt.Sprintf("%v %v %v %v\nReply 'STOP' to opt out.", greetings[greetingsIdx], thanks[thanksIdx], callToAction[callToActionIdx], link)

}

// respondToInvocation receives an incoming tweet from a stream, and will respond to it with a link to donate via the
// Charity Yeti website. The donation link includes an id for a Mongo document for the front end to retrieve and add
// on the donation value after a successful donation.
func respondToInvocation(yeti yetiInvokedData) error {
	if yeti.honorary.ScreenName != "" {
		dataID := primitive.NewObjectID()
		donateLink := fmt.Sprintf("https://charityyeti.casadecook.com?id=%v", dataID.Hex()) // TODO: change this to production
		tweetText := generateResponseTweetText(donateLink)

		if cfg.SendTweets {
			// create the record in Mongo
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			data := bson.M{
				"_id":             dataID,
				"invoker":         yeti.invoker,
				"honorary":        yeti.honorary,
				"invokerTweetID":  yeti.invokerTweetID,
				"originalTweetID": yeti.originalTweetID,
			}
			log.Infow("Creating mongo document")
			collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)
			_, err := collection.InsertOne(ctx, data)

			if err != nil {
				log.Errorf(fmt.Sprintf("could not create Mongo document: %v", err))
			}

			// send the tweet
			log.Infow("Actually sending this!")

			// send a DM
			_, _, err = twitterClient.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
				Event: &twitter.DirectMessageEvent{
					Type: "message_create",
					Message: &twitter.DirectMessageEventMessage{
						Target: &twitter.DirectMessageTarget{
							RecipientID: yeti.invoker.IDStr,
						},
						Data: &twitter.DirectMessageData{
							Text: tweetText,
						},
					},
				},
			})

			if err != nil {
				log.Errorf("Could not send a DM: %v", err)
				// if we can't send a DM (like they have DMs off or something), we fall back on a good old fashioned tweet reply
				params := twitter.StatusUpdateParams{InReplyToStatusID: yeti.invokerTweetID}
				responseTweet, _, err := twitterClient.Statuses.Update(tweetText, &params)
				if err != nil {
					return err
				}

				// now that we have a response tweet, we need to save it's ID back to the db so we can reply to this later
				update := bson.M{
					"$set": bson.M{"invokerResponseTweetID": &responseTweet.ID},
				}

				filter := bson.M{"_id": dataID}

				_, err = collection.UpdateOne(ctx, filter, update)
				if err != nil {
					log.Errorf("could not update document with this responded tweet ID: %v", err)
					return err
				}
			}
		}

		log.Info(tweetText)

		return nil
	}

	/* TODO:
	Right now, Charity Yeti only works if the invoker is **responding** to a tweet. We can't properly handle a case
	where a user retweets with comment because there's not a tweet.InResponseTo attribute. Having this in response
	to attribute is the only mechanism we presently have to detect and track *who* the invoker wants to credit their
	donation for. There may be other attributes (I haven't looked into what data we can get from a retweeted tweet,
	but it is probably similar), but we should decide if we want to interact with both replies and retweeted tweets.

	See issue #4 for discussion.
	*/
	return errors.New("no honorary to respond to")
}

// goodDonation gets called when BrainTree returns an OK transaction back to us from the frontend
// and we know a donation was processed successfully. It's responsible for updating the Mongo document
// with the donation value, and then sends a tweet letting the original tweeter that
// someone donated because of their tweets.
func goodDonation(c charityYetiData) error {
	log.Info("Good donation received - responding to it")

	// set the values for a successfulDonationData struct
	tweet := successfulDonationData{
		invoker:                c.Invoker.ScreenName,
		honorary:               c.Honorary.ScreenName,
		donationValue:          c.DonationValue,
		invokerTweetID:         c.InvokerTweetID,
		originalTweetID:        c.OriginalTweetID,
		invokerResponseTweetID: c.InvokerResponseTweetID,
	}

	// update the Mongo document
	if _, err := updateDocument(c); err != nil {
		log.Errorf("Could not update Mongo after a good donation: %v", err)
		// I don't want to return here becuase the donation was still successful
		// and we want to spread awareness
		// TODO: some sort of backup for this so we have record
	}

	log.Info(fmt.Sprintf(
		"{Data: { invoker: %v, honorary: %v, invokerTweetID: %v, originalTweetID: %v, donationValue: %v}}",
		tweet.invoker, tweet.honorary, tweet.invokerTweetID, tweet.originalTweetID, tweet.donationValue))

	if err := respondToDonation(tweet); err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// respondToDonation gets called after a successful donation. It parses the data sent from the Charity Yeti front end
// client to make sure that our responses get sent to the original invocation tweet
func respondToDonation(tweet successfulDonationData) error {
	tweetText := fmt.Sprintf("Good news! @%v thought your tweet was so great, they donated $%v to Partner's in Health on your behalf! See https://charityyeti.com for details.", tweet.invoker, tweet.donationValue)
	log.Debugf(fmt.Sprintf("Tweet to send: %+v", tweetText))
	log.Debugf(fmt.Sprintf("Responding to: %v", tweet.invokerTweetID))

	var params twitter.StatusUpdateParams
	if tweet.invokerResponseTweetID != 0 {
		// We couldn't DM this person, so we need to respond on our tweet with the donation link
		params = twitter.StatusUpdateParams{
			InReplyToStatusID: tweet.invokerResponseTweetID,
		}
	} else {
		// This was from a DM, so we need to respond on the invoker's tweet
		params = twitter.StatusUpdateParams{
			InReplyToStatusID: tweet.invokerTweetID,
		}
	}

	if cfg.SendTweets {
		log.Infow("Actually sending this!")
		_, _, err := twitterClient.Statuses.Update(tweetText, &params)

		if err != nil {
			return err
		}

		// TODO: this needs testing
		if retweetGoods {
			log.Infow("We're retweeting the invoked tweet. We might break twitter TOS for this.")
			rtParams := &twitter.StatusRetweetParams{ID: tweet.originalTweetID}
			_, _, err := twitterClient.Statuses.Retweet(tweet.originalTweetID, rtParams)
			if err != nil {
				log.Errorf("Could not retweet: %v", err)
			}
		}
	}
	return nil
}
