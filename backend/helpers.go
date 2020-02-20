package main

import (
	"errors"
	"go.uber.org/zap"
	"net/http"
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
