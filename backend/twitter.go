package main

import (
	"net/url"
	"strings"

	"github.com/dghubble/go-twitter/twitter"
)

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
			} `json:"message_data,omitempty"`
		} `json:"message_create,omitempty"`
	} `json:"direct_message_events,omitempty"`
	// Tweets
	UserHasBlocked    bool `json:"user_has_blocked,omitempty"`
	TweetCreateEvents []struct {
		ID                  int64         `json:"id,omitempty"`
		Text                string        `json:"text,omitempty"`
		InReplyToStatusID   int64         `json:"in_reply_to_status_id,omitempty"`
		InReplyToUserID     int           `json:"in_reply_to_user_id,omitempty"`
		InReplyToScreenName string        `json:"in_reply_to_screen_name,omitempty"`
		User                *twitter.User `json:"user,omitempty"`
	} `json:"tweet_create_events,omitempty"`
}


func initTwitterClient() *twitter.Client {
	return twitter.NewClient(httpClient)
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
		// checks and makes sure we're listening at the correct domain and it's valid
		if !strings.Contains(webhooks[0].URL, cfg.WebhookCallbakURL) || !webhooks[0].Valid {
			// we aren't registered to the current deployment, and since we can only have
			// one webhook we need to delete the existing one, or the webhook is not valid
			deleteWebhook(webhooks[0].ID)
		} else {
			// we have a valid existing webhook
			webhookId = webhooks[0].ID
		}
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