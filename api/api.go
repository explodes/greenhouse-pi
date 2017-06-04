package api

import (
	"log"
	"net/http"
	"time"

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

type Api struct {
	Storage stats.Storage
}

type KnownStat struct {
	When  time.Time `json:"when"`
	Value float64 `json:"value"`
}

type varsHandler func(w http.ResponseWriter, r *http.Request, vars map[string]string)

func (vh varsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vh(w, r, vars)
}

func New(storage stats.Storage) *Api {
	return &Api{
		Storage: storage,
	}
}

func (api *Api) Serve(bind string) {

	router := mux.NewRouter()
	router.Handle("/{stat}/history/{start}/{end}", varsHandler(api.History))

	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: rwTimeout,
		ReadTimeout:  rwTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}
