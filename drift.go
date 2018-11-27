package main

import (
	"flag"
)

func main() {
	isServer := flag.Bool("d", false, "Run as daemon.")
	runConfig := flag.String("config", "run-config.json",
		"JSON config of targets to compare.")
	askPass := flag.Bool("p", false, "Ask for password.")
	passwd := flag.String("pass", "", "Password for clients to connect.")
	reportFN := flag.String("o", "drift-report.html",
		"File name for the report to be generated.")
	flag.Parse()

	if *isServer {
		startServer()
	} else {
		startClient(*runConfig, *reportFN)
	}
}

// TODO: get password: https://stackoverflow.com/questions/2137357/getpasswd-functionality-in-go
// TODO: make https connections
// TODO: make http basic auth
