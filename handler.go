package trainwatcher

import (
	"encoding/json"
	"log"
	"net/http"
)

type errorResponse struct {
	Message string `json:"message"`
}

type health struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := health{
		Status: "healthy",
	}

	resp, err := http.Get(apiEndpoint)
	if err != nil {
		health.Status = "unhealthy"
	}
	defer resp.Body.Close()

	if err == nil && resp.StatusCode != http.StatusOK {
		health.Status = "unhealthy"
	}

	sendJSON(w, http.StatusOK, health)
}

type delay struct {
	Name           string
	Company        string
	LastUpdatedGMT int
	Source         string
}

type delaying struct {
	Name    string
	Company string
}

func delayHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		log.Println(err)
		msg := errorResponse{
			Message: "Internal Server Error",
		}
		sendJSON(w, http.StatusInternalServerError, msg)
		return
	}
	defer resp.Body.Close()

	var statuses []delay
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&statuses); err != nil {
		log.Println(err)
		msg := errorResponse{
			Message: "Internal Server Error",
		}
		sendJSON(w, http.StatusInternalServerError, msg)
		return
	}

	result := make([]delaying, 0, len(statuses))
	for _, status := range statuses {
		for _, w := range watching {
			if status.Company == w {
				r := delaying{
					Name:    status.Name,
					Company: status.Company,
				}
				result = append(result, r)
			}
		}
	}

	sendJSON(w, http.StatusOK, result)
}

type watchingCompanies struct {
	Companies []string `json:"companies"`
}

func watchingHandler(w http.ResponseWriter, r *http.Request) {
	watchingCompanies := watchingCompanies{
		Companies: watching,
	}
	sendJSON(w, http.StatusOK, watchingCompanies)
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		log.Println("Failed to send JSON")
		log.Printf("JSON: %+v\n", v)
	}
}
