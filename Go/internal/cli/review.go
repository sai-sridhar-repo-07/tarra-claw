package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sai-sridhar-repo-07/tarra-claw/internal/config"
	"github.com/sai-sridhar-repo-07/tarra-claw/internal/engine"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "AI code review of your current git diff",
	Long: `Runs git diff, sends it to the AI, and returns a structured code review.

Checks for: bugs, security issues, performance problems, style issues.

Examples:
  claw review              # review unstaged changes
  claw review --staged     # review staged (committed) changes
  claw review --branch main # compare current branch vs main`,
	RunE: runReview,
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "AI writes your git commit message from staged changes",
	Long: `Runs git diff --staged, sends it to the AI, and prints a commit message.
Does NOT commit — it just writes the message so you can review and use it.

Example:
  git add .
  claw commit`,
	RunE: runCommit,
}

func init() {
	reviewCmd.Flags().Bool("staged", false, "Review staged changes only")
	reviewCmd.Flags().String("branch", "", "Compare current branch against this branch (e.g. main)")
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(commitCmd)
}

func runReview(cmd *cobra.Command, args []string) error {
	staged, _ := cmd.Flags().GetBool("staged")
	branch, _ := cmd.Flags().GetString("branch")

	var diff string
	var err error

	if branch != "" {
		diff, err = gitDiff("diff", branch+"...HEAD")
	} else if staged {
		diff, err = gitDiff("diff", "--staged")
	} else {
		diff, err = gitDiff("diff", "HEAD")
	}

	if err != nil {
		return fmt.Errorf("git diff failed: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Println("No changes to review.")
		fmt.Println("Tip: make some edits, or use --staged to review staged changes.")
		return nil
	}

	lines := strings.Count(diff, "\n")
	fmt.Printf("Reviewing %d lines of diff...\n\n", lines)

	prompt := fmt.Sprintf(`You are an expert code reviewer. Review the following git diff carefully.

Structure your review as:

## Summary
One sentence describing what this change does.

## Issues Found
List any bugs, security vulnerabilities, performance problems, or logic errors.
If none, say "No issues found."

## Suggestions
Improvements to readability, naming, structure, or efficiency (non-blocking).

## Verdict
One of: ✅ Looks good  |  ⚠️ Minor issues  |  ❌ Needs changes

---

Git diff:
%s`, diff)

	cfg := config.Get()
	eng, err := engine.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	return eng.RunOnceDirect(cmd.Context(), prompt)
}

func runCommit(cmd *cobra.Command, args []string) error {
	diff, err := gitDiff("diff", "--staged")
	if err != nil {
		return fmt.Errorf("git diff failed: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Println("No staged changes found.")
		fmt.Println("Stage your changes first:")
		fmt.Println("  git add <files>")
		fmt.Println("  claw commit")
		return nil
	}

	// Also get list of changed files for context
	files, _ := gitOutput("git", "diff", "--staged", "--name-only")

	fmt.Println("Generating commit message...\n")

	prompt := fmt.Sprintf(`You are an expert at writing git commit messages.

Write a commit message for the following staged changes.

Rules:
- First line: short summary, max 72 chars, imperative mood (e.g. "Add", "Fix", "Remove")
- Leave a blank line after the first line
- Body: explain WHAT changed and WHY (not how), max 3-4 bullet points
- Use conventional commits format if applicable (feat:, fix:, refactor:, docs:, etc.)
- Do NOT include the diff itself in the message
- Output ONLY the commit message, nothing else

Changed files:
%s

Git diff:
%s`, files, diff)

	cfg := config.Get()
	eng, err := engine.New(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Println("─────────────────────────────────")
	err = eng.RunOnceDirect(cmd.Context(), prompt)
	fmt.Println("─────────────────────────────────")
	if err == nil {
		fmt.Println("\nTo use it:")
		fmt.Println("  git commit -m \"<paste message above>\"")
	}
	return err
}

func gitDiff(args ...string) (string, error) {
	return gitOutput("git", args...)
}

func gitOutput(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
