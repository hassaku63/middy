// Command server runs the middy HTTP API.
package main

import (
	"context"
	"errors"
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
)

func main() {
	if err := run(); err != nil {
		slog.Error("server exited with error", "err", err)
		os.Exit(1)
	}
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

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           httpapi.NewHandler(httpapi.Deps{LLM: llmClient, Logger: logger}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", srv.Addr, "llm_provider", cfg.LLMProvider)
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
