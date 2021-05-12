package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const BAD_ROBOT = "twitter: 226 This request looks like it might be automated. To protect our users from spam and other malicious activity, we can't complete this action right now. Please try again later."

// processInvocation parses an incoming tweet from the tweetQueue, pulls out the user who sent it, the ID of the
// originating tweet, and passes it to respondToInvocation to send to Twitter
func processInvocation() {

	// loop forever to listen for incoming tweets
	for {
		select {
		// when a tweet gets received from the queue, start processing
		case incomingTweet := <-tweetQueue:
			// check and make sure this is specifically invoking us and not
			// just replying or randomly @'ing us
			if !strings.Contains(strings.ToLower(incomingTweet.TweetCreateEvents[0].Text), "hey @charityetidev") {
				// this isn't a specific invokation so going to ignore it
				log.Infof("this isn't an invocation so we are going to ignore it: %v", incomingTweet.TweetCreateEvents[0].Text)
				break
			}

			honorary := getInReplyToTwitterUser(int64(incomingTweet.TweetCreateEvents[0].InReplyToUserID))

			yeti := yetiInvokedData{
				invoker:         &incomingTweet.TweetCreateEvents[0].User,
				honorary:        honorary,
				invokerTweetID:  incomingTweet.TweetCreateEvents[0].ID,
				originalTweetID: incomingTweet.TweetCreateEvents[0].InReplyToStatusID,
			}

			err := respondToInvocation(yeti)
			if err != nil {
				log.Error(err)
			}
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

	// now start sticking them together but check and make sure our length is less than the 240 character requirement
	tweetText := fmt.Sprintf("%v %v %v %v\nReply 'STOP' to opt out.", greetings[greetingsIdx], thanks[thanksIdx], callToAction[callToActionIdx], link)
	if len(tweetText) > 240 {
		log.Errorf("Tweet/DM text is too long. Cannot exceed 240 but we made a %v character long string", len(tweetText))
		return generateResponseTweetText(link)
	} else {
		return tweetText
	}

}

// respondToInvocation receives an incoming tweet from a stream, and will respond to it with a link to donate via the
// Charity Yeti website. The donation link includes an id for a Mongo document for the front end to retrieve and add
// on the donation value after a successful donation.
func respondToInvocation(yeti yetiInvokedData) error {
	if existsInBlockList(strconv.Itoa(int(yeti.invoker.ID))) {
		// this user asked us not to contact them so we'll skip over
		return errors.New(fmt.Sprintf("user %v (%v) asked us not to contact them", yeti.invoker.ScreenName, yeti.invoker.ID))
	}
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

// generateSuccessfulDonationTweetText mixes and matches response words to generate some different, human sounding phrases
// to make us look less spammy
func generateSuccessfulDonationTweetText(invoker string, donation float32) string {
	// our message strings we're gonna mix and match
	greetings := []string{
		"Good news!",
		"Now this is exciting!",
		"World suck will decrease because of you!",
		"Congratulations!",
		"Surpise! Some good news coming your way -",
		"*Excited Yeti Noises*",
	}
	thanks := []string{
		fmt.Sprintf("Thanks to this extremely excellent tweet @%v donated $%v to Partner's in Health!", invoker, donation),
		fmt.Sprintf("@%v thought your tweet was so great, they donated $%v to Partner's in Health to celebrate!", invoker, donation),
		fmt.Sprintf("@%v loved your tweet so much they gave $%v to Partner's in Health to show some gratitude!", invoker, donation),
		fmt.Sprintf("@%v donated $%v to Partner's because your tweet was THAT GOOD.", invoker, donation),
		fmt.Sprintf("Partner's In Health has $%v extra thanks to this awesome tweet that @%v loved so much.", donation, invoker),
	}

	congrats := []string{
		"Congrats!",
		"Great job!",
		"Way to go!",
		"Thank you!",
		"Keep it up!",
		"How neat is that!",
	}

	// grab random index from each
	source := rand.NewSource(time.Now().Unix())
	randomizer := rand.New(source) // initialize local pseudorandom generator
	greetingsIdx := randomizer.Intn(len(greetings))
	thanksIdx := randomizer.Intn(len(thanks))
	congratsIdx := randomizer.Intn(len(congrats))

	// now start sticking them together but check and make sure our length is less than the 240 character requirement
	tweetText := fmt.Sprintf("%v %v %v", greetings[greetingsIdx], thanks[thanksIdx], congrats[congratsIdx])
	if len(tweetText) > 240 {
		log.Errorf("Tweet text is too long. Cannot exceed 240 but we made a %v character long string", len(tweetText))
		return generateSuccessfulDonationTweetText(invoker, donation)
	} else {
		return tweetText
	}

}

// respondToDonation gets called after a successful donation. It parses the data sent from the Charity Yeti front end
// client to make sure that our responses get sent to the original invocation tweet
func respondToDonation(tweet successfulDonationData) error {
	tweetText := generateSuccessfulDonationTweetText(tweet.invoker, tweet.donationValue)
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

// replyToDM DMs the user who asked us to stop/start contacting them that we will do so, or
// let them know that we can't quite understand what they mean
func replyToDM(userID, dmText string) error {
	// send a DM
	_, resp, err := twitterClient.DirectMessages.EventsNew(&twitter.DirectMessageEventsNewParams{
		Event: &twitter.DirectMessageEvent{
			Type: "message_create",
			Message: &twitter.DirectMessageEventMessage{
				Target: &twitter.DirectMessageTarget{
					RecipientID: userID,
				},
				Data: &twitter.DirectMessageData{
					Text: dmText,
				},
			},
		},
	})

	if err != nil {
		if err.Error() == BAD_ROBOT || resp.StatusCode == http.StatusTooManyRequests {
			log.Error("twitter is limiting our ability to send messages, we need to take a break")
			// there's a x-rate-limit-reset header on the response that tells us how long
			// until rates are reset so we wait that long
			waitUntilString := resp.Header.Get("x-rate-limit-reset")
			// try and marshal the wait-until into an int
			waitUntil, err := strconv.Atoi(waitUntilString)
			if err != nil {
				log.Info("couldn't get a time until the rates are reset")
				waitUntil = 15 // so we're going to wait 15 minutes, the longest time before time resets
			}
			time.Sleep(time.Duration(waitUntil) * time.Minute)
			return err
		}
		log.Errorf("Could not send a DM to user %v: %v", userID, err)
		return err
	}
	return nil
}

// processDM parses an incoming direct message from the DM webhook, pulls out the user who sent it, the ID of the
// originating tweet, and passes it to respondToInvocation to send to Twitter
func processDM() {
	// loop forever to listen for incoming DMs
	for {
		select {
		// when a DM gets received from the queue, start processing
		case incomingMessage := <-dmQueue:
			log.Info("incoming message on DM Queue")
			confirmBlock := "Thanks for lettings us know you don't want us contacting you. If you change your mind, you can reply with START."
			confirmUnblock := "*Excited Yeti Noises* We're so glad you want to use Charity Yeti again! You're good to go from here. If you change your mind, you can reply with STOP."
			unknownMessage := "Charity Yeti is a work in progress and isn't too smart yet. If you want us to leave you alone, reply with STOP."
			if len(incomingMessage.DirectMessageEvents) == 0 {
				log.Info("no events on webhook")
				break
			}
			senderId := incomingMessage.DirectMessageEvents[0].MessageCreate.SenderID
			log.Infof("message from %v", senderId)
			if senderId == cfg.CharityYetiId {
				// this is a message _we_ generated
				log.Info("This is a message _from_ Charity Yeti, returning")
				break
			}
			action := "default"
			check := incomingMessage.DirectMessageEvents[0].MessageCreate.MessageData.Text
			if strings.Contains(strings.ToLower(check), "stop") {
				action = "stop"
			}
			if strings.Contains(strings.ToLower(check), "start") {
				action = "start"
			}

			switch action {
			case "stop":
				// add this user to our block list so we don't contact them again
				log.Info("This user wants us to stop contacting them")
				// let the user know that we won't contact them again
				if err := replyToDM(senderId, confirmBlock); err != nil {
					log.Errorf("Could not DM user %v affirming no contact: %v", senderId, err)
				}
				// process the add to block list
				if err := addBlockList(senderId); err != nil {
					log.Errorf("Could not add this user to the block list. Manually add user %v to block list: %v", senderId, err)
				}
			case "start":
				// user was previously removed but is fine with us contacting them again
				log.Info("This user is allowing us to contact them again")
				// let the user know that they can use Charity Yeti again
				if err := replyToDM(senderId, confirmUnblock); err != nil {
					log.Errorf("Could not DM user %v affirming consent to contact: %v", senderId, err)
				}
				// remove this user from the block list
				if err := removeBlockList(senderId); err != nil {
					log.Errorf("Could not remove this user from the block list. Manually remove user %v from block list: %v", senderId, err)
				}
			default:
				// we don't know what this user wants
				log.Info("Received a DM without a keyword")
				// first check and make sure we aren't contacting them if they don't want us to
				if existsInBlockList(senderId) {
					// they don't want us contacting them
					log.Info("user is in block list and doesn't want us to contact them")
				} else {
					// respond saying we aren't quite that clever (yet!)
					if err := replyToDM(senderId, unknownMessage); err != nil {
						log.Errorf("Could not DM user %v with ambiguous message: %v", senderId, err)
					}
				}
			}
		}
	}
}
