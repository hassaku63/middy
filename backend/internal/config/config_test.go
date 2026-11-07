package config

import (
	"strings"
	"testing"
)

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    Config
		wantErr string
	}{
		{
			name: "defaults",
			env:  map[string]string{},
			want: Config{Port: 8080, LLMProvider: "fake", AnthropicModel: "claude-opus-5-5"},
		},
		{
			name: "custom port",
			env:  map[string]string{"PORT": "9090"},
			want: Config{Port: 9090, LLMProvider: "fake", AnthropicModel: "claude-opus-5-5"},
		},
		{
			name:    "invalid port",
			env:     map[string]string{"PORT": "abc"},
			wantErr: "PORT must be an integer",
		},
		{
			name:    "port out of range",
			env:     map[string]string{"PORT": "70000"},
			wantErr: "PORT must be an integer",
		},
		{
			name:    "anthropic requires api key",
			env:     map[string]string{"LLM_PROVIDER": "anthropic"},
			wantErr: "ANTHROPIC_API_KEY is required",
		},
		{
			name: "anthropic with key",
			env:  map[string]string{"LLM_PROVIDER": "anthropic", "ANTHROPIC_API_KEY": "k", "ANTHROPIC_MODEL": "m"},
			want: Config{Port: 8080, LLMProvider: "anthropic", AnthropicAPIKey: "k", AnthropicModel: "m"},
		},
		{
			name:    "unknown provider",
			env:     map[string]string{"LLM_PROVIDER": "nope"},
			wantErr: "unknown LLM_PROVIDER",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Load(envOf(tt.env))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("want error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestAddr(t *testing.T) {
	if got := (Config{Port: 1234}).Addr(); got != ":1234" {
		t.Fatalf("got %q", got)
	}
}
