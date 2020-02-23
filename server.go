package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	log.Info("New server started")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/donate", goodDonation)
	router.HandleFunc("/get", getRecord)
	router.HandleFunc("/update", updateRecord)
	router.HandleFunc("/mostdonatedto", getMostDonatedTo)
	router.HandleFunc("/highestdonor", getHighestDonor)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", cfg.Port), router))
}

// goodDonation captures the body off an incoming request and sets up the struct necessary to respond to a successful
// donation event.
func goodDonation(w http.ResponseWriter, r *http.Request) {

	log.Info("Good donation received - responding to it")

	// Read the incoming request body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		log.Error(err)
		if _, werr := w.Write([]byte(fmt.Sprintf("could not read request body: %v", err))); werr != nil {
			log.Error(werr)
		}
		return
	}

	// Unmarshal the request into our charityYetiData struct
	var c charityYetiData
	if err := json.Unmarshal(body, &c); err != nil {
		log.Error(err)
		if _, werr := w.Write([]byte(fmt.Sprintf("could not marshal request body: %v", err))); werr != nil {
			log.Error(werr)
		}
		return
	}

	// set the values for a successfulDonationData struct
	tweet := successfulDonationData{
		invoker:         c.Invoker.ScreenName,
		honorary:        c.Honorary.ScreenName,
		donationValue:   c.DonationValue,
		invokerTweetID:  c.InvokerTweetID,
		originalTweetID: c.OriginalTweetID,
	}

	log.Info(fmt.Sprintf(
		"{Data: { invoker: %v, honorary: %v, invokerTweetID: %v, originalTweetID: %v, donationValue: %v}}",
		tweet.invoker, tweet.honorary, tweet.invokerTweetID, tweet.originalTweetID, tweet.donationValue))

	err = respondToDonation(tweet)

	if err != nil {
		log.Error(err)
		if _, err := w.Write([]byte(fmt.Sprintf("could not respond to donation: %v", err))); err != nil {
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

// getMostDonatedTo finds the tweet with the highest collective donationValue and returns it to the requestor
func getMostDonatedTo(w http.ResponseWriter, r *http.Request) {
	aggregate, err := aggregateHighestDonatedTweet()
	if err != nil {
		log.Error(err)
	}

	log.Info(fmt.Sprintf("Found aggreagate: %+v", aggregate))
}

// getHighestDonor finds the Twitter user screen name who has donated the most across all donated tweets and returns
// that user to the requester 
func getHighestDonor(w http.ResponseWriter, r *http.Request) {
	aggregate, err := aggregateHighestDonor()
	if err != nil {
		log.Error(err)
	}

	log.Info(fmt.Sprintf("Found aggreagate: %+v", aggregate))
}