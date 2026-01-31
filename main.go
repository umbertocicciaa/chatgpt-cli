package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Environment variable names
const (
	envAPIKey      = "OPENAI_API_KEY"
	envAPIURL      = "OPENAI_API_URL"
	envModel       = "OPENAI_MODEL"
	envTimeout     = "OPENAI_TIMEOUT"
	envMaxTokens   = "OPENAI_MAX_TOKENS"
	envTemperature = "OPENAI_TEMPERATURE"
	envConfigDir   = "CHATGPT_CLI_CONFIG_DIR"
)

// Default configuration values
const (
	defaultAPIURL      = "https://api.openai.com/v1/chat/completions"
	defaultModel       = "gpt-3.5-turbo"
	defaultTimeout     = 60 * time.Second
	defaultMaxTokens   = 1000
	defaultTemperature = 0.7
)

// Application configuration
type Config struct {
	APIKey      string
	APIURL      string
	Model       string
	Timeout     time.Duration
	MaxTokens   int
	Temperature float64
	ConfigDir   string
}

// OpenAI API request/response structures
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Command   string    `json:"command"`
	Prompt    string    `json:"prompt,omitempty"`
	Response  string    `json:"response,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Handler     func(*Config, []string) error
}

// loadConfig loads configuration from config file and environment variables with defaults
func loadConfig() (*Config, error) {
	configDir := getConfigDir()

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load from config file first
	fileConfig := loadConfigFile(configDir)

	// Environment variables override file config
	config := &Config{
		APIKey:      getEnvOrFileConfig(envAPIKey, fileConfig["OPENAI_API_KEY"]),
		APIURL:      getEnvOrFileOrDefault(envAPIURL, fileConfig["OPENAI_API_URL"], defaultAPIURL),
		Model:       getEnvOrFileOrDefault(envModel, fileConfig["OPENAI_MODEL"], defaultModel),
		Timeout:     parseDurationOrDefault(getEnvOrFileConfig(envTimeout, fileConfig["OPENAI_TIMEOUT"]), defaultTimeout),
		MaxTokens:   parseIntOrDefault(getEnvOrFileConfig(envMaxTokens, fileConfig["OPENAI_MAX_TOKENS"]), defaultMaxTokens),
		Temperature: parseFloatOrDefault(getEnvOrFileConfig(envTemperature, fileConfig["OPENAI_TEMPERATURE"]), defaultTemperature),
		ConfigDir:   configDir,
	}

	return config, nil
}

// loadConfigFile loads configuration from file
func loadConfigFile(configDir string) map[string]string {
	configFile := filepath.Join(configDir, "config")
	config := make(map[string]string)

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config // File doesn't exist or can't be read, return empty config
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return config
}

// saveConfigFile saves configuration to file
func saveConfigFile(configDir string, config map[string]string) error {
	configFile := filepath.Join(configDir, "config")

	// Read existing config to preserve all values
	existingConfig := loadConfigFile(configDir)

	// Merge with new values
	for key, value := range config {
		existingConfig[key] = value
	}

	// Write config file
	var lines []string
	lines = append(lines, "# ChatGPT CLI Configuration")
	lines = append(lines, "# Generated on "+time.Now().Format("2006-01-02 15:04:05"))
	lines = append(lines, "")

	keys := []string{
		"OPENAI_API_KEY",
		"OPENAI_API_URL",
		"OPENAI_MODEL",
		"OPENAI_TIMEOUT",
		"OPENAI_MAX_TOKENS",
		"OPENAI_TEMPERATURE",
	}

	for _, key := range keys {
		if value, exists := existingConfig[key]; exists && value != "" {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return os.WriteFile(configFile, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

// getEnvOrFileConfig gets value from env var first, then file config
func getEnvOrFileConfig(envKey, fileValue string) string {
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return fileValue
}

// getEnvOrFileOrDefault gets value from env var, then file config, then default
func getEnvOrFileOrDefault(envKey, fileValue, defaultValue string) string {
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	if fileValue != "" {
		return fileValue
	}
	return defaultValue
}

// getConfigDir returns the configuration directory path
func getConfigDir() string {
	if dir := os.Getenv(envConfigDir); dir != "" {
		return dir
	}

	// Use user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return ".chatgpt-cli"
	}
	return filepath.Join(home, ".chatgpt-cli")
}

// Helper functions for environment variable parsing
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntOrDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseFloatOrDefault(value string, defaultValue float64) float64 {
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseDurationOrDefault(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}

// Command handlers

// helpCommand displays usage information
func helpCommand(config *Config, args []string) error {
	help := `ChatGPT CLI - Command Line Interface for ChatGPT

Usage:
  chatgpt-cli <command> [arguments]

Available Commands:
  help                    Show this help message
  prompt <text>           Send a prompt to ChatGPT
  logs                    Display application logs
  config list             List current configuration
  config get <key>        Get a configuration value
  config set <key> <val>  Set a configuration value

Examples:
  chatgpt-cli prompt "Explain Go interfaces"
  chatgpt-cli logs
  chatgpt-cli config list
  chatgpt-cli config set OPENAI_MODEL gpt-4

Configuration:
  Configuration is managed via environment variables:
    OPENAI_API_KEY       - Your OpenAI API key (required)
    OPENAI_API_URL       - API endpoint URL (default: %s)
    OPENAI_MODEL         - Model to use (default: %s)
    OPENAI_TIMEOUT       - Request timeout (default: %s)
    OPENAI_MAX_TOKENS    - Max tokens in response (default: %d)
    OPENAI_TEMPERATURE   - Response randomness 0.0-2.0 (default: %.1f)
    CHATGPT_CLI_CONFIG_DIR - Config directory (default: ~/.chatgpt-cli)

For more information, visit: https://github.com/umbertocicciaa/chatgpt-cli
`
	fmt.Printf(help, defaultAPIURL, defaultModel, defaultTimeout, defaultMaxTokens, defaultTemperature)
	return nil
}

// promptCommand sends a prompt to ChatGPT
func promptCommand(config *Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("prompt text is required\nUsage: chatgpt-cli prompt \"your prompt here\"")
	}

	// Validate API key
	if config.APIKey == "" {
		return fmt.Errorf("missing API key: %s environment variable not set", envAPIKey)
	}

	// Combine all arguments as the prompt
	prompt := strings.Join(args, " ")
	if strings.TrimSpace(prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}

	// Send request
	response, err := sendChatRequest(config, prompt)
	if err != nil {
		logEntry(config, "prompt", prompt, "", err.Error())
		return fmt.Errorf("failed to get response: %w", err)
	}

	// Format and display response
	content := formatResponse(response)
	fmt.Println(content)

	// Log successful interaction
	logEntry(config, "prompt", prompt, content, "")

	return nil
}

// logsCommand displays application logs
func logsCommand(config *Config, args []string) error {
	logFile := filepath.Join(config.ConfigDir, "logs.jsonl")

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("No logs found.")
		return nil
	}

	// Read log file
	data, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read logs: %w", err)
	}

	if len(data) == 0 {
		fmt.Println("No logs found.")
		return nil
	}

	// Parse and display logs
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	fmt.Printf("Showing %d log entries:\n\n", len(lines))

	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid entries
		}

		fmt.Printf("[%d] %s - %s\n", i+1, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Command)
		if entry.Prompt != "" {
			fmt.Printf("    Prompt: %s\n", truncate(entry.Prompt, 80))
		}
		if entry.Response != "" {
			fmt.Printf("    Response: %s\n", truncate(entry.Response, 80))
		}
		if entry.Error != "" {
			fmt.Printf("    Error: %s\n", entry.Error)
		}
		fmt.Println()
	}

	return nil
}

// configCommand manages configuration
func configCommand(config *Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config subcommand required\nUsage: chatgpt-cli config <list|get|set>")
	}

	subcommand := args[0]

	switch subcommand {
	case "list":
		return configListCommand(config, args[1:])
	case "get":
		return configGetCommand(config, args[1:])
	case "set":
		return configSetCommand(config, args[1:])
	default:
		return fmt.Errorf("unknown config subcommand: %s\nValid subcommands: list, get, set", subcommand)
	}
}

// configListCommand lists all configuration values
func configListCommand(config *Config, args []string) error {
	fmt.Println("Current Configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Show API key masked
	apiKey := config.APIKey
	if apiKey != "" {
		if len(apiKey) > 8 {
			apiKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
		} else {
			apiKey = "***"
		}
	} else {
		apiKey = "(not set)"
	}

	fmt.Printf("%-25s %s\n", "OPENAI_API_KEY:", apiKey)
	fmt.Printf("%-25s %s\n", "OPENAI_API_URL:", config.APIURL)
	fmt.Printf("%-25s %s\n", "OPENAI_MODEL:", config.Model)
	fmt.Printf("%-25s %s\n", "OPENAI_TIMEOUT:", config.Timeout)
	fmt.Printf("%-25s %d\n", "OPENAI_MAX_TOKENS:", config.MaxTokens)
	fmt.Printf("%-25s %.1f\n", "OPENAI_TEMPERATURE:", config.Temperature)
	fmt.Printf("%-25s %s\n", "CHATGPT_CLI_CONFIG_DIR:", config.ConfigDir)

	return nil
}

// configGetCommand gets a specific configuration value
func configGetCommand(config *Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("configuration key required\nUsage: chatgpt-cli config get <key>")
	}

	key := strings.ToUpper(args[0])

	switch key {
	case "OPENAI_API_KEY":
		apiKey := config.APIKey
		if apiKey != "" {
			if len(apiKey) > 8 {
				apiKey = apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
			} else {
				apiKey = "***"
			}
		} else {
			apiKey = "(not set)"
		}
		fmt.Println(apiKey)
	case "OPENAI_API_URL":
		fmt.Println(config.APIURL)
	case "OPENAI_MODEL":
		fmt.Println(config.Model)
	case "OPENAI_TIMEOUT":
		fmt.Println(config.Timeout)
	case "OPENAI_MAX_TOKENS":
		fmt.Println(config.MaxTokens)
	case "OPENAI_TEMPERATURE":
		fmt.Println(config.Temperature)
	case "CHATGPT_CLI_CONFIG_DIR":
		fmt.Println(config.ConfigDir)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

// configSetCommand sets a configuration value
func configSetCommand(config *Config, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("both key and value required\nUsage: chatgpt-cli config set <key> <value>")
	}

	key := strings.ToUpper(args[0])
	value := args[1]

	// Validate the value
	switch key {
	case "OPENAI_API_KEY":
		if value == "" {
			return fmt.Errorf("API key cannot be empty")
		}

	case "OPENAI_API_URL":
		if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return fmt.Errorf("API URL must start with http:// or https://")
		}

	case "OPENAI_MODEL":
		if value == "" {
			return fmt.Errorf("model cannot be empty")
		}

	case "OPENAI_TIMEOUT":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("invalid timeout format (use format like '60s', '1m', '90s'): %w", err)
		}

	case "OPENAI_MAX_TOKENS":
		tokens, err := strconv.Atoi(value)
		if err != nil || tokens <= 0 {
			return fmt.Errorf("max tokens must be a positive integer")
		}

	case "OPENAI_TEMPERATURE":
		temp, err := strconv.ParseFloat(value, 64)
		if err != nil || temp < 0 || temp > 2 {
			return fmt.Errorf("temperature must be a number between 0.0 and 2.0")
		}

	case "CHATGPT_CLI_CONFIG_DIR":
		return fmt.Errorf("CHATGPT_CLI_CONFIG_DIR cannot be set via config set command. Use the environment variable instead.")

	default:
		return fmt.Errorf("unknown configuration key: %s\nValid keys: OPENAI_API_KEY, OPENAI_API_URL, OPENAI_MODEL, OPENAI_TIMEOUT, OPENAI_MAX_TOKENS, OPENAI_TEMPERATURE", key)
	}

	// Save to config file
	if err := saveConfigFile(config.ConfigDir, map[string]string{key: value}); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Set %s=%s\n", key, value)
	fmt.Printf("Configuration saved to %s\n", filepath.Join(config.ConfigDir, "config"))
	return nil
}

// sendChatRequest sends a request to the OpenAI API
func sendChatRequest(config *Config, prompt string) (*ChatResponse, error) {
	// Construct request payload
	requestBody := ChatRequest{
		Model: config.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.APIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	// Create client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code first
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, response: %s",
			resp.StatusCode, string(body))
	}

	// Parse response
	var chatResponse ChatResponse
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if chatResponse.Error != nil {
		return nil, fmt.Errorf("API error: %s (type: %s)",
			chatResponse.Error.Message, chatResponse.Error.Type)
	}

	return &chatResponse, nil
}

// formatResponse formats the ChatGPT response for display
func formatResponse(response *ChatResponse) string {
	if len(response.Choices) == 0 {
		return "No response received from ChatGPT"
	}

	content := response.Choices[0].Message.Content
	return strings.TrimSpace(content)
}

// logEntry logs an application event
func logEntry(config *Config, command, prompt, response, errorMsg string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Command:   command,
		Prompt:    prompt,
		Response:  response,
		Error:     errorMsg,
	}

	logFile := filepath.Join(config.ConfigDir, "logs.jsonl")

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return // Silent failure for logging
	}

	// Append to log file
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // Silent failure
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return // Silent failure
	}
	if _, err := f.WriteString("\n"); err != nil {
		return // Silent failure
	}
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getCommands returns all available commands
func getCommands() map[string]Command {
	return map[string]Command{
		"help": {
			Name:        "help",
			Description: "Show help message",
			Handler:     helpCommand,
		},
		"prompt": {
			Name:        "prompt",
			Description: "Send a prompt to ChatGPT",
			Handler:     promptCommand,
		},
		"logs": {
			Name:        "logs",
			Description: "Display application logs",
			Handler:     logsCommand,
		},
		"config": {
			Name:        "config",
			Description: "Manage configuration",
			Handler:     configCommand,
		},
	}
}

// parseCommand parses command-line arguments and returns the command and its arguments
func parseCommand(args []string) (string, []string) {
	if len(args) < 2 {
		return "help", []string{}
	}
	return args[1], args[2:]
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Parse command
	commandName, commandArgs := parseCommand(os.Args)

	// Get available commands
	commands := getCommands()

	// Find and execute command
	command, exists := commands[commandName]
	if !exists {
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", commandName)
		_ = helpCommand(config, []string{})
		os.Exit(1)
	}

	// Execute command
	if err := command.Handler(config, commandArgs); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
