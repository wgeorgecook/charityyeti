package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	// define the new router, define paths, and handlers on the router
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/post/donate", receiveBtRequest)
	router.HandleFunc("/get", getRecord)
	router.HandleFunc("/get/record", getRecord)
	router.HandleFunc("/get/token", getBtToken)
	router.HandleFunc("/get/donated/all", getAllDonatedTweets)
	router.HandleFunc("/get/donated", getDonatedTweets)
	router.HandleFunc("/get/donors", getDonors)
	router.HandleFunc("/get/health", getHealth)
	router.HandleFunc("/middleware/health", checkMiddlewareHealth)

	// create a new http server with a default timeout for incoming requests
	timeout := 15 * time.Second
	srv = &http.Server{
		Addr:              fmt.Sprintf(":%v", cfg.Port),
		Handler:           router,
		ReadTimeout:       timeout,
		ReadHeaderTimeout: timeout,
		WriteTimeout:      timeout,
		IdleTimeout:       timeout,
	}

	// start the server
	log.Info("Charity Yeti is now running. Please press CTRL + C to stop.")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// getHealth returns a 200 OK and that's it
func getHealth(w http.ResponseWriter, r *http.Request) {
	log.Info("Checking backend health")
	w.WriteHeader(http.StatusOK)
	return
}

// getRecord takes a mongo _id in the body of the request and returns the collection with that data on the response body
func getRecord(w http.ResponseWriter, r *http.Request) {
	log.Info("Incoming request to get Mongo document")

	id, err := extractID(w, r)
	if err != nil {
		log.Error(err)
		// there's no ID on the request so we return early here.
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Error(err)
		}
		return
	}

	data, err := getDocument(id)
	if err != nil {
		log.Error(err)

		// there's no found document, so we return early here
		if _, err := w.Write([]byte(fmt.Sprintf("error getting Mongo document: %v", err))); err != nil {
			log.Error(err)
		}

		return
	}

	// transform the data from Mongo to a byte map so we can write it back on the request
	dataBytes, err := json.Marshal(data)
	if err != nil {
		if _, err := w.Write([]byte(fmt.Sprintf("error marshaling Mongo document: %v", err))); err != nil {
			log.Error(err)
		}
	}
	if _, err := w.Write(dataBytes); err != nil {
		log.Error(err)
	}

}

// returns an array of all data we have on all tweets with a donationValue to the requester
func getAllDonatedTweets(w http.ResponseWriter, r *http.Request) {
	tweets, err := aggregateAllDonatedTweets()
	if err != nil {
		log.Error(err)
	}

	log.Info(fmt.Sprintf("Donated tweets: %+v", tweets))
	// marshal the response into a map of our twitter data
	tweetBytes, err := json.Marshal(tweets)
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("an internal server error occured: %v", err)))
		return
	}

	// write the tweets out on the wire
	if _, err := w.Write(tweetBytes); err != nil {
		log.Error(err)
		return
	}

}

// getDonatedTweets finds all tweets with a donationValue and returns an array of tweet IDs and their
// respective summed donationValues to the requester
// note that the `_id` on this response is the originalTweetID from the database
func getDonatedTweets(w http.ResponseWriter, r *http.Request) {
	aggregate, err := aggregateDonatedTweets()
	if err != nil {
		log.Error(err)
	}

	log.Info(fmt.Sprintf("Found aggreagate: %+v", aggregate))

	aggregateBytes, err := json.Marshal(aggregate)
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("an internal server error occured: %v", err)))
		return
	}

	// write the tweets out on the wire
	if _, err := w.Write(aggregateBytes); err != nil {
		log.Error(err)
		return
	}
}

// getDonors finds all Twitter user screen name who has donated donated tweets and returns
// that array of user screennames to the requester
// note that the `_id` on this response is the invoker.screenname from the database
func getDonors(w http.ResponseWriter, r *http.Request) {
	aggregate, err := aggregateDonors()
	if err != nil {
		log.Error(err)
	}

	log.Info(fmt.Sprintf("Found aggreagate: %+v", aggregate))

	aggregateBytes, err := json.Marshal(aggregate)
	if err != nil {
		log.Error(err)
		w.WriteHeader(500)
		_, _ = w.Write([]byte(fmt.Sprintf("an internal server error occured: %v", err)))
		return
	}

	// write the tweets out on the wire
	if _, err := w.Write(aggregateBytes); err != nil {
		log.Error(err)
		return
	}
}
