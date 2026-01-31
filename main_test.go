package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test helper functions

func setupTestEnv(t *testing.T) func() {
	// Save original env vars
	originalVars := make(map[string]string)
	envVars := []string{
		envAPIKey, envAPIURL, envModel, envTimeout,
		envMaxTokens, envTemperature, envConfigDir,
	}

	for _, key := range envVars {
		originalVars[key] = os.Getenv(key)
		os.Unsetenv(key)
	}

	// Return cleanup function
	return func() {
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

func setTestEnv(key, value string) {
	os.Setenv(key, value)
}

// TestLoadConfig tests configuration loading with defaults
func TestLoadConfig(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name           string
		envVars        map[string]string
		expectedModel  string
		expectedURL    string
		expectedTokens int
		expectedTemp   float64
	}{
		{
			name:           "default values",
			envVars:        map[string]string{},
			expectedModel:  defaultModel,
			expectedURL:    defaultAPIURL,
			expectedTokens: defaultMaxTokens,
			expectedTemp:   defaultTemperature,
		},
		{
			name: "custom values",
			envVars: map[string]string{
				envModel:       "gpt-4",
				envAPIURL:      "https://custom.api.com/v1/chat",
				envMaxTokens:   "2000",
				envTemperature: "1.5",
			},
			expectedModel:  "gpt-4",
			expectedURL:    "https://custom.api.com/v1/chat",
			expectedTokens: 2000,
			expectedTemp:   1.5,
		},
		{
			name: "partial custom values",
			envVars: map[string]string{
				envModel: "gpt-4-turbo",
			},
			expectedModel:  "gpt-4-turbo",
			expectedURL:    defaultAPIURL,
			expectedTokens: defaultMaxTokens,
			expectedTemp:   defaultTemperature,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment before setting test values
			cleanup := setupTestEnv(t)
			defer cleanup()

			// Set environment variables
			for key, value := range tt.envVars {
				setTestEnv(key, value)
			}

			// Load config
			config, err := loadConfig()
			if err != nil {
				t.Fatalf("loadConfig() error = %v", err)
			}

			// Verify config values
			if config.Model != tt.expectedModel {
				t.Errorf("Model = %q, want %q", config.Model, tt.expectedModel)
			}
			if config.APIURL != tt.expectedURL {
				t.Errorf("APIURL = %q, want %q", config.APIURL, tt.expectedURL)
			}
			if config.MaxTokens != tt.expectedTokens {
				t.Errorf("MaxTokens = %d, want %d", config.MaxTokens, tt.expectedTokens)
			}
			if config.Temperature != tt.expectedTemp {
				t.Errorf("Temperature = %f, want %f", config.Temperature, tt.expectedTemp)
			}
		})
	}
}

// TestParseIntOrDefault tests integer parsing with defaults
func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		defVal   int
		expected int
	}{
		{"empty string", "", 100, 100},
		{"valid int", "500", 100, 500},
		{"invalid int", "abc", 100, 100},
		{"negative int", "-50", 100, -50},
		{"zero", "0", 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIntOrDefault(tt.input, tt.defVal)
			if result != tt.expected {
				t.Errorf("parseIntOrDefault(%q, %d) = %d, want %d",
					tt.input, tt.defVal, result, tt.expected)
			}
		})
	}
}

// TestParseFloatOrDefault tests float parsing with defaults
func TestParseFloatOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		defVal   float64
		expected float64
	}{
		{"empty string", "", 0.7, 0.7},
		{"valid float", "1.5", 0.7, 1.5},
		{"invalid float", "abc", 0.7, 0.7},
		{"integer as float", "2", 0.7, 2.0},
		{"zero", "0", 0.7, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFloatOrDefault(tt.input, tt.defVal)
			if result != tt.expected {
				t.Errorf("parseFloatOrDefault(%q, %f) = %f, want %f",
					tt.input, tt.defVal, result, tt.expected)
			}
		})
	}
}

// TestParseDurationOrDefault tests duration parsing with defaults
func TestParseDurationOrDefault(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		defVal   time.Duration
		expected time.Duration
	}{
		{"empty string", "", 60 * time.Second, 60 * time.Second},
		{"valid duration", "90s", 60 * time.Second, 90 * time.Second},
		{"invalid duration", "abc", 60 * time.Second, 60 * time.Second},
		{"minutes", "2m", 60 * time.Second, 2 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDurationOrDefault(tt.input, tt.defVal)
			if result != tt.expected {
				t.Errorf("parseDurationOrDefault(%q, %v) = %v, want %v",
					tt.input, tt.defVal, result, tt.expected)
			}
		})
	}
}

// TestParseCommand tests command parsing
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedCmd     string
		expectedCmdArgs []string
	}{
		{
			name:            "no arguments",
			args:            []string{"chatgpt-cli"},
			expectedCmd:     "help",
			expectedCmdArgs: []string{},
		},
		{
			name:            "help command",
			args:            []string{"chatgpt-cli", "help"},
			expectedCmd:     "help",
			expectedCmdArgs: []string{},
		},
		{
			name:            "prompt command with text",
			args:            []string{"chatgpt-cli", "prompt", "Hello", "World"},
			expectedCmd:     "prompt",
			expectedCmdArgs: []string{"Hello", "World"},
		},
		{
			name:            "config command",
			args:            []string{"chatgpt-cli", "config", "list"},
			expectedCmd:     "config",
			expectedCmdArgs: []string{"list"},
		},
		{
			name:            "logs command",
			args:            []string{"chatgpt-cli", "logs"},
			expectedCmd:     "logs",
			expectedCmdArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, cmdArgs := parseCommand(tt.args)

			if cmd != tt.expectedCmd {
				t.Errorf("command = %q, want %q", cmd, tt.expectedCmd)
			}

			if len(cmdArgs) != len(tt.expectedCmdArgs) {
				t.Errorf("cmdArgs length = %d, want %d", len(cmdArgs), len(tt.expectedCmdArgs))
				return
			}

			for i, arg := range cmdArgs {
				if arg != tt.expectedCmdArgs[i] {
					t.Errorf("cmdArgs[%d] = %q, want %q", i, arg, tt.expectedCmdArgs[i])
				}
			}
		})
	}
}

// TestGetCommands tests command registration
func TestGetCommands(t *testing.T) {
	commands := getCommands()

	expectedCommands := []string{"help", "prompt", "logs", "config"}

	for _, cmdName := range expectedCommands {
		if _, exists := commands[cmdName]; !exists {
			t.Errorf("command %q not found in registered commands", cmdName)
		}
	}

	if len(commands) != len(expectedCommands) {
		t.Errorf("expected %d commands, got %d", len(expectedCommands), len(commands))
	}
}

// TestHelpCommand tests the help command
func TestHelpCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	config := &Config{
		APIURL:      defaultAPIURL,
		Model:       defaultModel,
		Timeout:     defaultTimeout,
		MaxTokens:   defaultMaxTokens,
		Temperature: defaultTemperature,
	}

	err := helpCommand(config, []string{})
	if err != nil {
		t.Errorf("helpCommand() unexpected error: %v", err)
	}
}

// TestPromptCommand tests the prompt command
func TestPromptCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create temp config directory
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		config      *Config
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name: "no arguments",
			config: &Config{
				APIKey:    "test-key",
				ConfigDir: tmpDir,
			},
			args:        []string{},
			wantErr:     true,
			errContains: "prompt text is required",
		},
		{
			name: "empty prompt",
			config: &Config{
				APIKey:    "test-key",
				ConfigDir: tmpDir,
			},
			args:        []string{"   "},
			wantErr:     true,
			errContains: "prompt cannot be empty",
		},
		{
			name: "missing API key",
			config: &Config{
				APIKey:    "",
				ConfigDir: tmpDir,
			},
			args:        []string{"test prompt"},
			wantErr:     true,
			errContains: "missing API key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := promptCommand(tt.config, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("promptCommand() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("promptCommand() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestLogsCommand tests the logs command
func TestLogsCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create temp config directory
	tmpDir := t.TempDir()

	config := &Config{
		ConfigDir: tmpDir,
	}

	// Test with no logs
	err := logsCommand(config, []string{})
	if err != nil {
		t.Errorf("logsCommand() with no logs error: %v", err)
	}

	// Create a log file with entries
	logFile := filepath.Join(tmpDir, "logs.jsonl")
	entry1 := LogEntry{
		Timestamp: time.Now(),
		Command:   "prompt",
		Prompt:    "test prompt",
		Response:  "test response",
	}
	entry2 := LogEntry{
		Timestamp: time.Now(),
		Command:   "prompt",
		Prompt:    "another prompt",
		Error:     "test error",
	}

	f, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}

	enc := json.NewEncoder(f)
	enc.Encode(entry1)
	enc.Encode(entry2)
	f.Close()

	// Test with logs
	err = logsCommand(config, []string{})
	if err != nil {
		t.Errorf("logsCommand() with logs error: %v", err)
	}
}

// TestConfigCommand tests the config command
func TestConfigCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		APIKey:      "test-key",
		APIURL:      defaultAPIURL,
		Model:       defaultModel,
		Timeout:     defaultTimeout,
		MaxTokens:   defaultMaxTokens,
		Temperature: defaultTemperature,
		ConfigDir:   tmpDir,
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "no subcommand",
			args:        []string{},
			wantErr:     true,
			errContains: "config subcommand required",
		},
		{
			name:        "invalid subcommand",
			args:        []string{"invalid"},
			wantErr:     true,
			errContains: "unknown config subcommand",
		},
		{
			name:    "list subcommand",
			args:    []string{"list"},
			wantErr: false,
		},
		{
			name:        "get without key",
			args:        []string{"get"},
			wantErr:     true,
			errContains: "configuration key required",
		},
		{
			name:    "get valid key",
			args:    []string{"get", "OPENAI_MODEL"},
			wantErr: false,
		},
		{
			name:        "get invalid key",
			args:        []string{"get", "INVALID_KEY"},
			wantErr:     true,
			errContains: "unknown configuration key",
		},
		{
			name:        "set without value",
			args:        []string{"set", "OPENAI_MODEL"},
			wantErr:     true,
			errContains: "both key and value required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configCommand(config, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("configCommand() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("configCommand() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestConfigSetCommand tests setting configuration values
func TestConfigSetCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		ConfigDir: tmpDir,
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "set valid API key",
			args:    []string{"OPENAI_API_KEY", "sk-test123"},
			wantErr: false,
		},
		{
			name:        "set empty API key",
			args:        []string{"OPENAI_API_KEY", ""},
			wantErr:     true,
			errContains: "API key cannot be empty",
		},
		{
			name:    "set valid model",
			args:    []string{"OPENAI_MODEL", "gpt-4"},
			wantErr: false,
		},
		{
			name:    "set valid API URL",
			args:    []string{"OPENAI_API_URL", "https://api.example.com"},
			wantErr: false,
		},
		{
			name:        "set invalid API URL",
			args:        []string{"OPENAI_API_URL", "not-a-url"},
			wantErr:     true,
			errContains: "must start with http",
		},
		{
			name:    "set valid timeout",
			args:    []string{"OPENAI_TIMEOUT", "90s"},
			wantErr: false,
		},
		{
			name:        "set invalid timeout",
			args:        []string{"OPENAI_TIMEOUT", "invalid"},
			wantErr:     true,
			errContains: "invalid timeout format",
		},
		{
			name:    "set valid max tokens",
			args:    []string{"OPENAI_MAX_TOKENS", "2000"},
			wantErr: false,
		},
		{
			name:        "set invalid max tokens",
			args:        []string{"OPENAI_MAX_TOKENS", "-100"},
			wantErr:     true,
			errContains: "must be a positive integer",
		},
		{
			name:    "set valid temperature",
			args:    []string{"OPENAI_TEMPERATURE", "1.5"},
			wantErr: false,
		},
		{
			name:        "set invalid temperature",
			args:        []string{"OPENAI_TEMPERATURE", "3.0"},
			wantErr:     true,
			errContains: "between 0.0 and 2.0",
		},
		{
			name:        "set unknown key",
			args:        []string{"UNKNOWN_KEY", "value"},
			wantErr:     true,
			errContains: "unknown configuration key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configSetCommand(config, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("configSetCommand() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("configSetCommand() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestTruncate tests string truncation
func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello world",
			maxLen:   11,
			expected: "hello world",
		},
		{
			name:     "needs truncation",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestLogEntry tests logging functionality
func TestLogEntry(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		ConfigDir: tmpDir,
	}

	// Log an entry
	logEntry(config, "test-command", "test prompt", "test response", "")

	// Verify log file was created
	logFile := filepath.Join(tmpDir, "logs.jsonl")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file was not created")
		return
	}

	// Read and verify log content
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var entry LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}

	if entry.Command != "test-command" {
		t.Errorf("entry.Command = %q, want %q", entry.Command, "test-command")
	}
	if entry.Prompt != "test prompt" {
		t.Errorf("entry.Prompt = %q, want %q", entry.Prompt, "test prompt")
	}
	if entry.Response != "test response" {
		t.Errorf("entry.Response = %q, want %q", entry.Response, "test response")
	}
}

// TestFormatResponse tests response formatting
func TestFormatResponse(t *testing.T) {
	tests := []struct {
		name     string
		response *ChatResponse
		expected string
	}{
		{
			name: "valid response",
			response: &ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: "  Hello, world!  ",
						},
					},
				},
			},
			expected: "Hello, world!",
		},
		{
			name: "empty choices",
			response: &ChatResponse{
				Choices: []Choice{},
			},
			expected: "No response received from ChatGPT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatResponse(tt.response)
			if result != tt.expected {
				t.Errorf("formatResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestChatRequestMarshaling tests JSON marshaling
func TestChatRequestMarshaling(t *testing.T) {
	request := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello",
			},
		},
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ChatRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Model != request.Model {
		t.Errorf("Model = %q, want %q", unmarshaled.Model, request.Model)
	}
	if unmarshaled.MaxTokens != request.MaxTokens {
		t.Errorf("MaxTokens = %d, want %d", unmarshaled.MaxTokens, request.MaxTokens)
	}
	if unmarshaled.Temperature != request.Temperature {
		t.Errorf("Temperature = %f, want %f", unmarshaled.Temperature, request.Temperature)
	}
}

// TestGetConfigDir tests config directory determination
func TestGetConfigDir(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name      string
		envValue  string
		expectEnv bool
	}{
		{
			name:      "custom config dir from env",
			envValue:  "/custom/config/dir",
			expectEnv: true,
		},
		{
			name:      "default config dir",
			envValue:  "",
			expectEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnv(t)
			defer cleanup()

			if tt.envValue != "" {
				os.Setenv(envConfigDir, tt.envValue)
			}

			result := getConfigDir()

			if tt.expectEnv && result != tt.envValue {
				t.Errorf("getConfigDir() = %q, want %q", result, tt.envValue)
			}

			if !tt.expectEnv && result == "" {
				t.Errorf("getConfigDir() returned empty string")
			}
		})
	}
}

// TestGetEnvOrDefault tests environment variable retrieval
func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env var set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "env var not set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnvOrDefault(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestPromptCommandSuccess tests successful prompt execution
func TestPromptCommandSuccess(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := ChatResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "gpt-3.5-turbo",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role:    "assistant",
						Content: "Test response from ChatGPT",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	config := &Config{
		APIKey:      "test-api-key",
		APIURL:      server.URL,
		Model:       "gpt-3.5-turbo",
		Timeout:     10 * time.Second,
		MaxTokens:   1000,
		Temperature: 0.7,
		ConfigDir:   tmpDir,
	}

	err := promptCommand(config, []string{"test", "prompt"})
	if err != nil {
		t.Errorf("promptCommand() error = %v", err)
	}

	// Verify log was created
	logFile := filepath.Join(tmpDir, "logs.jsonl")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("log file was not created")
	}
}

// TestLogsCommandEmptyFile tests logs command with empty file
func TestLogsCommandEmptyFile(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		ConfigDir: tmpDir,
	}

	// Create empty log file
	logFile := filepath.Join(tmpDir, "logs.jsonl")
	f, err := os.Create(logFile)
	if err != nil {
		t.Fatalf("failed to create log file: %v", err)
	}
	f.Close()

	err = logsCommand(config, []string{})
	if err != nil {
		t.Errorf("logsCommand() error = %v", err)
	}
}

// TestConfigListCommand tests config list with various API key lengths
func TestConfigListCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()

	tests := []struct {
		name   string
		apiKey string
	}{
		{
			name:   "long API key",
			apiKey: "sk-1234567890abcdef",
		},
		{
			name:   "short API key",
			apiKey: "short",
		},
		{
			name:   "empty API key",
			apiKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				APIKey:      tt.apiKey,
				APIURL:      defaultAPIURL,
				Model:       defaultModel,
				Timeout:     defaultTimeout,
				MaxTokens:   defaultMaxTokens,
				Temperature: defaultTemperature,
				ConfigDir:   tmpDir,
			}

			err := configListCommand(config, []string{})
			if err != nil {
				t.Errorf("configListCommand() error = %v", err)
			}
		})
	}
}

// TestConfigGetCommand tests getting all configuration keys
func TestConfigGetCommand(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		APIKey:      "test-api-key",
		APIURL:      defaultAPIURL,
		Model:       defaultModel,
		Timeout:     defaultTimeout,
		MaxTokens:   defaultMaxTokens,
		Temperature: defaultTemperature,
		ConfigDir:   tmpDir,
	}

	validKeys := []string{
		"OPENAI_API_KEY",
		"OPENAI_API_URL",
		"OPENAI_MODEL",
		"OPENAI_TIMEOUT",
		"OPENAI_MAX_TOKENS",
		"OPENAI_TEMPERATURE",
		"CHATGPT_CLI_CONFIG_DIR",
	}

	for _, key := range validKeys {
		t.Run("get_"+key, func(t *testing.T) {
			err := configGetCommand(config, []string{key})
			if err != nil {
				t.Errorf("configGetCommand() for %s error = %v", key, err)
			}
		})
	}
}

// TestConfigSetCommandExtended tests additional config set scenarios
func TestConfigSetCommandExtended(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		ConfigDir: tmpDir,
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
		errContains string
	}{
		{
			name:        "set empty model",
			args:        []string{"OPENAI_MODEL", ""},
			wantErr:     true,
			errContains: "model cannot be empty",
		},
		{
			name:    "set valid API URL http",
			args:    []string{"OPENAI_API_URL", "http://localhost:8080"},
			wantErr: false,
		},
		{
			name:        "set invalid max tokens non-numeric",
			args:        []string{"OPENAI_MAX_TOKENS", "abc"},
			wantErr:     true,
			errContains: "must be a positive integer",
		},
		{
			name:        "set zero max tokens",
			args:        []string{"OPENAI_MAX_TOKENS", "0"},
			wantErr:     true,
			errContains: "must be a positive integer",
		},
		{
			name:    "set temperature at lower bound",
			args:    []string{"OPENAI_TEMPERATURE", "0.0"},
			wantErr: false,
		},
		{
			name:    "set temperature at upper bound",
			args:    []string{"OPENAI_TEMPERATURE", "2.0"},
			wantErr: false,
		},
		{
			name:        "set temperature above upper bound",
			args:        []string{"OPENAI_TEMPERATURE", "2.1"},
			wantErr:     true,
			errContains: "between 0.0 and 2.0",
		},
		{
			name:        "set temperature below lower bound",
			args:        []string{"OPENAI_TEMPERATURE", "-0.1"},
			wantErr:     true,
			errContains: "between 0.0 and 2.0",
		},
		{
			name:        "set invalid temperature non-numeric",
			args:        []string{"OPENAI_TEMPERATURE", "abc"},
			wantErr:     true,
			errContains: "between 0.0 and 2.0",
		},
		{
			name:        "set valid config dir",
			args:        []string{"CHATGPT_CLI_CONFIG_DIR", "/tmp/test-config"},
			wantErr:     true,
			errContains: "cannot be set via config set command",
		},
		{
			name:        "set empty config dir",
			args:        []string{"CHATGPT_CLI_CONFIG_DIR", ""},
			wantErr:     true,
			errContains: "cannot be set via config set command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := configSetCommand(config, tt.args)

			if tt.wantErr {
				if err == nil {
					t.Errorf("configSetCommand() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("configSetCommand() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSendChatRequest tests the HTTP request to OpenAI API
func TestSendChatRequest(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		wantErr       bool
		errContains   string
	}{
		{
			name: "successful request",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request headers
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type header = %q, want %q", r.Header.Get("Content-Type"), "application/json")
				}

				response := ChatResponse{
					ID:      "test-id",
					Object:  "chat.completion",
					Created: time.Now().Unix(),
					Model:   "gpt-3.5-turbo",
					Choices: []Choice{
						{
							Message: Message{
								Content: "Test response",
							},
						},
					},
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}),
			wantErr: false,
		},
		{
			name: "API error response",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := ChatResponse{
					Error: &APIError{
						Message: "Invalid API key",
						Type:    "invalid_request_error",
					},
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(response)
			}),
			wantErr:     true,
			errContains: "API error",
		},
		{
			name: "non-200 status code",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"error": "bad request"}`)
			}),
			wantErr:     true,
			errContains: "unexpected status code",
		},
		{
			name: "invalid JSON response",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "not json")
			}),
			wantErr:     true,
			errContains: "failed to parse response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			config := &Config{
				APIKey:      "test-key",
				APIURL:      server.URL,
				Model:       "gpt-3.5-turbo",
				Timeout:     10 * time.Second,
				MaxTokens:   1000,
				Temperature: 0.7,
			}

			_, err := sendChatRequest(config, "test prompt")

			if tt.wantErr {
				if err == nil {
					t.Errorf("sendChatRequest() expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want error containing %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("sendChatRequest() unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSendChatRequestTimeout tests timeout handling
func TestSendChatRequestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	config := &Config{
		APIKey:    "test-key",
		APIURL:    server.URL,
		Model:     "gpt-3.5-turbo",
		Timeout:   100 * time.Millisecond, // Very short timeout
		MaxTokens: 1000, Temperature: 0.7,
	}

	_, err := sendChatRequest(config, "test prompt")
	if err == nil {
		t.Errorf("sendChatRequest() expected timeout error, got nil")
	}
}

// TestFormatResponseExtended tests additional response formatting scenarios
func TestFormatResponseExtended(t *testing.T) {
	tests := []struct {
		name     string
		response *ChatResponse
		expected string
	}{
		{
			name: "response with newlines",
			response: &ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: "\n\nTest response\n\n",
						},
					},
				},
			},
			expected: "Test response",
		},
		{
			name: "multiple choices uses first",
			response: &ChatResponse{
				Choices: []Choice{
					{
						Message: Message{
							Content: "First choice",
						},
					},
					{
						Message: Message{
							Content: "Second choice",
						},
					},
				},
			},
			expected: "First choice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatResponse(tt.response)
			if result != tt.expected {
				t.Errorf("formatResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestChatResponseMarshaling tests JSON unmarshaling of API response
func TestChatResponseMarshaling(t *testing.T) {
	jsonData := `{
		"id": "test-id",
		"object": "chat.completion",
		"created": 1234567890,
		"model": "gpt-3.5-turbo",
		"choices": [{
			"index": 0,
			"message": {
				"role": "assistant",
				"content": "Test response"
			},
			"finish_reason": "stop"
		}],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 20,
			"total_tokens": 30
		}
	}`

	var response ChatResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.ID != "test-id" {
		t.Errorf("ID = %q, want %q", response.ID, "test-id")
	}
	if len(response.Choices) != 1 {
		t.Errorf("len(Choices) = %d, want 1", len(response.Choices))
	}
	if response.Choices[0].Message.Content != "Test response" {
		t.Errorf("Content = %q, want %q", response.Choices[0].Message.Content, "Test response")
	}
	if response.Usage.TotalTokens != 30 {
		t.Errorf("TotalTokens = %d, want 30", response.Usage.TotalTokens)
	}
}

// TestAPIErrorMarshaling tests API error response parsing
func TestAPIErrorMarshaling(t *testing.T) {
	jsonData := `{
		"error": {
			"message": "Invalid API key provided",
			"type": "invalid_request_error",
			"code": "invalid_api_key"
		}
	}`

	var response ChatResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if response.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if response.Error.Message != "Invalid API key provided" {
		t.Errorf("Error.Message = %q, want %q", response.Error.Message, "Invalid API key provided")
	}
	if response.Error.Type != "invalid_request_error" {
		t.Errorf("Error.Type = %q, want %q", response.Error.Type, "invalid_request_error")
	}
}

// TestLogEntryMultiple tests multiple log entries
func TestLogEntryMultiple(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	tmpDir := t.TempDir()
	config := &Config{
		ConfigDir: tmpDir,
	}

	// Log multiple entries
	logEntry(config, "prompt", "prompt 1", "response 1", "")
	logEntry(config, "prompt", "prompt 2", "", "error 2")
	logEntry(config, "config", "", "", "")

	// Verify log file
	logFile := filepath.Join(tmpDir, "logs.jsonl")
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 log entries, got %d", len(lines))
	}
}
