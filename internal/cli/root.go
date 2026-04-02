package cli

import (
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "claw",
	Short: "Tarra Claw — AI agent CLI harness in Go",
	Long: `Tarra Claw is an open-source AI coding agent CLI built in Go.
Concurrent tool execution, MCP protocol support, streaming Anthropic API,
and a Bubble Tea TUI — engineered for speed and clarity.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(config.Init)
	rootCmd.PersistentFlags().StringP("model", "m", "", "Claude model to use (overrides config)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().String("api-key", "", "Anthropic API key (overrides ANTHROPIC_API_KEY)")
}
