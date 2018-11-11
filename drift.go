package main

import (
	"drift/checker"

	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	fc checker.FileChecker
)

func main() {
	fmt.Println("vim-go")

	isServer := flag.Bool("d", false, "Run as daemon.")
	flag.Parse()

	if *isServer {
		startServer()
	} else {
		startClient()
	}
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/checkers/FileChecker", getFileChecker).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func startClient() {
	// TODO get server addresses and passwords
	// TODO connect and gather info from checkers
	// TODO calculate differences
	// TODO show diffs...
}

func getFileChecker(w http.ResponseWriter, r *http.Request) {
	// TODO fill exclusion map from request params
	skips := make(map[string]bool)
	hrs, err := fc.List("/", skips)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(hrs)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
