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

const MONGO_NO_RESULT = "mongo: no documents in result"

// respondToInvocation receives an incoming tweet from a stream, and will respond to it with a link to donate via the
// Charity Yeti website. The donation link includes an id for a Mongo document for the front end to retrieve and add
// on the donation value after a successful donation.
func respondToInvocation(yeti yetiInvokedData) error {

	if yeti.honorary.ScreenName == "" {
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

	// check if this donor already has a mongo document
	var existingDonor bool
	donorDoc, err := getDocumentByDonorID(yeti.invoker.ID)
	if err != nil {
		// there was an error, but if this is the first time this person donated we
		// expect Mongo to not return a document, so we check that here
		if err.Error() != MONGO_NO_RESULT {
			// this error is not acceptable and we need to return it
			return err
		} else {
			// if the Mongo check only failed because a doc didn't exist
			// then we can continue but need to create a document for
			// this user
			existingDonor = false
		}
	} else {
		// since there was no error, Mongo returned a document to us of a previous
		// existing donor
		existingDonor = true
	}

	if existingDonor {
		// we have a document for this donor already so send along the id for this document
		dataID := donorDoc.ID
		originalTweetID := yeti.originalTweetID
		invokerTweetID := yeti.invokerTweetID
		inReplyToUser := yeti.honorary.ID
		donateLink := fmt.Sprintf("https://charityyeti.com?id=%v&originalTweetId=%v%invokerTweetId=%v%inReplyToUser=%v", dataID, originalTweetID, invokerTweetID, inReplyToUser)
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", yeti.invoker.ScreenName, yeti.honorary.ScreenName, donateLink)
		params := twitter.StatusUpdateParams{InReplyToStatusID: yeti.invokerTweetID}

		if sendResponses {
			// send the tweet
			log.Infow("Actually sending this!")
			_, _, err = twitterClient.Statuses.Update(tweetText, &params)
			if err != nil {
				return err
			}
		}

	} else {
		// we need to create a new donor doc, and then add the first donation to it
		dataID := primitive.NewObjectID()
		originalTweetID := yeti.originalTweetID
		invokerTweetID := yeti.invokerTweetID
		inReplyToUser := yeti.honorary.ID
		donateLink := fmt.Sprintf("https://charityyeti.com?id=%v&originalTweetId=%v%invokerTweetId=%v%inReplyToUser=%v", dataID, originalTweetID, invokerTweetID, inReplyToUser)
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", yeti.invoker.ScreenName, yeti.honorary.ScreenName, donateLink)

		params := twitter.StatusUpdateParams{InReplyToStatusID: yeti.invokerTweetID}

		if sendResponses {
			// create the record in Mongo
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			newDonor := twitterUser{
				ID:         yeti.invoker.ID,
				Name:       yeti.invoker.Name,
				Email:      yeti.invoker.Email,
				ScreenName: yeti.invoker.ScreenName,
			}

			honorary := twitterUser{
				ID:         yeti.honorary.ID,
				Name:       yeti.honorary.Name,
				Email:      yeti.honorary.Email,
				ScreenName: yeti.honorary.ScreenName,
			}

			thisDonation := donation{
				OriginalTweetID: yeti.originalTweetID,
				InvokerTweetID:  yeti.invokerTweetID,
				Honorary:        honorary,
			}

			data := bson.M{
				"_id":   dataID,
				"donor": newDonor,
				"$push": bson.M{
					"donations": thisDonation,
				},
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
	}

	log.Info(tweetText)
	return nil

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

		honorary := getInReplyToTwitterUser(incomingTweet.InReplyToUserID)

		invoker := &twitterUser{
			ID:         incomingTweet.User.ID,
			Name:       incomingTweet.User.Name,
			Email:      incomingTweet.User.Email,
			ScreenName: incomingTweet.User.ScreenName,
		}

		yeti := yetiInvokedData{
			invoker:         invoker,
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
