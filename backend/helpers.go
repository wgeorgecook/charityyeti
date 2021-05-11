package main

import (
	"errors"
	"net/http"

	"github.com/ChimeraCoder/anaconda"
	"go.uber.org/zap"
)

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
