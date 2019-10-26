package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var client *twitter.Client
var httpClient *http.Client

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Configure global Twitter client
	config := &clientcredentials.Config{
		ClientID:     os.Getenv("CONSUMER_KEY"),
		ClientSecret: os.Getenv("CONSUMER_SECRET"),
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}

	httpClient := config.Client(oauth2.NoContext)
	client = twitter.NewClient(httpClient)
}

func main() {
	fmt.Println("The start of something beautiful")
	fmt.Println(fmt.Sprintf("Twitter client: %+v", client))
}
