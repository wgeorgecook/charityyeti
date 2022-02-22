package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// TODO: Graceful panic handling

// type to hold environment variables
type config struct {
	ConsumerKey       string `env:"CONSUMER_KEY"`
	ConsumerSecret    string `env:"CONSUMER_SECRET"`
	AccessToken       string `env:"ACCESS_TOKEN"`
	AccessSecret      string `env:"ACCESS_SECRET"`
	Port              string `env:"PORT" envDefault:"8080"`
	ConnectionURI     string `env:"MONGO_URI"`
	Database          string `env:"MONGO_DATABASE"`
	Collection        string `env:"MONGO_COLLECTION"`
	BlockList         string `env:"BLOCK_LIST"`
	SendTweets        bool   `env:"SEND_TWEETS" envDefault:"false"`
	BearerToken       string `env:"BEARER_TOKEN"`
	WebhookCallbakURL string `env:"WEBHOOK_CALLBACK_URL"`
	EnvironmentName   string `env:"ENVIRONMENT_NAME"`
	CharityYetiId     string `env:"CHARITY_YETI_ID"`
	PublicURL         string `env:"PUBLIC_URL"`
	InvocationPhrase  string `env:"INVOCATION_PHRASE"`
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
	ID                     string        `json:"_id" bson:"_id"`
	OriginalTweetID        int64         `json:"originalTweetID,omitempty" bson:"originalTweetID,omitempty"`
	InvokerTweetID         int64         `json:"invokerTweetID,omitempty" bson:"invokerTweetID,omitempty"`
	Invoker                *twitter.User `json:"invoker,omitempty" bson:"invoker,omitempty"`
	Honorary               *twitter.User `json:"honorary,omitempty" bson:"honorary,omitempty"`
	DonationValue          float32       `json:"donationValue,omitempty" bson:"donationValue,omitempty"`
	DonationID             string        `json:"donationID,omitempty" bson:"donationID,omitempty"`
	InvokerResponseTweetID int64         `json:"invokerResponseTweetID,omitempty" bson:"invokerResponseTweetID,omitempty"`
}

// aggregated Mongo data
type charityYetiAggregation struct {
	Map []charityYetiData `bson:"map"`
}

type User struct {
	Email      *string `json:"email,omitempty" bson:"email,omitempty"`
	ID         int64   `json:"id,omitempty" bson:"id,omitempty"`
	Name       string  `json:"name,omitempty" bson:"name,omitempty"`
	ScreenName string  `json:"screen_name,omitempty" bson:"screen_name,omitempty"`
}

var srv *http.Server
var httpClient *http.Client
var twitterClient *twitter.Client
var tweetQueue chan *IncomingWebhook
var dmQueue chan *IncomingWebhook
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

	// HTTP client
	oauthConfig := oauth1.NewConfig(cfg.ConsumerKey, cfg.ConsumerSecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret)
	httpClient = oauthConfig.Client(oauth1.NoContext, token)

}

func main() {
	// Configure global Twitter twitterClient
	log.Info("Configuring Twitter twitterClient")
	twitterClient = twitter.NewClient(httpClient)

	// check if we're going to send tweets
	if cfg.SendTweets {
		log.Infow("WRITE MODE IS ENABLED")
	} else {
		log.Infow("No write access. This is a dry run.")
	}

	// tweetQueue is a channel that holds tweets that come in on webhooks
	tweetQueue = make(chan *IncomingWebhook)

	// dmQueue is a channel that holds all the DMs we get on webhooks
	dmQueue = make(chan *IncomingWebhook)

	// starts a worker who processes tweets once Charity Yeti is invoked
	go processInvocation()

	// Starts the server that responds after donation
	go startServer()

	// make sure we're listening to webhooks and are subscribed to one
	if err := initWebhooks(); err != nil {
		log.Infof("could not start charity yeti webhooks: %v", err)
		log.Info("we can't listen to webhooks, but we'll still start this up")
	}

	// listen for those sweet DMs
	go processDM()

	log.Info("Charity Yeti is now running. Please press CTRL + C to stop.")

	// Wait for SIGINT and SIGTERM (HIT CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Info(<-ch)

	// set up the context so we can cancel any straggler connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Stop the HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("could not gracefully shutdown server: %v", err)
	}
	defer log.Info("Server stopped")
}
