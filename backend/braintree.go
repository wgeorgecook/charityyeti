package main

import (
	"encoding/json"
	"net/http"
)

type nonceError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// receiveNonce looks for a query param called Nonce, which the client already sent to BrainTree
// prior to sending to the server. The server relays this to BrainTree to verify that a transaction
// is acceptable and BrainTree will process this.
func receiveNonce(w http.ResponseWriter, r *http.Request) {
	log.Info("Received request for BrainTree token")

	// response object
	var response nonceError

	log.Info("checking for nonce...")
	// make sure the client sent a nonce as a get param
	var nonce string
	if nonceCheck, ok := r.URL.Query()["nonce"]; ok {
		// there's a nonce on the request, hurray!
		// r.url.query returns an array so take the first element
		nonce = nonceCheck[0]
	} else {
		// no nonce on the request, so we have a problem here
		response.Error = true
		response.Message = "No nonce on request"
		respBytes, err := json.Marshal(response)
		if err != nil {
			// this really shouldn't ever happen but it's good to catch it anyway
			log.Errorf("could not marshal response: %v", err)
			http.Error(w, "bad request", 400)
			return
		}
		w.WriteHeader(400)

		// technically also returns an error but ehhhh
		_, _ = w.Write(respBytes)

		// we're done here
		return
	}

	// oh good, things are kosher, at least so far
	log.Debugf("Nonce found: %v", nonce)

	// we need to send this nonce to BrainTree so they can process this transaction
	// TODO: we need to actually do the needful, kindly

	// the front end derives state from our response, so we marshal
	// a response and send a null error
	response.Error = false
	response.Message = "successful post to BrainTree"

	// marshal our response for the front end
	respBytes, err := json.Marshal(response)
	if err != nil {
		// this really shouldn't ever happen but it's good to catch it anyway
		log.Errorf("could not marshal response: %v", err)
		http.Error(w, "internal server error", 500)
		return
	}

	// and return the response to the requestor
	w.WriteHeader(200)

	// technically also returns an error but ehhhh
	_, _ = w.Write(respBytes)

	// not really necessary but I like closure
	return
}
