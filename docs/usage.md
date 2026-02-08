# Usage

## Command Syntax

```bash
chatgpt-cli <command> [arguments]
```

Running `chatgpt-cli` without any arguments displays the help message.

## Available Commands

| Command | Description |
|---------|-------------|
| `help` | Show help message with all available commands |
| `prompt <text>` | Send a prompt to ChatGPT and display the response |
| `logs` | Display application logs |
| `config <subcommand>` | Manage configuration (list, get, set) |

---

## `help`

Displays usage information, all available commands, configuration options, and examples.

**Syntax:**

```bash
chatgpt-cli help
```

This is also the default command when no arguments are provided.

---

## `prompt`

Sends a prompt to ChatGPT using the OpenAI API and prints the response to standard output.

**Syntax:**

```bash
chatgpt-cli prompt <text>
```

**Arguments:**

| Argument | Required | Description |
|----------|----------|-------------|
| `text` | Yes | The prompt text to send to ChatGPT. Can be a quoted string or multiple words. |

**Behavior:**

- Requires the `OPENAI_API_KEY` to be set (via environment variable or config file).
- Multiple arguments are joined with spaces to form the prompt.
- The prompt cannot be empty or whitespace-only.
- Successful interactions are logged to the log file.
- Failed interactions are also logged with the error message.

**Examples:**

```bash
# Ask a simple question
chatgpt-cli prompt "What is Go?"

# Ask for code generation
chatgpt-cli prompt "Write a function to reverse a string in Go"

# Multi-word prompt without quotes
chatgpt-cli prompt Explain async await in JavaScript

# Override model for a single query using an environment variable
OPENAI_MODEL=gpt-4 chatgpt-cli prompt "Explain quantum computing"

# Save response to a file
chatgpt-cli prompt "Write a README for my project" > output.md

# Use in a script
RESPONSE=$(chatgpt-cli prompt "Generate a git commit message")
git commit -m "$RESPONSE"
```

**Errors:**

| Error | Cause |
|-------|-------|
| `prompt text is required` | No prompt text was provided |
| `prompt cannot be empty` | The prompt text is empty or whitespace-only |
| `missing API key` | `OPENAI_API_KEY` is not set |
| `failed to get response` | API request failed (network error, timeout, invalid key, etc.) |

---

## `logs`

Displays all stored application logs, including timestamps, prompts, responses, and errors.

**Syntax:**

```bash
chatgpt-cli logs
```

Logs are stored in JSONL format at `<config_dir>/logs.jsonl`. Each log entry contains:

- **Timestamp** — when the event occurred
- **Command** — the command that was executed
- **Prompt** — the prompt text (if applicable)
- **Response** — the ChatGPT response (if applicable)
- **Error** — the error message (if the command failed)

Long prompts and responses are truncated to 80 characters in the display output.

**Example Output:**

```
Showing 2 log entries:

[1] 2024-01-31 14:30:15 - prompt
    Prompt: What is Go?
    Response: Go is a statically typed, compiled programming language...

[2] 2024-01-31 14:35:22 - prompt
    Prompt: Explain channels
    Error: failed to get response: unexpected status code: 401
```

---

## `config`

Manages application configuration. Has three subcommands: `list`, `get`, and `set`.

**Syntax:**

```bash
chatgpt-cli config <subcommand> [arguments]
```

### `config list`

Lists all current configuration values. The API key is masked for security (shows first 4 and last 4 characters, or `***` for short keys).

```bash
chatgpt-cli config list
```

**Example Output:**

```
Current Configuration:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
OPENAI_API_KEY:           sk-a...b1c2
OPENAI_API_URL:           https://api.openai.com/v1/chat/completions
OPENAI_MODEL:             gpt-3.5-turbo
OPENAI_TIMEOUT:           1m0s
OPENAI_MAX_TOKENS:        1000
OPENAI_TEMPERATURE:       0.7
CHATGPT_CLI_CONFIG_DIR:   /home/user/.chatgpt-cli
```

### `config get`

Gets the current value of a specific configuration key. The key name is case-insensitive.

```bash
chatgpt-cli config get <key>
```

**Valid keys:** `OPENAI_API_KEY`, `OPENAI_API_URL`, `OPENAI_MODEL`, `OPENAI_TIMEOUT`, `OPENAI_MAX_TOKENS`, `OPENAI_TEMPERATURE`, `CHATGPT_CLI_CONFIG_DIR`

**Examples:**

```bash
chatgpt-cli config get OPENAI_MODEL
# Output: gpt-3.5-turbo

chatgpt-cli config get openai_max_tokens
# Output: 1000

# Use in a script
MODEL=$(chatgpt-cli config get OPENAI_MODEL)
echo "Current model: $MODEL"
```

### `config set`

Sets a configuration value and persists it to the config file.

```bash
chatgpt-cli config set <key> <value>
```

**Settable keys and validation rules:**

| Key | Validation |
|-----|------------|
| `OPENAI_API_KEY` | Cannot be empty |
| `OPENAI_API_URL` | Must start with `http://` or `https://` |
| `OPENAI_MODEL` | Cannot be empty |
| `OPENAI_TIMEOUT` | Must be a valid Go duration (e.g., `60s`, `1m`, `90s`) |
| `OPENAI_MAX_TOKENS` | Must be a positive integer |
| `OPENAI_TEMPERATURE` | Must be a number between `0.0` and `2.0` |

!!! note
    `CHATGPT_CLI_CONFIG_DIR` cannot be set via `config set`. Use the environment variable instead.

**Examples:**

```bash
# Switch to GPT-4
chatgpt-cli config set OPENAI_MODEL gpt-4

# Increase max tokens
chatgpt-cli config set OPENAI_MAX_TOKENS 2000

# Set a longer timeout
chatgpt-cli config set OPENAI_TIMEOUT 120s

# Adjust creativity
chatgpt-cli config set OPENAI_TEMPERATURE 1.5

# Set a custom API endpoint
chatgpt-cli config set OPENAI_API_URL https://my-proxy.example.com/v1/chat/completions
```

**Successful output:**

```
Set OPENAI_MODEL=gpt-4
Configuration saved to /home/user/.chatgpt-cli/config
```

---

## Unknown Commands

If you provide an unrecognized command, the CLI prints an error message and displays the help text:

```bash
$ chatgpt-cli foo
Unknown command: foo

ChatGPT CLI - Command Line Interface for ChatGPT
...
```

For configuration details, see the [Configuration](configuration.md) page.
