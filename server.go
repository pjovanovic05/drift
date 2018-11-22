package main

import (
	"drift/checker"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	fc  checker.FileChecker
	pmc checker.PackageChecker
)

//StatusRep is checker status report.
type StatusRep struct {
	Host     string
	Progress string
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/checkers/FileChecker/start", startFileChecker).Methods("POST")
	router.HandleFunc("/checkers/FileChecker/status", getFCStatus).Methods("GET")
	router.HandleFunc("/checkers/FileChecker/results", getFCResults).Methods("GET")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", router))
}

func startFileChecker(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	go fc.Collect(config)
	log.Println("Collection started...")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\":\"OK\"}\n"))
}

func getFCStatus(w http.ResponseWriter, r *http.Request) {
	rep := StatusRep{Progress: fc.Progress()}
	data, err := json.Marshal(rep)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getFCResults(w http.ResponseWriter, r *http.Request) {
	collected, err := fc.GetCollected()
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(collected)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func startRPMChecker(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	config["manager"] = "rpm"
	go pmc.Collect(config)
	log.Println("Collecting packages...")
	w.Header()
	w.Write([]byte(`{"Status": "OK"}\n`))
}
