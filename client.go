package main

import (
	"bytes"
	"drift/checker"
	"drift/differ"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
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

// CLI client that takes json config of hosts to target, and generates html report.
//
// As it is now:
// it reads run configuration,
// starts checkers on remote servers,
// starts progress polling goroutines,
// prints progress reports and waits for 'done' statuses,
// collects results from remotes,
// generates report using html template.
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
	// start file checkers
	if runConfig.FileCheckerConf.Path != "" {
		body, err := json.Marshal(runConfig.FileCheckerConf)
		if err != nil {
			log.Fatal(err)
		}
		// start check on left
		letfURL := "http://" + runConfig.Left.HostName + ":" +
			strconv.Itoa(runConfig.Left.Port) + "/checkers/FileChecker/start"
		res, err := http.Post(letfURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(os.Stdout, res.Body)
		res.Body.Close()
		// start check on right
		rightURL := "http://" + runConfig.Right.HostName + ":" +
			strconv.Itoa(runConfig.Right.Port) + "/checkers/FileChecker/start"
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

	//print progress reports
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
	ds, err := differ.Diff(psL, psR)
	if err != nil {
		log.Fatal(err)
	}

	html, err := differ.GetHtmlReport(ds)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("test.html", []byte(html), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func startFC(conf RunConf) {
	// TODO
}

func checkFCProgress(host Host, resc chan<- StatusRep, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(2 * time.Second)
		res, err := http.Get("http://" + host.HostName + ":" +
			strconv.Itoa(host.Port) + "/checkers/FileChecker/status")
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

func fetchFCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get("http://" + host.HostName + ":" +
		strconv.Itoa(host.Port) + "/checkers/FileChecker/results")
	if err != nil {
		log.Fatal(err)
	}
	err = json.NewDecoder(res.Body).Decode(&ps)
	return
}
