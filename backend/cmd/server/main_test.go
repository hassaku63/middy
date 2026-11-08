package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hassaku63/middy/backend/internal/llm/anthropic"
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

func TestServerOutlastsAnIdeasTurn(t *testing.T) {
	srv := newServer(":0", http.NotFoundHandler())
	if srv.WriteTimeout <= 2*anthropic.DefaultTimeout {
		t.Fatalf("WriteTimeout %v cuts off a turn that retries the model once (2 × %v)", srv.WriteTimeout, anthropic.DefaultTimeout)
	}
	if srv.ReadTimeout == 0 || srv.IdleTimeout == 0 || srv.ReadHeaderTimeout == 0 {
		t.Fatalf("a slow or idle client could hold a connection forever: %+v", srv)
	}
}

func TestShutdownClosesWhatOutlivesTheBudget(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-release
	}))
	ts.Start()
	defer func() { close(release); ts.Close() }()
	go func() { _, _ = http.Get(ts.URL) }() //nolint:noctx // test client

	<-started
	if err := shutdown(ts.Config, slog.New(slog.NewJSONHandler(io.Discard, nil)), 50*time.Millisecond); err != nil {
		t.Fatalf("shutdown with a request still running = %v, want nil (a signal is not a failure)", err)
	}
}
