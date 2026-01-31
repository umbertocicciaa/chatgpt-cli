# ChatGPT CLI

A modern, extensible command-line interface for ChatGPT with subcommand support, configuration management, and comprehensive logging.

## 🚀 Quick Start

```bash
# 1. Set your API key
export OPENAI_API_KEY="sk-your-api-key-here"

# 2. Build
go build -o chatgpt-cli main.go

# 3. Use it!
chatgpt-cli prompt "Explain Go channels"
```

## 📖 Usage

### Basic Command Structure

```bash
chatgpt-cli <command> [arguments]
```

### Available Commands

#### 1. Help Command

```bash
chatgpt-cli help
```

Displays usage information and all available commands.

#### 2. Prompt Command

```bash
chatgpt-cli prompt "your prompt here"
```

Send a prompt to ChatGPT and get a response.

**Examples:**

```bash
chatgpt-cli prompt "What is Go?"
chatgpt-cli prompt "Write a function to reverse a string"
chatgpt-cli prompt "Explain async/await in JavaScript"
```

#### 3. Logs Command

```bash
chatgpt-cli logs
```

Display all stored application logs including prompts, responses, and errors.

**Example Output:**

```plaintext
Showing 2 log entries:

[1] 2024-01-31 14:30:15 - prompt
    Prompt: What is Go?
    Response: Go is a statically typed, compiled programming language...

[2] 2024-01-31 14:35:22 - prompt
    Prompt: Explain channels
    Response: Channels in Go are a typed conduit through which you can send...
```

#### 4. Config Commands

**List all configuration:**

```bash
chatgpt-cli config list
```

**Get a specific value:**

```bash
chatgpt-cli config get OPENAI_MODEL
chatgpt-cli config get OPENAI_API_URL
```

**Set a configuration value:**

```bash
chatgpt-cli config set OPENAI_MODEL gpt-4
chatgpt-cli config set OPENAI_MAX_TOKENS 2000
chatgpt-cli config set OPENAI_TEMPERATURE 1.5
```

## ⚙️ Configuration

### Environment Variables

All configuration is managed via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | Your OpenAI API key | **(required)** |
| `OPENAI_API_URL` | API endpoint URL | `https://api.openai.com/v1/chat/completions` |
| `OPENAI_MODEL` | Model to use | `gpt-3.5-turbo` |
| `OPENAI_TIMEOUT` | Request timeout | `60s` |
| `OPENAI_MAX_TOKENS` | Max tokens in response | `1000` |
| `OPENAI_TEMPERATURE` | Response randomness (0.0-2.0) | `0.7` |
| `CHATGPT_CLI_CONFIG_DIR` | Config directory | `~/.chatgpt-cli` |

### Setting Environment Variables

**For current session:**

```bash
export OPENAI_MODEL="gpt-4"
export OPENAI_MAX_TOKENS="2000"
```

**Permanently (add to `~/.bashrc` or `~/.zshrc`):**

```bash
echo 'export OPENAI_API_KEY="sk-your-key"' >> ~/.bashrc
echo 'export OPENAI_MODEL="gpt-4"' >> ~/.bashrc
source ~/.bashrc
```

### Runtime Configuration

You can also set configuration at runtime (session only):

```bash
chatgpt-cli config set OPENAI_MODEL gpt-4
chatgpt-cli prompt "Now using GPT-4!"
```

**Note:** Runtime config changes are **not persisted**. To make them permanent, add them to your shell profile.

## 📂 File Locations

- **Config Directory**: `~/.chatgpt-cli/` (or `$CHATGPT_CLI_CONFIG_DIR`)
- **Logs**: `~/.chatgpt-cli/logs.jsonl`

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test -v

# Run tests with coverage
go test -v -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestConfigCommand

# Run benchmarks
go test -bench=.
```

## 🛠️ Development

### Adding New Commands

Adding a new command is straightforward:

1. **Create the command handler function:**

```go
func myNewCommand(config *Config, args []string) error {
    // Your command logic here
    fmt.Println("Executing my new command!")
    return nil
}
```

1. **Register the command in `getCommands()`:**

```go
func getCommands() map[string]Command {
    return map[string]Command{
        // ... existing commands ...
        "mynew": {
            Name:        "mynew",
            Description: "My new command",
            Handler:     myNewCommand,
        },
    }
}
```

1. **Update help text in `helpCommand()`** to document your new command.

2. **Write tests** in `main_test.go`.

That's it! Your command is now available:

```bash
chatgpt-cli mynew
```

### Project Structure

```plaintext
chatgpt-cli/
├── main.go          # Main application (~700 lines)
├── main_test.go     # Comprehensive tests (~650 lines)
├── go.mod           # Go module file
├── README.md        # This file
├── Makefile         # Build automation
└── .gitignore       # Git ignore rules
```

### Code Organization

The code is organized into logical sections:

1. **Constants & Types**: Configuration and data structures
2. **Configuration Loading**: Environment variable parsing
3. **Command Handlers**: Individual command implementations
4. **API Communication**: HTTP request handling
5. **Logging**: Event logging system
6. **Command Registry**: Command registration and routing
7. **Main Function**: Entry point

## 📊 Architecture

### Command Flow

```plaintext
User Input
    ↓
parseCommand() → Extracts command name and args
    ↓
getCommands() → Retrieves command registry
    ↓
Command.Handler() → Executes appropriate handler
    ↓
Result/Error
```

### Configuration Flow

```plaintext
Environment Variables
    ↓
loadConfig() → Parse with defaults
    ↓
Config struct → Passed to all commands
    ↓
Runtime modifications (config set) → Update env vars
```

## 💡 Examples

### Basic Usage

```bash
# Get help
chatgpt-cli help

# Ask a question
chatgpt-cli prompt "What are Go interfaces?"

# View logs
chatgpt-cli logs

# Check configuration
chatgpt-cli config list
```

### Advanced Usage

```bash
# Use GPT-4 for a single query
OPENAI_MODEL=gpt-4 chatgpt-cli prompt "Explain quantum computing"

# Set longer timeout for complex queries
OPENAI_TIMEOUT=120s chatgpt-cli prompt "Write a complete REST API server"

# Save response to file
chatgpt-cli prompt "Write a README for my project" > README.md

# Use in a script
RESPONSE=$(chatgpt-cli prompt "Generate a git commit message")
git commit -m "$RESPONSE"
```

### Configuration Management

```bash
# View current configuration
chatgpt-cli config list

# Get specific value
MODEL=$(chatgpt-cli config get OPENAI_MODEL)
echo "Current model: $MODEL"

# Change model for current session
chatgpt-cli config set OPENAI_MODEL gpt-4
chatgpt-cli prompt "Now using GPT-4"

# Set multiple values
chatgpt-cli config set OPENAI_MAX_TOKENS 2000
chatgpt-cli config set OPENAI_TEMPERATURE 1.2
```

## 🔒 Security

- **API Key Storage**: Never commit your API key to version control
- **Environment Variables**: Use environment variables for sensitive data
- **Session-only Config**: Runtime config changes don't persist to disk
- **Masked Display**: API key is masked in `config list` output
- **HTTPS**: All API communication is encrypted

## 🐛 Troubleshooting

### "Missing API key" error

```bash
# Make sure it's set
echo $OPENAI_API_KEY

# If not set
export OPENAI_API_KEY="sk-your-key"
```

### "Unknown command" error

```bash
# Check available commands
chatgpt-cli help

# Make sure you're using the right syntax
chatgpt-cli prompt "text"  # Correct
chatgpt-cli "text"          # Incorrect (missing 'prompt' command)
```

### Timeout errors

```bash
# Increase timeout
export OPENAI_TIMEOUT="120s"
# or
chatgpt-cli config set OPENAI_TIMEOUT 120s
```

### Rate limiting

If you encounter rate limit errors, wait a few moments before trying again. Check your usage at <https://platform.openai.com/account/usage>

## 📝 Changelog

### .0 (2024)

- **Breaking Change**: Now requires subcommands (e.g., `prompt`, `logs`)
- Added configuration management commands
- Added comprehensive logging system
- Added runtime configuration via `config set`
- All settings now configurable via environment variables
- Improved error messages and validation
- Full test coverage for all features
- Extensible command architecture

### v1.0.0 (2024)

- Initial release with basic functionality

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass (`go test -v`)
5. Submit a pull request

## 📄 License

MIT License - feel free to use and modify as needed.

## 🙏 Acknowledgments

- Built with Go's excellent standard library
- Powered by OpenAI's ChatGPT API
- Inspired by modern CLI design patterns

## 📚 Further Reading

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Go Documentation](https://go.dev/doc/)
- [Go Testing Guide](https://go.dev/doc/tutorial/add-a-test)

---

**Need help?** Run `chatgpt-cli help` or open an issue on GitHub.
