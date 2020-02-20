package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	log.Info("New server started")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/donate", parseResponse)
	router.HandleFunc("/get", getRecord)
	router.HandleFunc("/update", updateRecord)
	log.Fatal(http.ListenAndServe(cfg.Port, router))
}

// parseResponse captures the get params off the incoming request
// We use this to get data following a successful donation
// Example:
// http://localhost:3000/?honorary=@3leero&invoker=@WGeorgeCook&invokerTweetID=1199909334777352193&originalTweetID=815781148689395712
// TODO: We want to just be passing around uuids. Just look for the id=ObjectID() on the URL and do a db lookup to
// TODO: get this information from that db record
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


// updateRecord takes an update to a Mongo document in the body of the request and returns the pre-updated
// document in the body of the response
func updateRecord(w http.ResponseWriter, r *http.Request) {
	log.Info("Incoming request to update Mongo document")

	// Read out the request body into a byte stream we can digest
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		// we can't read the incoming body
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Error(err)
		}
		return
	}

	// Unmarshal the request body bytes into our Mongo document struct
	var update charityYetiData
	if err := json.Unmarshal(body, &update); err != nil {
		// we can't unmarshal the body into our struct
		log.Error(err)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Error(err)
		}
		return
	}

	// Pass the update into our actual update function
	updated, err := updateDocument(update)
	if err != nil {
		// we we're able to update the Mongo document for whatever reason
		log.Error(err)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Error(err)
		}
		return
	}

	// transform the data from Mongo to a byte map so we can write it back on the request
	dataBytes, err := json.Marshal(updated)
	if err != nil {
		if _, err := w.Write([]byte(fmt.Sprintf("error marshaling Mongo document: %v", err))); err != nil {
			log.Error(err)
		}
	}
	if _, err := w.Write(dataBytes); err != nil {
		log.Error(err)
	}

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
