package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration.
type Config struct {
	// Provider: "anthropic" or "ollama"
	Provider string

	// Anthropic settings
	APIKey string

	// Ollama settings
	OllamaHost  string // default: http://localhost:11434
	OllamaModel string // default: llama3.2

	// Shared
	Model        string
	MaxTokens    int
	Temperature  float64
	SystemPrompt string
	WorkingDir   string
	Verbose      bool
	AutoApprove  bool

	MCPServers []MCPServerConfig
}

type MCPServerConfig struct {
	Name    string
	Command string
	Args    []string
	Env     map[string]string
}

var global *Config

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	home, _ := os.UserHomeDir()
	viper.AddConfigPath(filepath.Join(home, ".config", "tarra-claw"))
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TARRA")

	// Defaults
	viper.SetDefault("provider", "")         // auto-detect
	viper.SetDefault("model", "claude-opus-4-6")
	viper.SetDefault("ollama_host", "http://localhost:11434")
	viper.SetDefault("ollama_model", "llama3.2")
	viper.SetDefault("max_tokens", 8096)
	viper.SetDefault("temperature", 1.0)
	viper.SetDefault("system_prompt", defaultSystemPrompt())

	_ = viper.ReadInConfig()

	cwd, _ := os.Getwd()

	apiKey := firstNonEmpty(viper.GetString("api_key"), os.Getenv("ANTHROPIC_API_KEY"))
	provider := firstNonEmpty(viper.GetString("provider"), os.Getenv("TARRA_PROVIDER"))

	// Auto-detect: if no API key, default to ollama
	if provider == "" {
		if apiKey != "" {
			provider = "anthropic"
		} else {
			provider = "ollama"
		}
	}

	ollamaModel := viper.GetString("ollama_model")
	model := viper.GetString("model")
	if provider == "ollama" {
		model = ollamaModel
	}

	global = &Config{
		Provider:     provider,
		APIKey:       apiKey,
		OllamaHost:   viper.GetString("ollama_host"),
		OllamaModel:  ollamaModel,
		Model:        model,
		MaxTokens:    viper.GetInt("max_tokens"),
		Temperature:  viper.GetFloat64("temperature"),
		SystemPrompt: viper.GetString("system_prompt"),
		WorkingDir:   cwd,
		Verbose:      viper.GetBool("verbose"),
		AutoApprove:  viper.GetBool("auto_approve"),
	}
}

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
Think step by step. When you change code, explain what changed and why.`
}
