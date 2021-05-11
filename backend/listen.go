package main

import (
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

func listen(client *anaconda.TwitterApi) {
	log.Infow("Starting tweet stream")

	// the values we're listening for
	v := url.Values{}
	// TODO: check for other invocations (like just @mentioning) and dump those into some sort of human readable
	// TODO: table or something so we can follow up with people trying to interact with us but are doing it wrong.
	v.Add("track", "Hey @charityyetidev")
	v.Add("stall_warnings", "true")

	tweetStream = twitterClient.PublicStreamFilter(v)

	for t := range tweetStream.C {
		switch tweet := t.(type) {
		case anaconda.Tweet:
			log.Infow("Incoming tweet: %+v", tweet)
			tweetQueue <- &tweet
		default:
			log.Infow("unrelated event: %+v", tweet)
		}
	}

}
