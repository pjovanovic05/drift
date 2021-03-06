package main

import (
	"flag"
	"fmt"
	"log"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	isServer := flag.Bool("d", false, "Run as daemon.")
	host := flag.String("host", "0.0.0.0", "Server hostname or IP on which to listen to.")
	port := flag.Int("port", 8000, "Server port.")
	runConfig := flag.String("config", "run-config.json",
		"JSON config of targets to compare.")
	askPass := flag.Bool("p", false, "Ask for password.")
	passwd := flag.String("pass", "", "Password for clients to connect.")
	cert := flag.String("cert", "", "cert file for the server")
	key := flag.String("key", "", "key file for the server")
	reportFN := flag.String("o", "drift-report.html",
		"File name for the report to be generated.")
	flag.Parse()
	password := *passwd
	if *askPass {
		fmt.Print("Enter remote password: ")
		bytePass, err := terminal.ReadPassword(syscall.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println() // newline after password input
		password = string(bytePass)
	}

	if *isServer {
		startServer(*host, *port, password, *cert, *key)
	} else {
		startClient(*runConfig, *reportFN)
	}
}
