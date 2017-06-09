package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
)

const (
	headerContentType = "Content-Type"
	contentTypeJson   = "application/json"
)

// Middleware is a function that wraps and http.Handler to perform some
// kind of extra functionality, like providing gzip support, or logging
type Middleware func(http.Handler) http.Handler

// WrapHandlerInMiddleware wraps an http.Handler in multiple layers of Middleware.
// The first Middleware in the list is ran last, so logging (especially for
// timing) should be at the end of the list so that it is run last.
func WrapHandlerInMiddleware(base http.Handler, middleware ...Middleware) http.Handler {
	for _, mw := range middleware {
		base = mw(base)
	}
	return base
}

// CompressMiddleware will compress the output stream with gzip
// if the request supports it
func CompressMiddleware(fn http.Handler) http.Handler {
	return handlers.CompressHandler(fn)
}

// statusRecorder records the response http status code
// for logging purposes
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.status = status
	sr.ResponseWriter.WriteHeader(status)
}

// LoggingMiddleware will log the response status and time
// as well as the request method and url
func LoggingMiddleware(fn http.Handler) http.Handler {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w}
		start := time.Now()
		fn.ServeHTTP(recorder, r)
		end := time.Now()
		log.Printf("--> %d %s %s %s", recorder.status, r.Method, r.URL, end.Sub(start))
	}
	return http.HandlerFunc(handlerFunc)
}

// CORSMiddleware will provide CORS support for requests
func CORSMiddleware(fn http.Handler) http.Handler {
	return handlers.CORS()(fn)
}

// RecoveryMiddleware will recover from a panic during the response
func RecoveryMiddleware(message string) Middleware {
	return func(fn http.Handler) http.Handler {
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(message))
					log.Printf("ERROR: 500 %s %s", r.Method, r.URL)
				}
			}()
			fn.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handlerFunc)
	}
}

// ContentTypeMiddleware will set the default Content-Type
// header of responses
func ContentTypeMiddleware(contentType string) Middleware {
	return func(fn http.Handler) http.Handler {
		handlerFunc := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerContentType, contentType)
			fn.ServeHTTP(w, r)
		}
		return http.HandlerFunc(handlerFunc)
	}
}

// JSONContentTypeMiddleware will set the default Content-Type
// to application/json
func JSONContentTypeMiddleware(fn http.Handler) http.Handler {
	return ContentTypeMiddleware(contentTypeJson)(fn)
}
