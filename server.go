package trainwatcher

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	apiEndpoint = "https://rti-giken.jp/fhc/api/train_tetsudo/delay.json"
)

var (
	watching []string
)

// Watcher watches train delay status and post to Slack if status changes
type Watcher struct {
	webhookURL string
	mux        *mux.Router
}

// NewWatcher returns new Watcher
func NewWatcher(webhook string) (Watcher, error) {
	if err := parseConfig(); err != nil {
		return Watcher{}, err
	}

	watcher := Watcher{
		webhookURL: webhook,
		mux:        newRouter(),
	}
	return watcher, nil
}

type config struct {
	Watching []string `json:"watching"`
}

func parseConfig() error {
	b, err := ioutil.ReadFile("config.json")
	if err != nil {
		return err
	}

	var c config
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	watching = c.Watching
	return nil
}

func newRouter() *mux.Router {
	mux := mux.NewRouter()
	mux.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	mux.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	mux.HandleFunc("/health", healthHandler).Methods("GET")
	mux.HandleFunc("/delay", delayHandler).Methods("GET")
	mux.HandleFunc("/watching", watchingHandler).Methods("GET")
	return mux
}

func (w Watcher) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	w.mux.ServeHTTP(writer, req)
}

type errorResponse struct {
	Message string `json:"message"`
}

type health struct {
	Status string `json:"status"`
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusNotFound, errorResponse{
		Message: "Not Found",
	})
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusMethodNotAllowed, errorResponse{
		Message: "Method Not Allowed",
	})
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
