package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/gorilla/mux"
)

// initHttpClient returns an oauth-enabled http client
func initHttpClient() *http.Client {
	oauthConfig := oauth1.NewConfig(cfg.ConsumerKey, cfg.ConsumerSecret)
	token := oauth1.NewToken(cfg.AccessToken, cfg.AccessSecret)
	return oauthConfig.Client(oauth1.NoContext, token)
}

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	// define the new router, define paths, and handlers on the router
	router := mux.NewRouter()
	router.HandleFunc("/post/donate", successfulDonation)
	router.HandleFunc("/get", getRecord)
	router.HandleFunc("/get/health", getHealth)
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
