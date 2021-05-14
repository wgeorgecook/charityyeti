package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/gorilla/mux"
)

type CRCResponse struct {
	ResponseToken string `json:"response_token"`
}

type IncomingWebhook struct {
	ForUserID string `json:"for_user_id,omitempty"`
	// DMs
	DirectMessageEvents []struct {
		Type             string `json:"type,omitempty"`
		ID               string `json:"id,omitempty"`
		CreatedTimestamp string `json:"created_timestamp,omitempty"`
		MessageCreate    struct {
			Target struct {
				RecipientID string `json:"recipient_id,omitempty"`
			} `json:"target,omitempty"`
			SenderID    string `json:"sender_id,omitempty"`
			MessageData struct {
				Text     string `json:"text,omitempty"`
				Entities struct {
					Hashtags     []interface{} `json:"hashtags,omitempty"`
					Symbols      []interface{} `json:"symbols,omitempty"`
					UserMentions []interface{} `json:"user_mentions,omitempty"`
					Urls         []interface{} `json:"urls,omitempty"`
				} `json:"entities,omitempty"`
			} `json:"message_data,omitempty"`
		} `json:"message_create,omitempty"`
	} `json:"direct_message_events,omitempty"`
	// Tweets
	UserHasBlocked    bool `json:"user_has_blocked,omitempty"`
	TweetCreateEvents []struct {
		CreatedAt           string        `json:"created_at,omitempty"`
		ID                  int64         `json:"id,omitempty"`
		Text                string        `json:"text,omitempty"`
		InReplyToStatusID   int64         `json:"in_reply_to_status_id,omitempty"`
		InReplyToUserID     int           `json:"in_reply_to_user_id,omitempty"`
		InReplyToScreenName string        `json:"in_reply_to_screen_name,omitempty"`
		User                *twitter.User `json:"user,omitempty"`
	} `json:"tweet_create_events,omitempty"`
}

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	// define the new router, define paths, and handlers on the router
	router := mux.NewRouter()
	router.HandleFunc("/post/donate", successfulDonation)
	router.HandleFunc("/get", getRecord)
	router.HandleFunc("/get/record", getRecord)
	router.HandleFunc("/get/donated/all", getAllDonatedTweets)
	router.HandleFunc("/get/donated", getDonatedTweets)
	router.HandleFunc("/get/donors", getDonors)
	router.HandleFunc("/get/health", getHealth)
	router.HandleFunc("/oauth2/callback", oauthCallback)
	router.HandleFunc("/webhook/listen", webhookListener)

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

func oauthRequest(w http.ResponseWriter, r *http.Request) {
	// redirect the user to Twitter and awaie callback
	path := "https://api.twitter.com/oauth/request_token"
	//Set parameters
	values := url.Values{}
	values.Set("url", "https://charityyeti.casadecook.com/webhook/listen")

	//Make Oauth Post with parameters
	resp, err := httpClient.PostForm(path, values)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not make Post for oauth request: %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	//Parse response and check response
	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof("response: %v", string(body))
	w.Write(body)
}

func oauthCallback(w http.ResponseWriter, r *http.Request) {
	log.Info("incoming oauth2 callback")
}

// webhookListener receives POST requests from Twitter with the payloads we subscribe to
// they will sometimes send a Challenge-Response Check via GET request, so we first
// check for that before processing the request
func webhookListener(w http.ResponseWriter, r *http.Request) {
	log.Info("Received webhook payload")
	defer r.Body.Close()

	// check for the CRC
	if token, ok := r.URL.Query()["crc_token"]; ok {
		// this is a challenge, so we need to take this token and
		// make an HMAC SHA-256 hash using it and our client secret
		hash := hmac.New(sha256.New, []byte(cfg.ConsumerSecret))

		log.Infof("incoming crc_token: %v", token[0])
		w.Header().Set("Content-Type", "application/json")

		// write the incoming crc_token using the hash
		hash.Write([]byte(token[0]))

		// save the sha as a string we can return to Twitter
		sha := base64.StdEncoding.EncodeToString(hash.Sum(nil))

		// marshal our response token
		response := CRCResponse{ResponseToken: fmt.Sprintf("sha256=%v", sha)}
		respBytes, err := json.Marshal(response)
		if err != nil {
			// ope
			log.Errorf("could not marshal response token: %v", err)
			http.Error(w, "I'm not even sure how this happens, but there was an error", 500)
			return
		}

		// write our response back on the request
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
	// log.Debugf("incoming webhook payload: %v", string(reqBytes))

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
	var data charityYetiData
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// if the front end sent us something we can't decode
		log.Errorf("Could not decode request from frontend: %v", err)
		http.Error(w, "could not decode request", http.StatusBadRequest)
		return
	}

	// get the mongo document associated with this mongo id
	doc, err := getDocument(data.ID)
	if err != nil {
		log.Errorf("could not get this document: %v", err)
		http.Error(w, "no document matches the id provided", http.StatusBadRequest)
		return
	}

	doc.DonationValue = data.DonationValue

	// write the donation value back to the document
	if err := goodDonation(*doc); err != nil {
		log.Errorf("Good donation call failed: %v", err)
		http.Error(w, fmt.Sprintf("An internal server error occured. We're very sorry, but here's some details: %v", err.Error()), 500)
		return
	}

	// if everything is cool then we done
	w.WriteHeader(200)

	// not really necessary but I like closure
	return

}

// initWebhooks will check and see if we've got a webhook registered with Twitter,
// and makes sure that charityyeti is subscribed to that webhook
func initWebhooks() error {
	// first get the webhooks we already have registered
	v := url.Values{}
	v.Add("env_name", "dev")
	webhooks, err := getWebhooks()
	if err != nil {
		log.Errorf("could not get registered webhooks: %v", err)
		return err
	}

	webhookId := ""
	if len(webhooks) != 0 {
		// checks and makes sure we're listening at the correct domain
		if !strings.Contains(webhooks[0].URL, cfg.WebhookCallbakURL) {
			// we aren't registered to the current deployment, and since we can only have
			// one webhook we need to delete the existing one...
			deleteWebhook(webhooks[0].ID)
		}

		// if we are, there's an existing webhook
		webhookId = webhooks[0].ID
	}

	if webhookId == "" {
		// register a new one
		log.Info("no current registered webhooks, creating a new one")
		webhook, err := createWebhook()
		if err != nil {
			log.Errorf("could not register a new webhook: %v", err)
			return err
		}
		webhookId = webhook.ID
	}
	// and then checks to see if we have charity yeti subscribed to it
	subscribed, err := getSubscriptions()
	if err != nil {
		log.Errorf("couldn't check subscriptions: %v", err)
	}
	if !subscribed {
		// subscribe CharityYeti to this webhook
		err = subscribeToWebhook(webhookId)
		if err != nil {
			log.Errorf("could not subscribe CharityYeti to webhook: %v", err)
			return err
		}
	}

	// all done!
	log.Infof("Registered webhook: %v", webhookId)
	return nil
}
