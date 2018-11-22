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
	PackageCheckerConf struct {
		Manager string `json:"manager"`
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
		err = startFC(runConfig)
		if err != nil {
			log.Fatalf("Error starting FileChecker on targets: %s\n", err)
		}
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

// Start File checker on targets.
func startFC(config RunConf) error {
	body, err := json.Marshal(config.FileCheckerConf)
	if err != nil {
		return err
	}
	leftURL := "http://" + config.Left.HostName + ":" +
		strconv.Itoa(config.Left.Port) + "/checkers/FileChecker/start"
	res, err := http.Post(leftURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()
	rightURL := "http://" + config.Right.HostName + ":" +
		strconv.Itoa(config.Right.Port) + "/checkers/FileChecker/start"
	res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res2.Body)
	res2.Body.Close()

	return nil
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

func startPC(config RunConf) error {
	rbody, err := json.Marshal(config.PackageCheckerConf)
	if err != nil {
		return err
	}
	leftURL := "http://" + config.Left.HostName + ":" +
		strconv.Itoa(config.Left.Port) + "/checkers/PackageChecker/start"
	res, err := http.Post(leftURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()
	rightURL := "http://" + config.Right.HostName + ":" +
		strconv.Itoa(config.Right.Port) + "/checkers/PackageChecker/start"
	res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res2.Body)
	res2.Body.Close()
	return nil
}

func fetchPCStatus(host Host, resc chan<- StatusRep, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(2 * time.Second)
		res, err := http.Get("http://" + host.HostName + ":" +
			strconv.Itoa(host.Port) + "/checkers/PackageChecker/status")
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

func fetchPCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get("http://" + host.HostName + ":" +
		strconv.Itoa(host.Port) + "/checker/PackageChecker/results")
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(res.Body).Decode(&ps)
	return
}
