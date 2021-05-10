package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/caarlos0/env/v6"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"go.uber.org/zap"
)

// TODO: Graceful panic handling

// type to hold environment variables
type config struct {
	ConsumerKey        string `env:"CONSUMER_KEY"`
	ConsumerSecret     string `env:"CONSUMER_SECRET"`
	AccessToken        string `env:"ACCESS_TOKEN"`
	AccessSecret       string `env:"ACCESS_SECRET"`
	Port               string `env:"PORT" envDefault:":8080"`
	ConnectionURI      string `env:"MONGO_URI"`
	Database           string `env:"DATABASE"`
	Collection         string `env:"COLLECTION"`
	BlockList          string `env:"BLOCK_LIST"`
	MiddlewareEndpoint string `env:"MIDDLEWARE_ENDPOINT"`
	MiddlewareToken    string `env:"MIDDLEWARE_TOKEN_ENDPOINT"`
	MiddlewareHealth   string `env:"MIDDLEWARE_HEALTH"`
	SendTweets         bool   `env:"SEND_TWEETS"`
	BearerToken        string `env:"BEARER_TOKEN"`
	WebhookID          string `env:"WEBHOOK_ID"`
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
	invoker                string
	honorary               string
	donationValue          float32
	invokerTweetID         int64
	originalTweetID        int64
	invokerResponseTweetID int64
}

// data we keep in Mongo
type charityYetiData struct {
	ID                     string       `json:"_id" bson:"_id"`
	OriginalTweetID        int64        `json:"originalTweetID" bson:"originalTweetID"`
	InvokerTweetID         int64        `json:"invokerTweetID" bson:"invokerTweetID"`
	Invoker                twitter.User `json:"invoker" bson:"invoker"`
	Honorary               twitter.User `json:"honorary" bson:"honorary"`
	DonationValue          float32      `json:"donationValue" bson:"donationValue"`
	DonationID             string       `json:"donationID" bson:"donationID"`
	InvokerResponseTweetID int64        `json:"invokerResponseTweetID" bson:"invokerResponseTweetID"`
}

// aggregated Mongo data
type charityYetiAggregation struct {
	Map []charityYetiData `bson:"map"`
}

var srv *http.Server
var httpClient *http.Client
var twitterClient *twitter.Client
var tweetStream *twitter.Stream
var dmStream *twitter.Stream
var tweetQueue chan *twitter.Tweet
var dmQueue chan *twitter.DirectMessage
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
	flag.BoolVar(&retweetGoods, "retweetGoods", false, "set to true to retweet the tweets that get the Yeti invoked on them")
	flag.Parse()

	// Load environment variables from .env file
	log.Infow("Loading env variables")
	err := godotenv.Load()
	if err != nil {
		log.Infof("Error loading .env file: %v", err)
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
	httpClient = config.Client(oauth1.NoContext, token)
	twitterClient = twitter.NewClient(httpClient)

	// check if we're going to send tweets
	if cfg.SendTweets {
		log.Infow("WRITE MODE IS ENABLED")
	} else {
		log.Infow("No write access. This is a dry run.")
	}

	// tweetQueue is a channel that holds tweets we've heard while listening to the stream
	tweetQueue = make(chan *twitter.Tweet)

	// dmQueue is a channel that holds all the DMs we get while listening to incoming DMs
	dmQueue = make(chan *twitter.DirectMessage)

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
	log.Info(<-ch)

	// set up the context so we can cancel any straggler connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// Stop the streams
		log.Info("Stopping tweet stream")
		tweetStream.Stop()
		log.Info("Tweet tream stopped")
		// cancel the context
		cancel()
	}()
	// Stop the HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("could not gracefully shutdown server: %v", err)
	}
	defer log.Info("Server stopped")
}
