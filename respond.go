package main

import (
	"fmt"
	"log"

	"github.com/dghubble/go-twitter/twitter"
)

// respondToTweet receives an incoming tweet from a stream,
// and will respond to it with a link to donate to either
// Foundation To Decrease World Suck or directly at Parners
// In Health
func respondToTweet(incomingTweet *twitter.Tweet) {
	log.Printf(fmt.Sprintf("Incoming tweet: %+v", incomingTweet))
}
