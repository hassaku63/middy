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
	if err := run(); err != nil {
		slog.Error("server exited with error", "err", err)
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
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url) //nolint:gosec // loopback probe of our own server
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		return err
	}

	llmClient, err := newLLMClient(cfg)
	if err != nil {
		return err
	}

	deps := httpapi.Deps{LLM: llmClient, Logger: logger}
	if dist, ok := web.Dist(); ok {
		deps.Static = dist
	} else {
		logger.Warn("frontend not embedded; UI routes answer 503 (run `make build`)")
	}

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           httpapi.NewHandler(deps),
		ReadHeaderTimeout: 5 * time.Second,
	}

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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// newLLMClient selects the LLM provider from configuration.
// The real Anthropic-backed client is intentionally not wired yet; see
// docs/architecture/llm-integration.md and the backlog.
func newLLMClient(cfg config.Config) (llm.Client, error) {
	switch cfg.LLMProvider {
	case config.LLMProviderFake:
		return fake.New(), nil
	default:
		return nil, errors.New("unsupported LLM_PROVIDER: " + cfg.LLMProvider)
	}
}
