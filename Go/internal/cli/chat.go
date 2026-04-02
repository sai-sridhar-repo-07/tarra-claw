package cli

import (
	"fmt"
	"os"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/engine"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/tui"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat [prompt]",
	Short: "Start an interactive chat session (default mode)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		if cfg.APIKey == "" {
			fmt.Fprintln(os.Stderr, "error: ANTHROPIC_API_KEY not set. Run: export ANTHROPIC_API_KEY=your-key")
			os.Exit(1)
		}

		eng, err := engine.New(cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize engine: %w", err)
		}

		// Non-interactive: single prompt passed as argument
		if len(args) > 0 {
			return eng.RunOnce(cmd.Context(), args[0])
		}

		// Interactive TUI mode
		m := tui.New(eng, cfg)
		return m.Run()
	},
}

var runCmd = &cobra.Command{
	Use:   "run <prompt>",
	Short: "Run a single prompt non-interactively and exit",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		eng, err := engine.New(cfg)
		if err != nil {
			return err
		}
		return eng.RunOnce(cmd.Context(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.RunE = chatCmd.RunE // default command is chat
}
