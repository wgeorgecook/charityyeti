package main

import (
	"github.com/dghubble/go-twitter/twitter"
)

type nonexistentTweet string

func (n nonexistentTweet) Error() string {
	return string(n)
}

type errTwitterRateLimitedError string

func (e errTwitterRateLimitedError) Error() string {
	return string(e)
}

const (
	TWITTER_RATE_LIMIT_UNDEFINED   = -1
	TWITTER_RATE_LIMIT_SOFT_LIMIT  = "too close"
	TWITTER_RATE_LIMIT_HARD_LIMIT  = "yikes"
	TWITTER_RATE_LIMIT_NOT_LIMITED = -1
)

// initRateLimitService returns a pointer to a new instance of a
// RateLimitService
func initRateLimitService() *twitter.RateLimitService {
	if rateLimitService != nil {
		return rateLimitService
	}
	return new(twitter.RateLimitService)
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

	// check that we haven't been rate limited
	wait, err := checkTwitterRateLimit()
	if err != nil {
		return nil, err
	}

	if wait != TWITTER_RATE_LIMIT_NOT_LIMITED {
		// chill for a bit 
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

// checkTwitterRateLimit uses the rate limit service to see how many
// requests we have before we get rate limited. We consider being
// within 5 requests to be rate limited to prevent ourselves from
// actually hitting the rate limit. If we
func checkTwitterRateLimit() (int, error) {
	params := &twitter.RateLimitParams{
		Resources: []string{"statuses"},
	}

	limit, resp, err := rateLimitService.Status(params)
	if err != nil {
		return TWITTER_RATE_LIMIT_UNDEFINED, err
	}

	// close that body
	resp.Body.Close()

	// check and see if we have any rate limits
	statusLimit, ok := limit.Resources.Statuses["statuses"]
	if ok {
		// give us a little breathing room
		if statusLimit.Remaining == 5 {
			return statusLimit.Reset, errTwitterRateLimitedError(TWITTER_RATE_LIMIT_SOFT_LIMIT)
		}

		// oops we went too far
		if statusLimit.Remaining < 5 {
			return statusLimit.Reset, errTwitterRateLimitedError(TWITTER_RATE_LIMIT_HARD_LIMIT)
		}

		return TWITTER_RATE_LIMIT_NOT_LIMITED, nil
	}

	// weird but ok, we just try and see what happens then
	return TWITTER_RATE_LIMIT_UNDEFINED, nil

}
