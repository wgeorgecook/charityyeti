package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	log.Print("Loading env variables")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Configure global Twitter client
	log.Print("Configuring Twitter client")
	config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
	token := oauth1.NewToken(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_SECRET"))
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	// Verify we've connected to Twitter
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}
	log.Print("Connecting to Twitter")
	_, _, err = client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		log.Fatal("Could not authenticate to Twitter")
	}

	// Demux allows us to use the twitter library handlers
	// and not need to type coerce or switch on incoming
	// messages. In this case, we're just printing out the
	// tweet object as it comes
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		fmt.Println(tweet.Text)
	}

	fmt.Println("Starting Stream")

	// These params configure what we are filtering our string for.
	// In this case, it's the user we're monitoring
	// realDonaldTrump is a filler just because there is MASSIVE
	// amounts of traffic for that account
	filterParams := &twitter.StreamFilterParams{
		Track:         []string{"@realDonaldTrump"},
		StallWarnings: twitter.Bool(true),
	}
	stream, err := client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatalf("Can't connect to stream: %v", err)
	}

	// Run the stream hander in its own goroutine
	go demux.HandleChan(stream.Messages)

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)

	// Stop the stream
	fmt.Println("Stopping Stream")
	stream.Stop()
}
