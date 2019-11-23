package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// startServer spins up an http listener for this service on the
// port and path specified
func startServer() {
	log.Printf("New server started")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", parseResponse)
	log.Fatal(http.ListenAndServe(":8080", router))
}

// parseResponse captures the get params off the incoming request
// We use this to get data following a successful donation
// Example:
// localhost:8080?screenname=@wgeorgecook&honorary=@charityyeti&replyToURL=https://twitter.com/WGeorgeCook/status/1197178917825630210&donationValue=5
func parseResponse(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query()["screenname"]
	hn := r.URL.Query()["honorary"]
	rt := r.URL.Query()["replyToURL"]
	dv := r.URL.Query()["donationValue"]

	log.Printf("Endpoint hit")
	log.Printf(fmt.Sprintf("Data: \n%v\n%v\n%v\n%v", sn, hn, rt, dv))

	if len(sn) == 0 || len(hn) == 0 || len(rt) == 0 || len(dv) == 0 {
		fmt.Fprintf(w, "All requests must include 'screenname', 'honorary', and 'replyToURL', and 'donationValue' params")
	} else {
		fmt.Fprintf(w, fmt.Sprintf("{Data: { screenname: %v, honorary: %v, replyToURL: %v, donationValue: %v}}", sn[0], hn[0], rt[0], dv[0]))
		err := respondToDonation(sn[0], hn[0], rt[0], dv[0])

		if err != nil {
			log.Printf(fmt.Sprintf(err.Error()))
		}
	}
}
