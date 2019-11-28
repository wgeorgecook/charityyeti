package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	log.Info("New server started")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", parseResponse)
	log.Fatal(http.ListenAndServe(cfg.Port, router))
}

// parseResponse captures the get params off the incoming request
// We use this to get data following a successful donation
// Example:
// http://localhost:3000/?honorary=@3leero&invoker=@WGeorgeCook&invokerTweetID=1199909334777352193&originalTweetID=815781148689395712
func parseResponse(w http.ResponseWriter, r *http.Request) {

	// TODO: This allows cross origin responses and is only good for deving
	w.Header().Set("Access-Control-Allow-Origin", "*")

	originalTweetID, _ := strconv.ParseInt(r.URL.Query()["originalTweetID"][0], 10, 64)
	invokerTweetID, _ := strconv.ParseInt(r.URL.Query()["invokerTweetID"][0], 10, 64)

	tweet := successfulDonationData{
		r.URL.Query()["invoker"][0],
		r.URL.Query()["honorary"][0],
		invokerTweetID,
		r.URL.Query()["donationValue"][0],
		originalTweetID,
	}

	log.Infow("Endpoint hit")

	if tweet.invoker == "" || tweet.honorary == "" || tweet.donationValue == "" || tweet.invokerTweetID == 0 || tweet.originalTweetID == 0 {
		fmt.Fprintf(w, "All requests must include 'invoker', 'honorary', and 'invokerTweetID', 'originalTweetID', and 'donationValue' params")
	} else {
		fmt.Fprintf(w, fmt.Sprintf("{Data: { invoker: %v, honorary: %v, invokerTweetID: %v, originalTweetID: %v, donationValue: %v}}", tweet.invoker, tweet.honorary, tweet.invokerTweetID, tweet.originalTweetID, tweet.donationValue))
		err := respondToDonation(tweet)

		if err != nil {
			log.Error(err)
		}
	}
}
