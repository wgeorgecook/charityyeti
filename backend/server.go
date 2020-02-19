package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
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
	router.HandleFunc("/get", getRecord)
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

// takes a mongo _id and returns the collection with that data
func getRecord(w http.ResponseWriter, r *http.Request) {
	log.Info("Incoming request to get Mongo document")

	// we expect the _id of the Mongo document to come in as a query param
	id := r.URL.Query()["id"]

	// query params are found as map[string], so a length of 0 means the id param wasn't found
	if len(id) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte("no id given on request but id query parameter is required")); err != nil {
			fmt.Printf(err.Error())
		}
	}

	log.Infow("Getting record", zap.String("id", id[0]))

	collection := mongoClient.Database("charityyeti-test").Collection("twitterData")

	// create an OID bson primitive based on the ID that comes in on the request
	oid, err := primitive.ObjectIDFromHex(id[0])
	if err != nil {
		log.Error(err)
	}

	// find and unmarshal the document to a struct we can return
	var data charityYetiData
	filter := bson.M{"_id": oid}
	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		if _, err := w.Write([]byte(fmt.Sprintf("could not decode Mongo data: %v", err))); err != nil {
			log.Error(err)
		}
	}

	// transform the data from Mongo to a byte map so we can write it back on the request
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
	}
	if _, err := w.Write(dataBytes); err != nil {
		log.Error(err)
	}

}
