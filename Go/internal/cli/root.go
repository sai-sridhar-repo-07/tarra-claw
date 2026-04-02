package cli

import (
	"fmt"

	"github.com/sai-sridhar-repo-07/forge/internal/api"
	"github.com/sai-sridhar-repo-07/forge/internal/config"
	"github.com/spf13/cobra"
)

var Version = "v0.3.0"

var rootCmd = &cobra.Command{
	Use:     "forge",
	Version: Version,
	Short:   "Forge — AI coding agent. Free with Ollama or Anthropic Claude.",
	Long: `Forge is an open-source AI coding agent CLI written in Go.

Works free with Ollama (no API key) or with Anthropic Claude.

Quick start — free with Ollama:
  brew install ollama && ollama serve &
  ollama pull llama3.2
  forge

Quick start — Anthropic Claude:
  export ANTHROPIC_API_KEY=sk-ant-...
  forge

Commands:
  forge            interactive AI chat
  forge review     AI code review of your git diff
  forge commit     AI writes your commit message
  forge run        one-shot prompt, exits when done
  forge models     list available models`,
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
