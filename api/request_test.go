package api_test

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"testing"
)

type requestBuilder struct {
	method string
	url    string
	vars   map[string]string
	body   []byte
}

func Request() *requestBuilder {
	return &requestBuilder{
		method: http.MethodGet,
		url:    "http://example.com",
		vars:   make(map[string]string),
	}
}

func (r *requestBuilder) Url(url string) *requestBuilder {
	r.url = url
	return r
}

func (r *requestBuilder) Method(method string) *requestBuilder {
	r.method = method
	return r
}

func (r *requestBuilder) Build(t *testing.T) *http.Request {
	parsedUrl, err := url.Parse(r.url)
	if err != nil {
		t.Fatalf("error building mock request: %v", err)
	}

	request := &http.Request{
		URL:    parsedUrl,
		Method: r.method,
		Body:   &mockRequestBody{buf: bytes.NewBuffer(r.body)},
	}
	for k, v := range r.vars {
		ctx := request.Context()
		ctx = context.WithValue(ctx, k, v)
		request = request.WithContext(ctx)
	}

	return request
}

type mockRequestBody struct {
	buf *bytes.Buffer
}

func (b *mockRequestBody) Read(buf []byte) (int, error) {
	return b.buf.Read(buf)
}

func (b *mockRequestBody) Close() error {
	return nil
}
