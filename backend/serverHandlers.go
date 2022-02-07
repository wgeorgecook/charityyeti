package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// getHealth returns a 200 OK and that's it
func getHealth(w http.ResponseWriter, r *http.Request) {
	log.Info("Checking backend health")
	w.WriteHeader(http.StatusOK)
}

// generateCrcResponseToken makes an HMAC SHA-256 hash with the client
// secret and the provided token then generates the token
// to write back
func generateCrcResponseToken(token string) ([]byte, error) {
	// this is a challenge, so we need to take this token and
	// make an HMAC SHA-256 hash using it and our client secret
	hash := hmac.New(sha256.New, []byte(cfg.ConsumerSecret))

	log.Infof("incoming crc_token: %v", token[0])

	// write the incoming crc_token using the hash
	hash.Write([]byte(token))

	// save the sha as a string we can return to Twitter
	sha := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	// a type to marshall the crc into
	type CRCResponse struct {
		ResponseToken string `json:"response_token"`
	}

	// marshal our response token
	response := CRCResponse{ResponseToken: fmt.Sprintf("sha256=%v", sha)}
	respBytes, err := json.Marshal(response)
	if err != nil {
		// ope
		return nil, err
	}

	return respBytes, nil

}

// webhookListener receives POST requests from Twitter with the payloads we subscribe to
// they will sometimes send a Challenge-Response Check via GET request, so we first
// check for that before processing the request
func webhookListener(w http.ResponseWriter, r *http.Request) {
	log.Info("Received webhook payload")
	defer r.Body.Close()

	// check for the CRC
	if token, ok := r.URL.Query()["crc_token"]; ok {
		respBytes, err := generateCrcResponseToken(token[0])
		if err != nil {
			// ope
			log.Errorf("could not marshal response token: %v", err)
			http.Error(w, "I'm not even sure how this happens, but there was an error", 500)
			return

		}

		// write our response back on the request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)

		// we're done here
		return
	}

	// read out the request
	reqBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("could not read webhook request: %v", err)
		http.Error(w, "malformed request", 400)
		return
	}

	// print it out for debug
	log.Debugf("incoming webhook payload: %v", string(reqBytes))

	// unmarshall it
	var wh IncomingWebhook
	if err := json.Unmarshal(reqBytes, &wh); err != nil {
		// this probably is not a DM event so we will just log and ignore it
		log.Debugf("could not unmarshal webhook: %v", err)
		return
	}

	processed := false
	if len(wh.DirectMessageEvents) > 0 {
		log.Infof("sending this to the DM queue: %+v", wh)
		// drop on the queue for processing
		dmQueue <- &wh
		processed = true
	}

	if len(wh.TweetCreateEvents) > 0 {
		// drop on the Tweet queue for processing
		log.Infof("sending this to the tweet queue: %+v", wh)
		tweetQueue <- &wh
		processed = true
	}

	// neither a tweet or a DM
	if !processed {
		log.Info("nothing to process on this webhook")
	}

}

// getRecord takes a mongo _id in the body of the request and returns the collection with that data on the response body
func getRecord(w http.ResponseWriter, r *http.Request) {
	log.Info("Incoming request to get Donation")

	id, err := extractID(w, r)
	if err != nil {
		log.Error(err)
		// there's no ID on the request so we return early here.
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Error(err)
		}
		return
	}

	ctx := generateContextWithRequestId(context.Background())
	data, err := getDonation(ctx, id)
	if err != nil {
		log.Error(err)

		// there's no found document, so we return early here
		if _, err := w.Write([]byte(fmt.Sprintf("error querying for document: %v", err))); err != nil {
			log.Error(err)
		}

		return
	}

	// transform the data from Mongo to a byte map so we can write it back on the request
	dataBytes, err := json.Marshal(data)
	if err != nil {
		if _, err := w.Write([]byte(fmt.Sprintf("error marshaling donation: %v", err))); err != nil {
			log.Error(err)
		}
	}
	if _, err := w.Write(dataBytes); err != nil {
		log.Error(err)
	}

}

// successfulDonation gets called from the front end when Brain Tree returns a success
// after processing a payment. This wraps the update document and respond to invoker functions.
func successfulDonation(w http.ResponseWriter, r *http.Request) {
	log.Info("Successful donation incoming!")
	defer r.Body.Close()

	// get a byte slice from our request
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("could not read request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// print out the incoming request for debug purposes
	log.Infof("request body: %v", string(bodyBytes))

	// read the incoming data into a struct
	var data Donation
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// if the front end sent us something we can't decode
		log.Errorf("Could not decode request from frontend: %v", err)
		http.Error(w, "could not decode request", http.StatusBadRequest)
		return
	}

	// get the donation associated with this donation id
	ctx := generateContextWithRequestId(context.Background())
	donation, err := getDonation(ctx, data.ID)
	if err != nil {
		log.Errorf("could not get this document: %v", err)
		http.Error(w, "no document matches the id provided", http.StatusBadRequest)
		return
	}

	donation.DonationValue = data.DonationValue

	// write the donation value back to the document
	if err := goodDonation(ctx, *donation); err != nil {
		log.Errorf("Good donation call failed: %v", err)
		http.Error(w, fmt.Sprintf("An internal server error occured. We're very sorry, but here's some details: %v", err.Error()), 500)
		return
	}

	// if everything is cool then we done
	w.WriteHeader(200)

}
