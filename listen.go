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
		log.Fatal("Could not authenticate to Twitter")
	}

	// Demux allows us to use the twitter library handlers
	// and not need to type coerce or switch on incoming
	// messages. In this case, we're just printing out the
	// tweet object as it comes
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		err := generateResponse(tweet)
		if err != nil {
			log.Error(err)
		}
	}

	log.Infow("Starting Stream")

	// These params configure what we are filtering our string for.
	// In this case, it's the user we're monitoring
	filterParams := &twitter.StreamFilterParams{
		// TODO: Make sure this doesn't hit recursion hell
		Track:         []string{"@charityyeti"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err = client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatalf("Can't connect to stream: %v", err)
	}

	// Run the stream handler in its own goroutine
	go demux.HandleChan(stream.Messages)

}
