package main

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/dghubble/go-twitter/twitter"
)

// respondToInvocation receives an incoming tweet from a stream,
// and will respond to it with a link to donate to either
// Foundation To Decrease World Suck or directly at Parners
// In Health
func respondToInvocation(yeti yetiInvokedData) error {
	if yeti.honorary.ScreenName != "" {
		dataID := primitive.NewObjectID()
		donateLink := fmt.Sprintf("https://charityyeti.com?id=%v", dataID.Hex())
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", yeti.invoker.ScreenName, yeti.honorary.ScreenName, donateLink)

		params := twitter.StatusUpdateParams{InReplyToStatusID: yeti.invokerTweetID}

		if sendResponses {
			// create the record in Mongo
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			data := bson.M{
				"_id":              dataID,
				"invoker":          yeti.invoker,
				"honorary":         yeti.honorary,
				"invokerTweetID":   yeti.invokerTweetID,
				"originalTweetID":  yeti.originalTweetID,
			}
			log.Infow("Creating mongo document")
			collection := mongoClient.Database("charityyeti-test").Collection("twitterData")
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

	return errors.New("no honorary to respond to")
}

// respondToDonation gets called after a successful donation. It parses the
// data sent from the server to make sure that our responses get sent to the
// original invocation tweet
func respondToDonation(tweet successfulDonationData) error {
	tweetText := fmt.Sprintf("Good news, %v! %v thought your tweet was so great, they donated $%v to Partner's in Health on your behalf! See https://charityyeti.com for details.", tweet.honorary, tweet.invoker, tweet.donationValue)
	log.Debugf(fmt.Sprintf("Tweet to send: %+v", tweetText))
	log.Debugf(fmt.Sprintf("Responding to: %v", tweet.invokerTweetID))

	params := &twitter.StatusUpdateParams{
		InReplyToStatusID:  tweet.invokerTweetID,
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

// generateResponse parses the tweet from the demux stream, pulls out
// the user who sent it, the ID of the originating tweet, and
// passes it to respondToInvocation to send to Twitter
func generateResponse(incomingTweet *twitter.Tweet) error {

	honorary := getInReplyToTwitterUser(incomingTweet.InReplyToScreenName)

	yeti := yetiInvokedData{
		invoker:        incomingTweet.User,
		honorary:       honorary,
		invokerTweetID: incomingTweet.ID,
		originalTweetID: incomingTweet.InReplyToStatusID,
	}

	err := respondToInvocation(yeti)
	if err != nil {
		return err
	}

	return nil
}


// getInReplyToTwitterUser takes the screen name of a Twitter user (IE - the honorary on an invoked tweet) and returns
// the full Twitter user details for that user
func getInReplyToTwitterUser(sn string) *twitter.User {
	user, _, err := twitterClient.Users.Show(&twitter.UserShowParams{
		ScreenName: sn,
	})

	if err != nil {
		log.Error("cannot get honorary user details: %v", err)
	}

	return user
}
