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

	"github.com/joho/godotenv"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// TODO: Graceful panic handling

// type to hold environment variables
type config struct {
	ConsumerKey           string `env:"CONSUMER_KEY"`
	ConsumerSecret        string `env:"CONSUMER_SECRET"`
	AccessToken           string `env:"ACCESS_TOKEN"`
	AccessSecret          string `env:"ACCESS_SECRET"`
	Port                  string `env:"PORT" envDefault:"8080"`
	PostgresConnectionURI string `env:"POSTGRES_CONNECTION_URI"`
	SendTweets            bool   `env:"SEND_TWEETS"`
	BearerToken           string `env:"BEARER_TOKEN"`
	WebhookCallbakURL     string `env:"WEBHOOK_CALLBACK_URL"`
	PublicURL             string `env:"PUBLIC_URL" envDefault:"https://charityyeti.casadecook.com"`
	EnvironmentName       string `env:"ENVIRONMENT_NAME"`
	CharityYetiId         string `env:"CHARITY_YETI_ID"`
	LOCAL                 bool   `env:"LOCAL" envDefault:"false"`
}

var (
	srv           *http.Server
	httpClient    *http.Client
	twitterClient *twitter.Client
	tweetQueue    chan *IncomingWebhook
	dmQueue       chan *IncomingWebhook
	retweetGoods  bool
	log           *zap.SugaredLogger
	cfg           config
)

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

	// Configure global Twitter twitterClient
	log.Info("Configuring Twitter twitterClient")
	twitterClient = initTwitterClient()

	// connect postgres
	initPostgres()

	// HTTP client
	httpClient = initHttpClient()

}

func main() {

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
