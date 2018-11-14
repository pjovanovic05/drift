package main

import (
	"drift/checker"
	"net"

	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zserge/webview"
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
	flag.Parse()

	if *isServer {
		startServer()
	} else {
		startClient()
	}
}

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/checkers/FileChecker/start", startFileChecker).Methods("GET") // TODO: ovo je mozda bolje da bude put ili post
	router.HandleFunc("/checkers/FileChecker/status", getFCStatus).Methods("GET")
	router.HandleFunc("/checkers/FileChecker/results", getFCResults).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func startClient() {
	router := mux.NewRouter()
	// TODO setup router
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./clientui"))) //TODO assetFS

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go func() {
		log.Fatal(http.Serve(ln, router))
	}()
	webview.Open("Drift v0.1", "http://"+ln.Addr().String()+"/index.html", 800, 600, true)
}

func startFileChecker(w http.ResponseWriter, r *http.Request) {
	// TODO fill exclusion map from request params
	// skips := make(map[string]bool)
	// hrs, err := fc.List("/", skips)
	// if err != nil {
	// 	log.Fatal(err)
	// }
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
