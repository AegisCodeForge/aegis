package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func WebInstaller() {
	fmt.Println("This is the Aegis web installer. We will start a web server, which allows us to provide you a more user-friendly interface for configuring your Aegis instance. This web server will be shut down when the installation is finished. You can always start the web installer by using the `-init` flag or the `install` command.")
	var portNum int = 0
	for {
		r, err := askString("Please enter the port number this web server would bind to.", "8001")
		if err != nil {
			fmt.Printf("Failed to get a response: %s\n", err.Error())
			os.Exit(1)
		}
		portNum, err = strconv.Atoi(strings.TrimSpace(r))
		if err == nil { break }
		fmt.Println("Please enter a valid number...")
	}

	server := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", portNum),
	}
	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
	})
	go func() {
		log.Println(fmt.Sprintf("Trying to serve at %s:%d", "0.0.0.0", portNum))
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

