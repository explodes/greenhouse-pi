package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/explodes/greenhouse-pi/controllers"
	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/stats"
	"github.com/gorilla/mux"
)

const (
	rwTimeout   = 15 * time.Second
	idleTimeout = 60 * time.Second
)

var (
	dateInputFormats = []string{
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05-0700",
		"2006-01-02",
	}
)

// Api is an object used to serve the JSON api for this system.
// See http://docs.greenhousepi.apiary.io for documentation
type Api struct {
	Storage stats.Storage
	Logger  logging.Logger
	Water   *controllers.Controller
	Fan     *controllers.Controller
}

// KnownStat is a stats.Stat but we know what stats.StatType it is already
type KnownStat struct {
	When  time.Time `json:"when"`
	Value float64   `json:"value"`
}

type varsHandler func(w http.ResponseWriter, r *http.Request, vars map[string]string)

func (vh varsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vh(w, r, vars)
}

func makeApiHandler(f func(w http.ResponseWriter, r *http.Request, vars map[string]string)) http.Handler {
	withVars := varsHandler(f)
	withGzip := CompressMiddleware(withVars)
	return withGzip
}

// New creates a new Api instance with the given storage
func New(storage stats.Storage, water, fan *controllers.Controller) *Api {
	return &Api{
		Storage: storage,
		Logger:  storage,
		Water:   water,
		Fan:     fan,
	}
}

// Serve will run this server and bind to the given address
func (api *Api) Serve(bind string) error {

	router := mux.NewRouter()
	router.Methods(http.MethodGet).Path("/{stat}/history/{start}/{end}").Handler(varsHandler(api.History))
	router.Methods(http.MethodGet).Path("/{stat}/latest").Handler(varsHandler(api.Latest))
	router.Methods(http.MethodGet).Path("/status").Handler(varsHandler(api.Status))
	router.Methods(http.MethodPost).Path("/{stat}/schedule/{start}/{end}").Handler(varsHandler(api.Status))
	router.Methods(http.MethodGet).Path("/logs/{level}/{start}/{end}").Handler(varsHandler(api.Logs))

	handler := WrapHandlerInMiddleware(router, CORSMiddleware, CompressMiddleware, LoggingMiddleware, RecoveryMiddleware)

	srv := &http.Server{
		Handler:      handler,
		Addr:         bind,
		WriteTimeout: rwTimeout,
		ReadTimeout:  rwTimeout,
		IdleTimeout:  idleTimeout,
	}

	if _, err := api.Storage.Log(logging.LevelInfo, "serving on %s", bind); err != nil {
		return fmt.Errorf("error logging server startup: %v", err)
	}

	return srv.ListenAndServe()
}
