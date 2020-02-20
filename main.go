package main

import (
	"flag"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
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
// TODO: Does this break if someone invokes on themselves?

// type to hold environment variables
type config struct {
	ConsumerKey string `env:"CONSUMER_KEY"`
	ConsumerSecret string `env:"CONSUMER_SECRET"`
	AccessToken string `env:"ACCESS_TOKEN"`
	AccessSecret string `env:"ACCESS_SECRET"`
	Port string `env:"PORT" envDefault:":8080"`
	ConnectionURI string `env:"MONGO_URI"`
}

// type to gather tweet data from an invocation of @CharityYeti
type yetiInvokedData struct {
	invoker        *twitter.User
	honorary       *twitter.User
	invokerTweetID int64
	originalTweetID int64
}

// type for building url params when we send a tweet
type successfulDonationData struct {
	invoker         string
	honorary        string
	donationValue   int
	invokerTweetID  int64
	originalTweetID int64
}

// data we keep in Mongo
type charityYetiData struct {
	ID string `json:"_id" bson:"_id"`
	OriginalTweetID int64 `json:"originalTweetID" bson:"originalTweetID"`
	InvokerTweetID int64 `json:"invokerTweetID" bson:"invokerTweetID"`
	Invoker twitter.User `json:"invoker" bson:"invoker"`
	Honorary twitter.User `json:"honorary" bson:"honorary"`
	DonationValue int `json:"donationValue" bson:"donationValue"`
}

var twitterClient *twitter.Client
var stream *twitter.Stream
var sendResponses bool
var retweetGoods bool
var log *zap.SugaredLogger
var cfg config
var mongoClient *mongo.Client


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
		log.Errorf("Error loading .env file: %v", err)
	}

	// Set environmental variables
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		log.Errorf("%+v\n", err)
	}
	log.Infof("Environment variables set: %+v", cfg)

	// configure Mongo
	log.Infow("Connecting to Mongo")
	mongoClient = initMongo(cfg.ConnectionURI)

}

func main() {
	// Configure global Twitter twitterClient
	log.Infow("Configuring Twitter twitterClient")
	config := oauth1.NewConfig(cfg.ConsumerKey, cfg.ConsumerSecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret )
	httpClient := config.Client(oauth1.NoContext, token)
	twitterClient = twitter.NewClient(httpClient)

	// Opens the Twitter feed for listening and sending initial tweet response
	// Must set writeable=true for write access
	go listen(twitterClient)

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
