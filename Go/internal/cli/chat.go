package cli

import (
	"fmt"
	"os"

	"github.com/sai-sridhar-repo-07/forge/internal/config"
	"github.com/sai-sridhar-repo-07/forge/internal/engine"
	"github.com/sai-sridhar-repo-07/forge/internal/tui"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat [prompt]",
	Short: "Start an interactive chat session (default mode)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runChat,
}

var runCmd = &cobra.Command{
	Use:   "run <prompt>",
	Short: "Run a single prompt non-interactively and exit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		eng, err := engine.New(cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return eng.RunOnce(cmd.Context(), args[0])
	},
}

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available Ollama models",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		models, err := listModels(cfg)
		if err != nil {
			return err
		}
		if len(models) == 0 {
			fmt.Println("No models found. Pull one with: ollama pull llama3.2")
			return nil
		}
		fmt.Println("Available models:")
		for _, m := range models {
			fmt.Println(" ", m)
		}
		return nil
	},
}

func runChat(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	eng, err := engine.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "\n"+err.Error())
		os.Exit(1)
	}

	// Single prompt passed as argument — non-interactive
	if len(args) > 0 {
		return eng.RunOnce(cmd.Context(), args[0])
	}

	// Interactive TUI
	m := tui.New(eng, cfg)
	return m.Run()
}

func init() {
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.RunE = runChat
}
