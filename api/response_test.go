package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type responseWriterRecorder struct {
	Status  int
	Headers http.Header
	Body    *bytes.Buffer
}

type assert struct {
	t *testing.T
	w *responseWriterRecorder
}

func NewResponseWriterRecorder() *responseWriterRecorder {
	return &responseWriterRecorder{
		Status:  0,
		Headers: make(http.Header),
		Body:    bytes.NewBuffer([]byte{}),
	}
}

func (w *responseWriterRecorder) Header() http.Header {
	return w.Headers
}

func (w *responseWriterRecorder) Write(buf []byte) (int, error) {
	return w.Body.Write(buf)
}

func (w *responseWriterRecorder) WriteHeader(status int) {
	w.Status = status
}

func (w *responseWriterRecorder) String() string {
	return fmt.Sprintf("%d: %s", w.Status, string(w.Body.Bytes()))
}

func (w *responseWriterRecorder) DeserializeJsonBody(v interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), v)
}

func (w *responseWriterRecorder) Assert(t *testing.T) *assert {
	return &assert{
		t: t,
		w: w,
	}
}

func (a *assert) StatusEquals(t *testing.T, status int) *assert {
	if a.w.Status != status {
		t.Fatalf("unexpected status: got %d need %d\nbody=%s", a.w.Status, status, a.w)
	}
	return a
}

func (a *assert) JsonBodyEquals(t *testing.T, expected interface{}) *assert {
	serialized, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("unable to marshal expected json: %v", err)
	}
	body := a.w.Body.Bytes()
	if !reflect.DeepEqual(serialized, body) {
		t.Fatalf("Unexpected json.\nGOT:  %s\nNEED: %s", string(body), string(serialized))
	}
	return a
}

func (a *assert) StringBodyEquals(t *testing.T, expected string) *assert {
	body := string(a.w.Body.Bytes())
	if body != expected {
		t.Fatalf("Unexpected body.\nGOT:  %s\nNEED: %s", body, expected)

	}
	return a
}
