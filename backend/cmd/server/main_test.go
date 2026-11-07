package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ok.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer bad.Close()

	if err := probe(ok.URL, time.Second); err != nil {
		t.Fatalf("healthy server reported unhealthy: %v", err)
	}
	if err := probe(bad.URL, time.Second); err == nil {
		t.Fatal("503 must be unhealthy")
	}
	if err := probe("http://127.0.0.1:1/nothing", 200*time.Millisecond); err == nil {
		t.Fatal("connection failure must be unhealthy")
	}
}
