package main

import (
	"bytes"
	"crypto/tls"
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
	HostName  string
	Password  string
	Port      int
	SSL       bool
	KeyVerify bool
}

// GetBaseURL generates base url for http(s) requests on this host.
func (h *Host) GetBaseURL() string {
	protocol := "http"
	auth := ""
	if h.SSL {
		protocol += "s"
	}
	if h.Password != "" {
		auth = "admin:" + h.Password + "@"
	}
	return protocol + "://" + auth + h.HostName + ":" + strconv.Itoa(h.Port)
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
	ACLCheckerConf struct {
		Path  string `json:"path"`
		Skips string `json:"skips"`
	}
	UserCheckerConf struct {
		Pattern string
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
	resc := make(chan StatusRep)

	// Skip https key verification on client
	if !runConfig.Left.KeyVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// start file checkers
	if runConfig.FileCheckerConf.Path != "" {
		err = startFC(runConfig)
		if err != nil {
			log.Fatalf("Error starting FileChecker on targets: %s\n", err)
		}
		wg.Add(2)
		go fetchFCProgress(runConfig.Left, resc, &wg)
		go fetchFCProgress(runConfig.Right, resc, &wg)
	}

	if runConfig.PackageCheckerConf.Manager != "" {
		err = startPC(runConfig)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(2)
		go fetchPCStatus(runConfig.Left, resc, &wg)
		go fetchPCStatus(runConfig.Right, resc, &wg)
	}

	if runConfig.UserCheckerConf.Pattern != "" {
		err = startUC(runConfig)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(2)
		go fetchUCStatus(runConfig.Left, resc, &wg)
		go fetchUCStatus(runConfig.Right, resc, &wg)
	}

	if runConfig.ACLCheckerConf.Path != "" {
		err = startACLC(runConfig)
		if err != nil {
			log.Fatal(err)
		}
		wg.Add(2)
		go fetchACLCStatus(runConfig.Left, resc, &wg)
		go fetchACLCStatus(runConfig.Right, resc, &wg)
	}

	// closer, waits for status checks to finish
	go func() {
		wg.Wait()
		close(resc)
	}()

	//print progress reports
	for res := range resc {
		fmt.Println(res.Host + ": " + res.Progress)
	}

	// when done, get results and make reports
	if runConfig.FileCheckerConf.Path != "" {
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
		err = ioutil.WriteFile("files-report.html", []byte(html), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	if runConfig.PackageCheckerConf.Manager != "" {
		psL, err := fetchPCResults(runConfig.Left)
		if err != nil {
			log.Fatal(err)
		}
		psR, err := fetchPCResults(runConfig.Right)
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
		err = ioutil.WriteFile("packages-report.html", []byte(html), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	if runConfig.UserCheckerConf.Pattern != "" {
		psL, err := fetchUCResults(runConfig.Left)
		if err != nil {
			log.Fatal(err)
		}
		psR, err := fetchUCResults(runConfig.Right)
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
		err = ioutil.WriteFile("users-report.html", []byte(html), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	if runConfig.ACLCheckerConf.Path != "" {
		psL, err := fetchACLCResults(runConfig.Left)
		if err != nil {
			log.Fatal(err)
		}
		psR, err := fetchACLCResults(runConfig.Right)
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
		err = ioutil.WriteFile("acls-report.html", []byte(html), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// Start File checker on targets.
func startFC(config RunConf) error {
	body, err := json.Marshal(config.FileCheckerConf)
	if err != nil {
		return err
	}
	leftURL := config.Left.GetBaseURL() + "/checkers/FileChecker/start"
	res, err := http.Post(leftURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()
	rightURL := config.Right.GetBaseURL() + "/checkers/FileChecker/start"
	res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res2.Body)
	res2.Body.Close()

	return nil
}

func fetchFCProgress(host Host, resc chan<- StatusRep, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(2 * time.Second)
		res, err := http.Get(host.GetBaseURL() + "/checkers/FileChecker/status")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		rep := StatusRep{}
		err = json.NewDecoder(res.Body).Decode(&rep)
		if err != nil {
			log.Fatalf("Error unmarshalling status report for FileChecker: %s\n", err)
		}
		rep.Host = host.HostName
		resc <- rep
		if rep.Progress == "file collection done" {
			break
		}
	}
}

func fetchFCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get(host.GetBaseURL() + "/checkers/FileChecker/results")
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
	leftURL := config.Left.GetBaseURL() + "/checkers/PackageChecker/start"
	res, err := http.Post(leftURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()
	rightURL := config.Right.GetBaseURL() + "/checkers/PackageChecker/start"
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
		res, err := http.Get(host.GetBaseURL() + "/checkers/PackageChecker/status")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		rep := StatusRep{}
		err = json.NewDecoder(res.Body).Decode(&rep)
		if err != nil {
			log.Fatalf("Error unmarshalling status rep. for package checker [%s]: %s\n",
				host.HostName, err)
		}
		rep.Host = host.HostName
		resc <- rep
		if rep.Progress == "package collection done" {
			break
		}
	}
}

func fetchPCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get(host.GetBaseURL() + "/checkers/PackageChecker/results")
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(res.Body).Decode(&ps)
	return
}

func startACLC(config RunConf) error {
	rbody, err := json.Marshal(config.ACLCheckerConf)
	if err != nil {
		return err
	}
	leftURL := config.Left.GetBaseURL() + "/checkers/ACLChecker/start"
	res, err := http.Post(leftURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()
	rightURL := config.Right.GetBaseURL() + "/checkers/ACLChecker/start"
	res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res2.Body)
	res2.Body.Close()
	return nil
}

func fetchACLCStatus(host Host, resc chan<- StatusRep, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(2 * time.Second)
		res, err := http.Get(host.GetBaseURL() + "/checkers/ACLChecker/status")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		rep := StatusRep{}
		err = json.NewDecoder(res.Body).Decode(&rep)
		if err != nil {
			log.Fatalf("Error unmarshalling status rep. for acl checkers [%s]: %s\n",
				host.HostName, err)
		}
		rep.Host = host.HostName
		resc <- rep
		if rep.Progress == "acl collection done" {
			break
		}
	}
}

func fetchACLCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get(host.GetBaseURL() + "/checkers/ACLChecker/results")
	if err != nil {
		return nil, err
	}
	log.Println(">>> ACLC results")
	err = json.NewDecoder(res.Body).Decode(&ps)
	return
}

func startUC(config RunConf) error {
	rbody, err := json.Marshal(config.UserCheckerConf)
	if err != nil {
		return err
	}
	leftURL := config.Left.GetBaseURL() + "/checkers/UserChecker/start"
	res, err := http.Post(leftURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res.Body)
	res.Body.Close()
	rightURL := config.Right.GetBaseURL() + "/checkers/UserChecker/start"
	res2, err := http.Post(rightURL, "application/json", bytes.NewBuffer(rbody))
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, res2.Body)
	res2.Body.Close()
	return nil
}

func fetchUCStatus(host Host, resc chan<- StatusRep, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		time.Sleep(2 * time.Second)
		res, err := http.Get(host.GetBaseURL() + "/checkers/UserChecker/status")
		if err != nil {
			log.Fatal(err)
		}
		defer res.Body.Close()
		rep := StatusRep{}
		err = json.NewDecoder(res.Body).Decode(&rep)
		if err != nil {
			log.Fatalf("Error unmarshalling status rep. for user checkers [%s]: %s\n",
				host.HostName, err)
		}
		rep.Host = host.HostName
		resc <- rep
		if rep.Progress == "user collection done" {
			break
		}
	}
}

func fetchUCResults(host Host) (ps []checker.Pair, err error) {
	res, err := http.Get(host.GetBaseURL() + "/checkers/UserChecker/results")
	if err != nil {
		return nil, err
	}
	log.Println(">>> UC results")
	err = json.NewDecoder(res.Body).Decode(&ps)
	return
}
