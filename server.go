package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	log.Info("New server started")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", parseResponse)
	log.Fatal(http.ListenAndServe(":8080", router))
}

// parseResponse captures the get params off the incoming request
// We use this to get data following a successful donation
// Example:
// localhost:8080?screenname=@wgeorgecook&honorary=@charityyeti&replyToURL=https://twitter.com/WGeorgeCook/status/1197178917825630210&donationValue=5
func parseResponse(w http.ResponseWriter, r *http.Request) {

	tweet := tweetData{
		r.URL.Query()["screenname"][0],
		r.URL.Query()["honorary"][0],
		r.URL.Query()["replyToURL"][0],
		r.URL.Query()["donationValue"][0],
	}

	log.Infow("Endpoint hit")

	if len(tweet.screenname) == 0 || len(tweet.honorary) == 0 || len(tweet.replyToURL) == 0 || len(tweet.donationValue) == 0 {
		fmt.Fprintf(w, "All requests must include 'screenname', 'honorary', and 'replyToURL', and 'donationValue' params")
	} else {
		fmt.Fprintf(w, fmt.Sprintf("{Data: { screenname: %v, honorary: %v, replyToURL: %v, donationValue: %v}}", tweet.screenname, tweet.honorary, tweet.replyToURL, tweet.donationValue))
		err := respondToDonation(tweet)

		if err != nil {
			log.Error(err)
		}
	}
}
