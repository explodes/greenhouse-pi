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
	dateInputFormats = []string{"2006-01-02", "2006-01-02T15:04:05-0700"}
)

type Api struct {
	storage stats.Storage
}

type statResponse struct {
	When  time.Time `json:"when"`
	Value float64 `json:"value"`
}

func New(storage stats.Storage) *Api {
	return &Api{
		storage: storage,
	}
}

func (api *Api) Serve(bind string) {

	router := mux.NewRouter()
	router.Handle("/{stat}/history/{start}/{end}", http.HandlerFunc(api.History))

	srv := &http.Server{
		Handler:      router,
		Addr:         bind,
		WriteTimeout: rwTimeout,
		ReadTimeout:  rwTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Fatal(srv.ListenAndServe())
}
