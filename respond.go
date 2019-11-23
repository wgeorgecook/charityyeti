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
func respondToInvocation(username string, honorary string, tweetID int64) error {
	if honorary != "" {
		donateLink := "https://charityyeti.com"
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", username, honorary, donateLink)

		if sendResponses {
			log.Warnw("Actually sending this!")
			_, _, err := client.Statuses.Update(tweetText, nil)
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
	tweetText := fmt.Sprintf("Good news, %v! %v thought your tweet was so great, they donated $%v to Partner's in Health on your behalf!", tweet.honorary, tweet.invoker, tweet.donationValue)
	log.Debugf(fmt.Sprintf("Tweet to send: %+v", tweetText))
	log.Debugf(fmt.Sprintf("Responding to: %v", tweet.inReplyToTweetID))

	params := &twitter.StatusUpdateParams{
		InReplyToStatusID:  tweet.inReplyToTweetID,
	}

	if sendResponses {
		log.Warnw("Actually sending this!")
		_, _, err := client.Statuses.Update(tweetText, params)

		if err != nil {
			return err
		}
	}
	return nil
}

// generateResponse parses the tweet from the demux stream, pulls out
// the user who sent it, the ID of the originating tweet, and
// passes it to respondToInvocation to send to Twitter
func generateResponse(incomingTweet *twitter.Tweet) error {
	user := incomingTweet.User
	honorary := incomingTweet.InReplyToScreenName
	tweetID := incomingTweet.ID

	err := respondToInvocation(user.ScreenName, honorary, tweetID)
	if err != nil {
		return err
	}

	return nil
}
