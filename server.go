package trainwatcher

import (
	"encoding/json"
	"io/ioutil"
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
	mux *mux.Router
}

// NewWatcher returns new Watcher
func NewWatcher() (Watcher, error) {
	if err := parseConfig(); err != nil {
		return Watcher{}, err
	}

	watcher := Watcher{
		mux: newRouter(),
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

func (w Watcher) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	w.mux.ServeHTTP(writer, req)
}
