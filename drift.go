package main

import (
	"flag"
)

func main() {
	isServer := flag.Bool("d", false, "Run as daemon.")
	runConfig := flag.String("config", "run-config.json",
		"JSON config of targets to compare.")
	reportFN := flag.String("o", "drift-report.html",
		"File name for the report to be generated.")
	flag.Parse()

	if *isServer {
		startServer()
	} else {
		startClient(*runConfig, *reportFN)
	}
}
