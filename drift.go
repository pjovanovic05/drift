package main

import (
	"bytes"
	"drift/checker"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	fc checker.FileChecker
)

// Host describes one host for checking.
type Host struct {
	HostName string
	Password string
	Port     int
}

// RunConf holds run configuration for hosts to be checked and checks to be executed.
type RunConf struct {
	Left            Host
	Right           Host
	FileCheckerConf map[string]string `json:"omitempty"`
}

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
	router.HandleFunc("/checkers/FileChecker/start", startFileChecker).Methods("POST")
	router.HandleFunc("/checkers/FileChecker/status", getFCStatus).Methods("GET")
	router.HandleFunc("/checkers/FileChecker/results", getFCResults).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// CLI client that takes json config of hosts to target, and generates html report.
func startClient(runConf, reportFN string) {
	// TODO: load configuration
	confStr, err := ioutil.ReadFile(runConf)
	if err != nil {
		log.Fatalf("Reading config file failed: %s\n", err)
	}
	runConfig := RunConf{}
	if err = json.Unmarshal(confStr, &runConfig); err != nil {
		log.Fatalf("JSON unmarshaling failed: %s\n", err)
	}
	// connect to targets
	// start file checkers
	if runConfig.FileCheckerConf != nil {
		// start check on left
		// TODO: construct url for posting...
		letfURL := runConfig.Left.HostName + ":" + string(runConfig.Left.Port) + "/checkers/FileChecker/start"
		body, err := json.Marshal(runConfig.FileCheckerConf)
		if err != nil {
			log.Fatal(err)
		}
		res, err := http.Post(letfURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		// start check on right
		rightURL := runConfig.Right.HostName + ":" + string(runConfig.Right.Port) + "/checkers/FileChecker/start"
		res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatal(err)
		}
		defer res2.Body.Close()
	}
	// TODO: start other checkers

	// poll for progress -> to goroutines and channels
	// make get request for both targets
	// wait until done
	// when done, get results
	// generate report
}

func checkFCProgress(host Host, resc chan<- string) {
	// TODO: goroutine koji proverava progress i pise to u neki kanal
	res, err := http.Get(host.HostName + ":" + string(host.Port) + "/checkers/FileChecker/status")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

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
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"Progress": "` + fc.Progress() + `"}`))
}

func getFCResults(w http.ResponseWriter, r *http.Request) {
	// TODO:
}
