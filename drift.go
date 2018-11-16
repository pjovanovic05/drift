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

var (
	// client stuff
	target1 string
	target2 string
	t1fccfg string
	t2fccfg string
)

func main() {
	fmt.Println("vim-go")

	isServer := flag.Bool("d", false, "Run as daemon.")
	runConfig := flag.String("config", "run-config.json", "JSON config of targets to compare.")
	reportFN := flag.String("o", "drift-report.html", "File name for the report to be generated.")
	flag.Parse()

	if *isServer {
		startServer()
	} else {
		startClient(*runConfig, *reportFN)
	}
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/checkers/FileChecker/start", startFileChecker).Methods("POST") // TODO: ovo je mozda bolje da bude put ili post
	router.HandleFunc("/checkers/FileChecker/status", getFCStatus).Methods("GET")
	router.HandleFunc("/checkers/FileChecker/results", getFCResults).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// CLI client that takes json config of hosts to target, and generates html report.
func startClient(runConf, reportFN string) {

}

func startFileChecker(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	fc.Collect(config)
	// data, err := json.Marshal(hrs)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(data)
}

func getFCStatus(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func getFCResults(w http.ResponseWriter, r *http.Request) {
	// TODO:
}
