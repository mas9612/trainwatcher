package main

import (
	"log"
	"net/http"

	"github.com/mas9612/trainwatcher"
)

func main() {
	webhookURL := "https://hooks.slack.com/services/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"
	watcher, err := trainwatcher.NewWatcher(webhookURL)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Starting API server...")
	log.Fatalln(http.ListenAndServe(":8080", watcher))
}
