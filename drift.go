package main

import (
	"bytes"
	"drift/checker"
	"drift/differ"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

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
	FileCheckerConf struct {
		Path  string `json:"path"`
		Skips string `json:"skips"`
		Hash  string `json:"hash"`
	}
}

func main() {
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
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", router))
}

// CLI client that takes json config of hosts to target, and generates html report.
func startClient(runConf, reportFN string) {
	fmt.Println("Started client...")
	var wg sync.WaitGroup
	confStr, err := ioutil.ReadFile(runConf)
	if err != nil {
		log.Fatalf("Reading config file failed: %s\n", err)
	}
	fmt.Println("Config string:", string(confStr))
	runConfig := RunConf{}
	if err = json.Unmarshal(confStr, &runConfig); err != nil {
		log.Fatalf("JSON unmarshaling failed: %s\n", err)
	}
	fmt.Println(">>", runConfig.FileCheckerConf.Path, runConfig.FileCheckerConf.Skips, runConfig.FileCheckerConf.Hash)
	// start file checkers
	if runConfig.FileCheckerConf.Path != "" {
		body, err := json.Marshal(runConfig.FileCheckerConf)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("request body:", body)
		// start check on left
		letfURL := "http://" + runConfig.Left.HostName + ":" + strconv.Itoa(runConfig.Left.Port) + "/checkers/FileChecker/start"
		res, err := http.Post(letfURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, res.Body)
		res.Body.Close()
		// start check on right
		rightURL := "http://" + runConfig.Right.HostName + ":" + strconv.Itoa(runConfig.Right.Port) + "/checkers/FileChecker/start"
		res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, res2.Body)
		res2.Body.Close()
	}
	// TODO: start other checkers
	resc := make(chan StatusRep)
	wg.Add(2)
	go checkFCProgress(runConfig.Left, resc, &wg)
	go checkFCProgress(runConfig.Right, resc, &wg)

	// closer
	go func() {
		wg.Wait()
		close(resc)
	}()

	for res := range resc {
		fmt.Println(res.Host + ": " + res.Progress)
	}

	// when done, get results
	psL, err := fetchFCResults(runConfig.Left)
	if err != nil {
		log.Fatal(err)
	}

	psR, err := fetchFCResults(runConfig.Right)
	if err != nil {
		log.Fatal(err)
	}
	// generate report

	ds, err := differ.Diff(psL, psR)
	if err != nil {
		log.Fatal(err)
	}

	// differ.SaveHTMLReport("test.html", ds)
	html, err := differ.GetHtmlReport(ds)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(html)
	err = ioutil.WriteFile("test.html", []byte(html), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func checkFCProgress(host Host, resc chan<- StatusRep, wg *sync.WaitGroup) {
	// TODO: goroutine koji proverava progress i pise to u neki kanal
	defer wg.Done()
	for {
		time.Sleep(2 * time.Second)
		res, err := http.Get("http://" + host.HostName + ":" + strconv.Itoa(host.Port) + "/checkers/FileChecker/status")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		rep := StatusRep{}
		err = json.NewDecoder(res.Body).Decode(&rep)
		rep.Host = host.HostName
		if err != nil {
			log.Fatal(err)
		}
		resc <- rep
		if rep.Progress == "done" {
			break
		}
	}
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

//StatusRep is checker status report.
type StatusRep struct {
	Host     string
	Progress string
}

func fetchFCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get("http://" + host.HostName + ":" + strconv.Itoa(host.Port) + "/checkers/FileChecker/results")
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(res.Body).Decode(&ps)
	return
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
