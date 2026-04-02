package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration for Tarra Claw.
type Config struct {
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float64
	SystemPrompt string
	WorkingDir  string
	Verbose     bool
	AutoApprove bool // bypass permission prompts
	MCPServers  []MCPServerConfig
}

type MCPServerConfig struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
}

var global *Config

// Init is called by cobra on startup.
func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	home, _ := os.UserHomeDir()
	viper.AddConfigPath(filepath.Join(home, ".config", "tarra-claw"))
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("TARRA")

	// Defaults
	viper.SetDefault("model", "claude-opus-4-6")
	viper.SetDefault("max_tokens", 8096)
	viper.SetDefault("temperature", 1.0)
	viper.SetDefault("system_prompt", defaultSystemPrompt())

	_ = viper.ReadInConfig()

	cwd, _ := os.Getwd()

	global = &Config{
		APIKey:       firstNonEmpty(viper.GetString("api_key"), os.Getenv("ANTHROPIC_API_KEY")),
		Model:        viper.GetString("model"),
		MaxTokens:    viper.GetInt("max_tokens"),
		Temperature:  viper.GetFloat64("temperature"),
		SystemPrompt: viper.GetString("system_prompt"),
		WorkingDir:   cwd,
		Verbose:      viper.GetBool("verbose"),
		AutoApprove:  viper.GetBool("auto_approve"),
	}
}

// Get returns the global config, initializing if needed.
func Get() *Config {
	if global == nil {
		Init()
	}
	return global
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func defaultSystemPrompt() string {
	return `You are Tarra Claw, an AI coding agent. You have access to tools to read and write files,
execute bash commands, search codebases, and more. Be concise, accurate, and helpful.
Always think step by step. When you make changes to code, explain what you changed and why.`
}
