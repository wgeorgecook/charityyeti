package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// brainTreeData is information we receive from the front end after a transaction
// gets started, plus the options array we add on. We pass this data back to
// Brain Tree opaquely.
type brainTreeData struct {
	Token      string            `json:"token,omitempty"`
	Nonce      string            `json:"paymentMethodNonce,omitempty"`
	Amount     string            `json:"amount,omitempty"`
	DeviceData string            `json:"deviceData,omitempty"`
	Options    []map[string]bool `json:"options,omitempty"`
	MongoID    string            `json:"_id,omitempty"`
}

// brainTreeTransaction is the return data the middleware sends us after a
// successful transaction
type brainTreeTransaction struct {
	Amount         string         `json:"amount,omitempty"`
	BillingDetails billingDetails `json:"billingDetails,omitempty"`
	ID             string         `json:"id,omitempty"`
	Token          string         `json:"token,omitempty"`
}

// brainTreeToken is the client token we need to initiate

// billingDetails is the data Brain Tree stored about this user and returned to us
type billingDetails struct {
	FirstName       string `json:"firstName,omitempty"`
	LastName        string `json:"lastName,omitempty"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	ExtendedAddress string `json:"extendedAddress,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Region          string `json:"region,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	Country         string `json:"countryName,omitempty"`
}

func getBtToken(w http.ResponseWriter, r *http.Request) {
	token, _, err := doBrainTreeTokenRequest()
	if err != nil {
		log.Errorf("Error hitting middleware: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	// insert the token into a struct we can return
	returnValue := brainTreeData{
		Token: token,
	}

	// marshal the struct to send over the wire
	jsonBytes, err := json.Marshal(returnValue)
	if err != nil {
		// f
		log.Errorf("Could not marshal Brain Tree token for response: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	// send it
	w.Write(jsonBytes)
	return
}

// receiveBtRequest receives a brainTreeData struct as an POST body,
// and then relays that request to the middleware that forwards directly
// to Brain Tree. We then send the middleware's response back on the request.
func receiveBtRequest(w http.ResponseWriter, r *http.Request) {
	log.Info("Received request for BrainTree transaction")

	// unmarshal the request body
	var btData brainTreeData
	if err := json.NewDecoder(r.Body).Decode(&btData); err != nil {
		// if we
		log.Errorf("could not decode incoming request: %v", err)
		http.Error(w, "bad request", 400)
		return
	}

	log.Debugf("decoded request: %+v", btData)

	log.Info("checking for nonce...")
	if btData.Nonce == "" {
		log.Error("no nonce found")
		http.Error(w, "bad request", 400)
		return
	}

	// oh good, things are kosher, at least so far
	log.Debugf("Nonce found: %v", btData.Nonce)

	// we need to send this nonce to BrainTree so they can process this transaction
	transaction, statusCode, err := doBrainTreeTransactionRequest(btData)
	if err != nil {
		// uh oh
		log.Errorf("Brain Tree request failed: %v", err)
		http.Error(w, err.Error(), statusCode)
		return
	}

	// at this point the Brain Tree stuff worked so we successfully processed
	// a donation! We need to look up the data we have based on this ID
	data, err := getDocument(btData.MongoID)
	if err != nil {
		log.Errorf("Could not get the Charity Yeti data from this ID: %v", err)
		http.Error(w, fmt.Sprintf("An internal server error occured. We're very sorry, but here's some details: %v", err.Error()), 500)
		return
	}

	// the transaction amount Brain Tree returns is a string, so we need to convert it to float32
	var donationValue float32
	donationValue64, err := strconv.ParseFloat(btData.Amount, 32)
	if err == nil {
		log.Errorf("Could not convert string to float: %v", err)
		donationValue = 0
	}
	donationValue = float32(donationValue64)

	// now we can include some transaction data to save back to Mongo
	data.DonationValue = donationValue
	data.DonationID = transaction.ID

	// we can save the transaction back to the database and then let
	// the original tweeter know someone did this cool thing!
	if err := goodDonation(*data); err != nil {
		log.Errorf("Good donation call failed: %v", err)
		http.Error(w, fmt.Sprintf("An internal server error occured. We're very sorry, but here's some details: %v", err.Error()), 500)
		return
	}

	// if everything is cool then we done
	w.WriteHeader(200)

	// not really necessary but I like closure
	return
}

// sends an empty request to api/payment/GetClientToken
// in the middleware
func doBrainTreeTokenRequest() (string, int, error) {
	log.Info("Received a client token request to forward to middleware")

	// setup the request
	req, err := http.NewRequest(http.MethodGet, cfg.MiddlewareToken, bytes.NewReader([]byte("")))
	if err != nil {
		log.Errorf("Could not create request to middleware: %v", err)
		return "", 500, err
	}

	// set the headers
	req.Header.Set("Content-Type", "application/json")

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Could not make the request to the middleware: %v", err)
		return "", 500, err
	}
	defer resp.Body.Close()

	// the middleware just returns a string over the wire, so we need to transform
	// that into a Golang string
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("could not read response from middleware: %v", err)
		return "", 500, err
	}

	token := string(respBytes)

	// everything is great!
	return token, 200, nil
}

// checkMiddlewareHealth hits the health endpoint in the middleware
// and returns the status code
func checkMiddlewareHealth(w http.ResponseWriter, r *http.Request) {
	log.Info("Checking middleware health")
	resp, err := http.Get(cfg.MiddlewareHealth)
	if err != nil {
		log.Errorf("Could not make request to middleware health: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}
	defer resp.Body.Close()

	respBytes, _ := ioutil.ReadAll(resp.Body)

	log.Infof("Response from middleware: %v", string(respBytes))
	w.WriteHeader(resp.StatusCode)
	w.Write(respBytes)
	return
}

// doBrainTreeTransactionRequest sends the request to the Brain Tree middleware
func doBrainTreeTransactionRequest(data brainTreeData) (*brainTreeTransaction, int, error) {
	log.Info("Building request to send data to BrainTree")

	// marshal the incoming data to send on the wire
	btBytes, err := json.Marshal(data)
	if err != nil {
		// not sure how this can even happen
		log.Errorf("Could not marshal incoming brain tree struct: %v", err)
		return nil, 500, err
	}

	// setup the request
	req, err := http.NewRequest(http.MethodPost, cfg.MiddlewareEndpoint, bytes.NewReader(btBytes))
	if err != nil {
		log.Errorf("Could not create request to middleware: %v", err)
		return nil, 500, err
	}

	// set the headers
	req.Header.Set("Content-Type", "application/json")

	// make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Could not make the request to the middleware: %v", err)
		return nil, 500, err
	}
	defer resp.Body.Close()

	// just for logging so I suppress the error here
	respBytes, _ := ioutil.ReadAll(resp.Body)
	log.Infof("Response from middleware: %v", string(respBytes))

	// the middleware returns a Transaction object back to us we need to unmarshal to return
	// https://developers.braintreepayments.com/reference/response/transaction/dotnet
	var t brainTreeTransaction
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		// uh oh
		log.Errorf("Could not decode response from middleware: %v", err)
		return nil, 500, err
	}

	// everything is great!
	return &t, 200, nil
}
