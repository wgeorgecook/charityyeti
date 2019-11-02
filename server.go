package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func startServer() {
	log.Printf("New server started")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", parseResponse)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func parseResponse(w http.ResponseWriter, r *http.Request) {
	sn := r.URL.Query()["screenname"]
	hn := r.URL.Query()["honorary"]
	rt := r.URL.Query()["replyToURL"]

	log.Printf("Endpoint hit")
	log.Printf(fmt.Sprintf("Data: \n%v\n%v\n%v", sn, hn, rt))

	if len(sn) == 0 || len(hn) == 0 || len(rt) == 0 {
		fmt.Fprintf(w, "All requests must include 'screenname', 'honorary', and 'replyToURL' params")
	} else {
		fmt.Fprintf(w, fmt.Sprintf("{Data: { screenname: %v, honorary: %v, replyToURL: %v}}", sn[0], hn[0], rt[0]))
	}
}
