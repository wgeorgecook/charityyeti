package main

import (
	"github.com/dghubble/go-twitter/twitter"
)

func listen(client *twitter.Client) {

	// Verify we've connected to Twitter
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	log.Infow("Connecting to Twitter")
	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		log.Fatalf("Could not authenticate to Twitter: %v", err)
	}

	// Demux allows us to use the twitter library handlers
	// and not need to type coerce or switch on incoming
	// messages.
	demux := twitter.NewSwitchDemux()

	// In this case, we're printing out the
	// tweet object as it comes and dropping it on a channel
	// for processing
	demux.Tweet = func(tweet *twitter.Tweet) {
		// send this tweet to a queue for processing
		log.Infof("Sending incoming tweet (%v) to channel", tweet.IDStr)
		tweetQueue <- tweet
	}

	// These params configure what we are filtering our string for.
	// In this case, it's the user we're monitoring
	filterParams := &twitter.StreamFilterParams{
		// []string{"Hey @charityyeti"}, // Makes sure we don't accidentally reply to folks chatting with us TODO: make this prod
		// TODO: check for other invocations (like just @mentioning) and dump those into some sort of human readable
		// table or something so we can follow up with people trying to interact with us but are doing it wrong.
		Track:         []string{"Hey @charityetidev"},
		StallWarnings: twitter.Bool(true),
	}

	log.Infow("Starting tweet stream")
	tweetStream, err = client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatalf("Can't connect to tweet stream: %v", err)
	}

}
