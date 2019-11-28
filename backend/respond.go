package main

import (
	"errors"
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
)

// respondToInvocation receives an incoming tweet from a stream,
// and will respond to it with a link to donate to either
// Foundation To Decrease World Suck or directly at Parners
// In Health
func respondToInvocation(yeti yetiInvokedData) error {
	if yeti.honorary != "" {
		donateLink := fmt.Sprintf("https://charityyeti.com?honorary=@%v&invoker=@%v&invokerTweetID=%v&originalTweetID=%v", yeti.honorary, yeti.invoker.ScreenName, yeti.invokerTweetID, yeti.originalTweetID)
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", yeti.invoker.ScreenName, yeti.honorary, donateLink)

		params := twitter.StatusUpdateParams{InReplyToStatusID: yeti.invokerTweetID}

		if sendResponses {
			log.Infow("Actually sending this!")
			_, _, err := client.Statuses.Update(tweetText, &params)
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
		_, _, err := client.Statuses.Update(tweetText, params)

		if err != nil {
			return err
		}

		// TODO: this needs testing
		if retweetGoods {
			log.Infow("We're retweeting the invoked tweet. We might break twitter TOS for this.")
			rtParams := &twitter.StatusRetweetParams{ID: tweet.originalTweetID}
			_, _, err := client.Statuses.Retweet(tweet.originalTweetID, rtParams)
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

	yeti := yetiInvokedData{
		invoker:        incomingTweet.User,
		honorary:       incomingTweet.InReplyToScreenName,
		invokerTweetID: incomingTweet.ID,
		originalTweetID: incomingTweet.InReplyToStatusID,
	}

	err := respondToInvocation(yeti)
	if err != nil {
		return err
	}

	return nil
}
