package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// TODO: Handle currency properly on the http server and respondToDonation
// TODO: Graceful shutdown on http server
// TODO: Graceful panic handling
// TODO: Queuing for tweets

// type for building url params when we send a tweet
type successfulDonationData struct {
	invoker          string
	honorary         string
	inReplyToTweetID int64
	donationValue    string
}

var client *twitter.Client
var stream *twitter.Stream
var sendResponses bool
var log *zap.SugaredLogger

func init() {
	// Configure logging
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()


	// Parse command line flags
	flag.BoolVar(&sendResponses, "sendResponses", false, "set to true to respond to tweets")
	flag.Parse()
	if sendResponses {
		log.Infow("WRITE MODE IS ENABLED")
	} else {
		log.Infow("No write access. This is a dry run.")
	}
}

func main() {

	// Load environment variables from .env file
	log.Infow("Loading env variables")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Configure global Twitter client
	log.Infow("Configuring Twitter client")
	config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
	token := oauth1.NewToken(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_SECRET"))
	httpClient := config.Client(oauth1.NoContext, token)
	client = twitter.NewClient(httpClient)

	// Opens the Twitter feed for listening and sending initial tweet response
	// Must set writeable=true for write access
	go listen(client)

	// Starts the server that responds after donation
	go startServer()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Warn(<-ch)

	// Stop the stream
	log.Warnw("Stopping stream")
	stream.Stop()
}
