package main

import (
	"flag"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"go.uber.org/zap"
)

// TODO: Handle currency properly on the http server and respondToDonation
// TODO: Graceful shutdown on http server
// TODO: Graceful panic handling
// TODO: Queuing for tweets

// type to hold environment variables
type config struct {
	ConsumerKey string `env:"CONSUMER_KEY"`
	ConsumerSecret string `env:"CONSUMER_SECRET"`
	AccessToken string `env:"ACCESS_TOKEN"`
	AccessSecret string `env:"ACCESS_SECRET"`
	Port string `env:"PORT" envDefault:":8080"`
}

// type to gather tweet data from an invocation of @CharityYeti
type yetiInvokedData struct {
	invoker *twitter.User
	honorary string
	originalTweetId int64
}

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
var retweetGoods bool
var log *zap.SugaredLogger
var cfg config

func init() {
	// Configure logging
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	log = logger.Sugar()


	// Parse command line flags
	flag.BoolVar(&sendResponses, "sendResponses", false, "set to true to respond to tweets")
	flag.BoolVar(&retweetGoods, "retweetGoods", false, "set to true to retweet the tweets that get the Yeti invoked on them")
	flag.Parse()
	if sendResponses {
		log.Infow("WRITE MODE IS ENABLED")
	} else {
		log.Infow("No write access. This is a dry run.")
	}

	// Load environment variables from .env file
	log.Infow("Loading env variables")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set environmental variables
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		log.Errorf("%+v\n", err)
	}
	log.Infow("Environment variables set")
}

func main() {
	// Configure global Twitter client
	log.Infow("Configuring Twitter client")
	config := oauth1.NewConfig(cfg.ConsumerKey, cfg.ConsumerSecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret )
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
