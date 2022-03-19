package main

import "github.com/dghubble/go-twitter/twitter"

type nonexistentTweet string

func (n nonexistentTweet) Error() string {
	return string(n)
}

// sendTweet updates Charity Yeti's status, creating a new tweet in reply
// to the replyTo id given with the tweetText given
func sendTweet(tweetText string, replyTo int64) (*twitter.Tweet, error) {
	// check if the tweet we're responding to still exists
	exists, err := checkTweetExists(replyTo)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nonexistentTweet("tweet does not exist")
	}

	params := twitter.StatusUpdateParams{
		InReplyToStatusID: replyTo,
	}

	// send this tweet
	response, _, err := twitterClient.Statuses.Update(tweetText, &params)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// checkTweetExists verifies that the provided tweetId exists
func checkTweetExists(tweetId int64) (bool, error) {
	tweet, _, err := twitterClient.Statuses.Show(tweetId, &twitter.StatusShowParams{})

	if err != nil {
		return false, err
	}

	if tweet == nil {
		return false, nil
	}

	return true, nil
}
