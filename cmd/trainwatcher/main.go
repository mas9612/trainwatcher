package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mas9612/trainwatcher"
)

func main() {
	webhookURL, ok := os.LookupEnv("SLACK_INCOMING_URL")
	if !ok {
		log.Fatalln("SLACK_INCOMING_URL not found")
	}

	watcher, err := trainwatcher.NewWatcher(webhookURL)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting API server...")
	log.Fatalln(http.ListenAndServe(":8080", watcher))
}
