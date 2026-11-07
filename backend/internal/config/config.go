// Package config loads server configuration from the environment.
//
// Every setting has a safe local default so that `go run ./cmd/server`
// works with no environment at all. Secrets are never defaulted.
package config

import (
	"fmt"
	"strconv"
)

// Supported LLM providers.
const (
	LLMProviderFake = "fake"
	// LLMProviderAnthropic uses internal/llm/anthropic (requires ANTHROPIC_API_KEY).
	LLMProviderAnthropic = "anthropic"
)

// Config holds all runtime settings for the server.
type Config struct {
	// Port is the TCP port the HTTP server listens on.
	Port int
	// LLMProvider selects the llm.Client implementation ("fake" or "anthropic").
	LLMProvider string
	// AnthropicAPIKey is read only when LLMProvider is "anthropic".
	AnthropicAPIKey string
	// AnthropicModel is the model id used by the Anthropic adapter.
	AnthropicModel string
}

// Addr returns the listen address in host:port form.
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// Load reads configuration using getenv (usually os.Getenv). Passing the
// lookup function in makes Load trivially testable.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:            8080,
		LLMProvider:     LLMProviderFake,
		AnthropicAPIKey: getenv("ANTHROPIC_API_KEY"),
		AnthropicModel:  "claude-opus-5-5",
	}

	if v := getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p <= 0 || p > 65535 {
			return Config{}, fmt.Errorf("PORT must be an integer in 1..65535, got %q", v)
		}
		cfg.Port = p
	}
	if v := getenv("LLM_PROVIDER"); v != "" {
		cfg.LLMProvider = v
	}
	if v := getenv("ANTHROPIC_MODEL"); v != "" {
		cfg.AnthropicModel = v
	}

	switch cfg.LLMProvider {
	case LLMProviderFake:
	case LLMProviderAnthropic:
		if cfg.AnthropicAPIKey == "" {
			return Config{}, fmt.Errorf("ANTHROPIC_API_KEY is required when LLM_PROVIDER=%s", LLMProviderAnthropic)
		}
	default:
		return Config{}, fmt.Errorf("unknown LLM_PROVIDER %q", cfg.LLMProvider)
	}

	return cfg, nil
}
