// Command server runs the middy HTTP API.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hassaku63/middy/backend/internal/config"
	"github.com/hassaku63/middy/backend/internal/httpapi"
	"github.com/hassaku63/middy/backend/internal/llm"
	"github.com/hassaku63/middy/backend/internal/llm/anthropic"
	"github.com/hassaku63/middy/backend/internal/llm/fake"
	"github.com/hassaku63/middy/backend/internal/web"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "probe a running server on $PORT and exit 0 if healthy (used by the container HEALTHCHECK)")
	flag.Parse()

	if *healthcheck {
		if err := runHealthcheck(); err != nil {
			fmt.Fprintln(os.Stderr, "unhealthy:", err)
			os.Exit(1)
		}
		return
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

// runHealthcheck exists because the runtime image (distroless) has neither
// a shell nor curl; the binary probes itself instead.
func runHealthcheck() error {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}
	return probe(fmt.Sprintf("http://127.0.0.1:%d%s/health", cfg.Port, httpapi.APIPrefix), 2*time.Second)
}

// probe fails unless url answers 200 within timeout.
func probe(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req) //nolint:gosec // loopback probe of our own server
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func run(logger *slog.Logger) error {
	slog.SetDefault(logger) // for libraries that log through the default

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	llmClient, err := newLLMClient(cfg, logger)
	if err != nil {
		return err
	}

	deps := httpapi.Deps{LLM: llmClient, Logger: logger}
	if dist, ok := web.Dist(); ok {
		deps.Static = dist
	} else {
		logger.Warn("frontend not embedded; UI routes answer 503 (run `make build`)")
	}

	srv := newServer(cfg.Addr(), httpapi.NewHandler(deps))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", srv.Addr, "llm_provider", cfg.LLMProvider, "frontend_embedded", deps.Static != nil)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		stop() // a second Ctrl-C or SIGTERM now ends the process at once
		return shutdown(srv, logger, shutdownTimeout)
	}
}

// Server timeouts. A turn of POST /ideas may call the model twice (one
// retry), each call bounded by anthropic.DefaultTimeout, so the response
// gets longer than that; reading a request (at most a few MiB of JSON)
// never needs long. Idle keep-alive connections are reaped.
const (
	readTimeout     = 30 * time.Second
	writeTimeout    = 2*anthropic.DefaultTimeout + 30*time.Second
	idleTimeout     = 2 * time.Minute
	shutdownTimeout = 10 * time.Second
)

func newServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// shutdown lets in-flight requests finish for budget (shutdownTimeout: a
// container stop waits about as long before it kills), then closes what is
// left. A
// request cut off that way is logged, but stopping on a signal is not a
// failure of the server.
func shutdown(srv *http.Server, logger *slog.Logger, budget time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()
	err := srv.Shutdown(ctx)
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Warn("requests still running at shutdown were cut off", "after", budget.String())
		return srv.Close()
	}
	return err
}

// newLLMClient selects the LLM provider from configuration.
func newLLMClient(cfg config.Config, logger *slog.Logger) (llm.Client, error) {
	switch cfg.LLMProvider {
	case config.LLMProviderFake:
		return fake.New(), nil
	case config.LLMProviderAnthropic:
		return anthropic.New(anthropic.Config{APIKey: cfg.AnthropicAPIKey, Model: cfg.AnthropicModel, Logger: logger})
	default:
		return nil, errors.New("unsupported LLM_PROVIDER: " + cfg.LLMProvider)
	}
}
