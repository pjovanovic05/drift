package main

import (
	"drift/checker"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var (
	fc     checker.FileChecker
	pmc    checker.PackageChecker
	aclc   checker.ACLChecker
	uc     checker.UserChecker
	passwd string
)

//StatusRep is checker status report.
type StatusRep struct {
	Host     string
	Progress string
}

func startServer(port int, password, cert, key string) {
	passwd = password
	router := mux.NewRouter()
	router.HandleFunc("/checkers/FileChecker/start", startFileChecker).Methods("POST")
	router.HandleFunc("/checkers/FileChecker/status", getFCStatus).Methods("GET")
	router.HandleFunc("/checkers/FileChecker/results", getFCResults).Methods("GET")
	router.HandleFunc("/checkers/PackageChecker/start", startPackageChecker).Methods("POST")
	router.HandleFunc("/checkers/PackageChecker/status", getPCStatus).Methods("GET")
	router.HandleFunc("/checkers/PackageChecker/results", getPCResults).Methods("GET")
	// log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(port), router))
	err := http.ListenAndServeTLS("0.0.0.0:"+strconv.Itoa(port), cert, key, router)
	if err != nil {
		log.Fatal(err)
	}
}

func basicAuth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !checkAuth(user, pass) {
			w.Header().Set("WWW-Authenticate",
				`Basic realm="please enter your username and password"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized.\n"))
			return
		}
		fn(w, r)
	}
}

func checkAuth(user, pass string) bool {
	return user == "admin" && pass == passwd
}

func startFileChecker(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	go fc.Collect(config)
	log.Println("Collecting files...")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}\n"))
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

func startPackageChecker(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	config["manager"] = "rpm"
	go pmc.Collect(config)
	log.Println("Collecting packages...")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}\n"))
}

func getPCStatus(w http.ResponseWriter, r *http.Request) {
	rep := StatusRep{Progress: pmc.Progress()}
	data, err := json.Marshal(rep)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getPCResults(w http.ResponseWriter, r *http.Request) {
	collected, err := pmc.GetCollected()
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

func startACLC(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	go aclc.Collect(config)
	log.Println("Collecting acls...")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}\n"))
}

func getACLCStatus(w http.ResponseWriter, r *http.Request) {
	rep := StatusRep{Progress: aclc.Progress()}
	data, err := json.Marshal(rep)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getACLCResults(w http.ResponseWriter, r *http.Request) {
	collected, err := aclc.GetCollected()
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

func startUC(w http.ResponseWriter, r *http.Request) {
	config := make(map[string]string)
	go uc.Collect(config)
	log.Println("Collecting users...")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}\n"))
}

func getUCStatus(w http.ResponseWriter, r *http.Request) {
	rep := StatusRep{Progress: uc.Progress()}
	data, err := json.Marshal(rep)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func getUCResults(w http.ResponseWriter, r *http.Request) {
	collected, err := uc.GetCollected()
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
