package main

import (
	"context"
	"flag"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"go.uber.org/zap"
)

// TODO: Graceful panic handling

// type to hold environment variables
type config struct {
	ConsumerKey    string `env:"CONSUMER_KEY"`
	ConsumerSecret string `env:"CONSUMER_SECRET"`
	AccessToken    string `env:"ACCESS_TOKEN"`
	AccessSecret   string `env:"ACCESS_SECRET"`
	Port           string `env:"PORT" envDefault:":8080"`
	ConnectionURI  string `env:"MONGO_URI"`
	Database       string `env:"DATABASE"`
	Collection     string `env:"COLLECTION"`
}

// type to gather tweet data from an invocation of @CharityYeti
type yetiInvokedData struct {
	invoker         *twitter.User
	honorary        *twitter.User
	invokerTweetID  int64
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
	ID              string       `json:"_id" bson:"_id"`
	OriginalTweetID int64        `json:"originalTweetID" bson:"originalTweetID"`
	InvokerTweetID  int64        `json:"invokerTweetID" bson:"invokerTweetID"`
	Invoker         twitter.User `json:"invoker" bson:"invoker"`
	Honorary        twitter.User `json:"honorary" bson:"honorary"`
	DonationValue   int          `json:"donationValue" bson:"donationValue"`
}

// aggregated Mongo data
type charityYetiAggregation struct {
	Map []charityYetiData `bson:"map"`
}

var srv *http.Server
var twitterClient *twitter.Client
var stream *twitter.Stream
var tweetQueue chan *twitter.Tweet
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

	// configure Mongo
	log.Infow("Connecting to Mongo")
	mongoClient = initMongo(cfg.ConnectionURI)

}

func main() {
	// Configure global Twitter twitterClient
	log.Info("Configuring Twitter twitterClient")
	config := oauth1.NewConfig(cfg.ConsumerKey, cfg.ConsumerSecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	twitterClient = twitter.NewClient(httpClient)

	// tweetQueue is a channel that holds tweets we've heard while listening to the stream
	tweetQueue = make(chan *twitter.Tweet)

	// Opens the Twitter feed for listening and sending initial tweet response
	// Must set writeable=true for write access
	go listen(twitterClient)

	// starts a worker who processes tweets once Charity Yeti is invoked
	go processInvocation()

	// Starts the server that responds after donation
	go startServer()

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Warn(<-ch)

	// Stop the stream
	log.Info("Stopping stream")
	stream.Stop()
	log.Info("Stream stopped")

	// Stop the HTTP server
	log.Info("Stopping server")
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Errorf("could not gracefully shutdown server: %v", err)
	}
	log.Info("Server stopped")
}
