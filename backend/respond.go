package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dghubble/go-twitter/twitter"
)

// respondToInvocation receives an incoming tweet from a stream, and will respond to it with a link to donate via the
// Charity Yeti website. The donation link includes an id for a Mongo document for the front end to retrieve and add
// on the donation value after a successful donation.
func respondToInvocation(yeti yetiInvokedData) error {
	if yeti.honorary.ScreenName != "" {
		dataID := primitive.NewObjectID()
		donateLink := fmt.Sprintf("https://charityyeti.com?id=%v", dataID.Hex())
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", yeti.invoker.ScreenName, yeti.honorary.ScreenName, donateLink)

		params := twitter.StatusUpdateParams{InReplyToStatusID: yeti.invokerTweetID}

		if sendResponses {
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
			_, _, err = twitterClient.Statuses.Update(tweetText, &params)
			if err != nil {
				return err
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

// respondToDonation gets called after a successful donation. It parses the data sent from the Charity Yeti front end
// client to make sure that our responses get sent to the original invocation tweet
func respondToDonation(tweet successfulDonationData) error {
	tweetText := fmt.Sprintf("Good news, %v! %v thought your tweet was so great, they donated $%v to Partner's in Health on your behalf! See https://charityyeti.com for details.", tweet.honorary, tweet.invoker, tweet.donationValue)
	log.Debugf(fmt.Sprintf("Tweet to send: %+v", tweetText))
	log.Debugf(fmt.Sprintf("Responding to: %v", tweet.invokerTweetID))

	params := &twitter.StatusUpdateParams{
		InReplyToStatusID: tweet.invokerTweetID,
	}

	if sendResponses {
		log.Infow("Actually sending this!")
		_, _, err := twitterClient.Statuses.Update(tweetText, params)

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

// processInvocation parses an incoming tweet from the tweetQueue, pulls out the user who sent it, the ID of the
// originating tweet, and passes it to respondToInvocation to send to Twitter
func processInvocation() {

	// loop forever to listen for incoming tweets
	for {
		// when a tweet gets received from the queue, start processing
		incomingTweet := <-tweetQueue

		honorary := getInReplyToTwitterUser(incomingTweet.InReplyToScreenName)

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
