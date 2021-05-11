package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	oauth1 "github.com/klaidas/go-oauth1"

	"github.com/ChimeraCoder/anaconda"
	"go.uber.org/zap"
)

type twitterToken struct {
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

func extractID(w http.ResponseWriter, r *http.Request) (string, error) {
	// we expect the _id of the Mongo document to come in as a query param
	id := r.URL.Query()["id"]

	// query params are found as map[string], so a length of 0 means the id param wasn't found
	if len(id) == 0 {
		return "", errors.New("no id given on request but id query parameter is required")
	}

	log.Infow("Getting record", zap.String("id", id[0]))

	return id[0], nil
}

// getInReplyToTwitterUser is a helper function takes the immutable ID of a Twitter user (IE - the honorary on an invoked
// tweet) and returns the full Twitter user details for that user
func getInReplyToTwitterUser(twitterId int64) *anaconda.User {
	// in the event of a retweet with comment where a user is invoking Charity Yeti, there will be no screenname so we
	// should return early
	if twitterId == 0 {
		log.Error("No twitterId provided and cannot get honorary user details")
		return nil
	}

	users, err := twitterClient.GetUsersLookupByIds([]int64{twitterId}, nil)
	if err != nil {
		log.Errorf("cannot get honorary user details: %v", err)
		return nil
	}

	if len(users) == 0 {
		log.Errorf("no users found")
		return nil
	}

	return &users[0]
}

// getBearerToken
func getBearerToken() (string, error) {
	// POST /oauth2/token HTTP/1.1
	// Host: api.twitter.com
	// User-Agent: My Twitter App v1.0.23
	// Authorization: Basic eHZ6MWV2R ... o4OERSZHlPZw==
	// Content-Type: application/x-www-form-urlencoded;charset=UTF-8
	// Content-Length: 29
	// Accept-Encoding: gzip
	// grant_type=client_credentials

	log.Info("start get bearer token")
	// check the environment first
	if cfg.BearerToken != "" {
		return cfg.BearerToken, nil
	}
	// else try the request
	params := url.Values{}
	params.Add("grant_type", `client_credentials`)
	b := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", "https://api.twitter.com/oauth2/token", b)
	if err != nil {
		log.Errorf("could not create request for bearer token: %v", err)
		return "", err
	}
	req.SetBasicAuth(cfg.ConsumerKey, cfg.ConsumerSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("could not make request for bearer token: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("could not read response for bearer token: %v", err)
		return "", err
	}
	var token twitterToken
	if err := json.Unmarshal(body, &token); err != nil {
		log.Errorf("could not unmarshal request for bearer token: %v", err)
		return "", err
	}

	// once we have the token set the environment variable so next time we call this
	// we can just retreive it from memory
	cfg.BearerToken = token.AccessToken
	return token.AccessToken, nil

}

func getOauth1AuthorizationHeader(method, endpoint string, params map[string]string) string {
	auth := oauth1.OAuth1{
		ConsumerKey:    cfg.ConsumerKey,
		ConsumerSecret: cfg.ConsumerSecret,
		AccessToken:    cfg.AccessToken,
		AccessSecret:   cfg.AccessSecret,
	}

	return auth.BuildOAuth1Header(method, endpoint, params)
}

// getWebhooks
func getWebhooks() ([]anaconda.WebHookResp, error) {
	// curl --request GET
	// --url https://api.twitter.com/1.1/account_activity/webhooks.json
	// --header 'authorization: Bearer TOKEN'
	log.Info("getting webhooks")
	var webhooks []anaconda.WebHookResp
	endpoint := "https://api.twitter.com/1.1/account_activity/all/dev/webhooks.json"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		log.Errorf("could not create request to webhook endpoint: %v", err)
		return nil, err
	}

	bearer, err := getBearerToken()
	if err != nil {
		log.Errorf("could not get bearer token: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", bearer))

	log.Infof("making this request: %+v", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("could not make request to webhook endpoint: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof("webhook response from Twitter: %v", string(body))
	if err := json.Unmarshal(body, &webhooks); err != nil {
		log.Errorf("could not unmarshal twitter response to webhook struct: %v", err)
		return nil, err
	}

	return webhooks, nil
}

// createWebhook
func createWebhook() (*anaconda.WebHookResp, error) {
	// curl --request POST
	// --url 'https://api.twitter.com/1.1/account_activity/all/:ENV_NAME/webhooks.json?url=https%3A%2F%2Fyour_domain.com%2Fwebhook%2Ftwitter'
	// --header 'authorization: OAuth oauth_consumer_key="CONSUMER_KEY", oauth_nonce="GENERATED",
	// oauth_signature="GENERATED", oauth_signature_method="HMAC-SHA1", oauth_timestamp="GENERATED",
	// oauth_token="ACCESS_TOKEN", oauth_version="1.0"'

	log.Info("creating webhook")
	var webhook anaconda.WebHookResp
	rawEndpoint := "https://api.twitter.com/1.1/account_activity/all/dev/webhooks.json"
	endpoint := fmt.Sprintf("https://api.twitter.com/1.1/account_activity/all/dev/webhooks.json?url=%v", url.QueryEscape(cfg.WebhookCallbakURL))
	params := map[string]string{"url": cfg.WebhookCallbakURL}
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		log.Errorf("could not create request to create webhook endpoint: %v", err)
		return nil, err
	}

	req.Header.Add("Authorization", getOauth1AuthorizationHeader(http.MethodPost, rawEndpoint, params))

	log.Infof("making this request: %+v", req)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("could not make request to create webhook endpoint: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof("webhook response from Twitter: %v", string(body))
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Errorf("could not unmarshal twitter response to webhook struct: %v", err)
		return nil, err
	}

	return &webhook, nil
}

// subscribeToWebhook
func subscribeToWebhook(webhookId string) error {
	// curl --request POST
	// --url https://api.twitter.com/1.1/account_activity/all/:ENV_NAME/subscriptions.json
	// --header 'authorization: OAuth oauth_consumer_key="CONSUMER_KEY", oauth_nonce="GENERATED",
	// oauth_signature="GENERATED", oauth_signature_method="HMAC-SHA1", oauth_timestamp="GENERATED",
	// oauth_token="SUBSCRIBING_USER'S_ACCESS_TOKEN", oauth_version="1.0"'
	log.Info("start subscribe to webhook")
	endpoint := "https://api.twitter.com/1.1/account_activity/all/dev/subscriptions.json"
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		log.Errorf("could not create request to webhook endpoint: %v", err)
		return err
	}

	req.Header.Add("Authorization", getOauth1AuthorizationHeader(http.MethodPost, endpoint, map[string]string{}))

	log.Infof("making this request: %+v", req)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("could not make request to webhook endpoint: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof("webhook response from Twitter: %v", string(body))

	if resp.StatusCode != http.StatusNoContent {
		log.Errorf("expected 204 but received %v", resp.StatusCode)
		return errors.New("unacceptable response from Twitter")
	}
	return nil

}

// getWebhooks
func deleteWebhook(webhookId string) ([]anaconda.WebHookResp, error) {
	// curl --request DELETE
	// --url https://api.twitter.com/1.1/account_activity/webhooks/:WEBHOOK_ID.json
	// --header 'authorization: OAuth oauth_consumer_key="CONSUMER_KEY", oauth_nonce="GENERATED",
	// oauth_signature="GENERATED", oauth_signature_method="HMAC-SHA1", oauth_timestamp="GENERATED",
	// oauth_token="ACCESS_TOKEN", oauth_version="1.0"'
	log.Info("deleting webhook")
	var webhooks []anaconda.WebHookResp
	endpoint := fmt.Sprintf("https://api.twitter.com/1.1/account_activity/all/dev/webhooks/%v.json", webhookId)
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		log.Errorf("could not create request to webhook endpoint: %v", err)
		return nil, err
	}

	req.Header.Add("Authorization", getOauth1AuthorizationHeader(http.MethodDelete, endpoint, map[string]string{}))

	log.Infof("making this request: %+v", req)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Errorf("could not make request to webhook endpoint: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof("webhook response from Twitter: %v", string(body))
	if err := json.Unmarshal(body, &webhooks); err != nil {
		log.Errorf("could not unmarshal twitter response to webhook struct: %v", err)
		return nil, err
	}

	return webhooks, nil
}
