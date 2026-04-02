package cli

import (
	"fmt"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/api"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/spf13/cobra"
)

var Version = "v0.1.0"

var rootCmd = &cobra.Command{
	Use:     "claw",
	Version: Version,
	Short:   "Tarra Claw — AI agent CLI. Works with Ollama (free) or Anthropic.",
	Long: `Tarra Claw is an open-source AI coding agent CLI written in Go.

Providers:
  ollama     Free, local, no API key (default when ANTHROPIC_API_KEY is not set)
  anthropic  Claude API (requires ANTHROPIC_API_KEY)

Quick start with Ollama (free):
  brew install ollama && ollama serve &
  ollama pull llama3.2
  claw

Quick start with Anthropic:
  export ANTHROPIC_API_KEY=sk-ant-...
  claw`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(config.Init)
	rootCmd.PersistentFlags().StringP("provider", "p", "", "AI provider: ollama or anthropic")
	rootCmd.PersistentFlags().StringP("model", "m", "", "Model name (e.g. llama3.2, claude-opus-4-6)")
	rootCmd.PersistentFlags().String("ollama-host", "", "Ollama host (default: http://localhost:11434)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}

func listModels(cfg *config.Config) ([]string, error) {
	if cfg.Provider == "ollama" {
		models, err := api.ListOllamaModels(cfg.OllamaHost)
		if err != nil {
			return nil, fmt.Errorf("cannot connect to ollama: %w\nStart it with: ollama serve", err)
		}
		return models, nil
	}
	return []string{"claude-opus-4-6", "claude-sonnet-4-6", "claude-haiku-4-5-20251001"}, nil
}
