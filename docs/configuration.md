# Configuration

ChatGPT CLI supports configuration through environment variables and a configuration file. Environment variables take precedence over the config file, and the config file takes precedence over built-in defaults.

## Configuration Precedence

```
Environment Variables  →  Config File  →  Default Values
   (highest priority)                     (lowest priority)
```

1. **Environment variables** — always checked first
2. **Config file** — checked if the environment variable is not set
3. **Default values** — used if neither the environment variable nor the config file provides a value

---

## Configuration Variables

| Variable | Description | Type | Default | Required |
|----------|-------------|------|---------|----------|
| `OPENAI_API_KEY` | Your OpenAI API key | `string` | *(none)* | **Yes** |
| `OPENAI_API_URL` | API endpoint URL | `string` | `https://api.openai.com/v1/chat/completions` | No |
| `OPENAI_MODEL` | Model to use for completions | `string` | `gpt-3.5-turbo` | No |
| `OPENAI_TIMEOUT` | HTTP request timeout | `duration` | `60s` (1 minute) | No |
| `OPENAI_MAX_TOKENS` | Maximum tokens in the response | `integer` | `1000` | No |
| `OPENAI_TEMPERATURE` | Response randomness (0.0–2.0) | `float` | `0.7` | No |
| `CHATGPT_CLI_CONFIG_DIR` | Path to the configuration directory | `string` | `~/.chatgpt-cli` | No |

### Variable Details

#### `OPENAI_API_KEY`

Your OpenAI API key, used to authenticate requests. You can obtain one from [platform.openai.com/api-keys](https://platform.openai.com/api-keys).

- **Required:** Yes — the `prompt` command will fail without it.
- **Security:** The key is masked in `config list` and `config get` output (shows first 4 and last 4 characters only).

#### `OPENAI_API_URL`

The OpenAI-compatible API endpoint URL. Change this to use a proxy server or an alternative API provider.

- **Validation:** Must start with `http://` or `https://`.

#### `OPENAI_MODEL`

The model to use for chat completions. Common values include `gpt-3.5-turbo`, `gpt-4`, and `gpt-4-turbo`.

#### `OPENAI_TIMEOUT`

The HTTP request timeout as a Go duration string. Controls how long the CLI waits for a response before timing out.

- **Format:** Go duration syntax — e.g., `30s`, `1m`, `90s`, `2m30s`.
- **Tip:** Increase this for complex prompts that require longer processing.

#### `OPENAI_MAX_TOKENS`

The maximum number of tokens in the API response. Higher values allow longer responses but consume more API credits.

- **Validation:** Must be a positive integer.

#### `OPENAI_TEMPERATURE`

Controls the randomness/creativity of the response. Lower values produce more deterministic output; higher values produce more varied output.

- **Range:** `0.0` to `2.0`
- **Tip:** Use `0.0`–`0.3` for factual/deterministic tasks; `0.7`–`1.0` for creative tasks.

#### `CHATGPT_CLI_CONFIG_DIR`

The directory where the configuration file and logs are stored.

- **Default:** `~/.chatgpt-cli` (i.e., `$HOME/.chatgpt-cli`)
- **Note:** This can only be set via the environment variable — it cannot be changed with `config set`.

---

## Config File

### Location

The config file is located at `<config_dir>/config`, which defaults to:

```
~/.chatgpt-cli/config
```

The directory is created automatically on first run if it does not exist.

### Format

The config file uses a simple `KEY=VALUE` format, one entry per line. Lines starting with `#` are treated as comments, and blank lines are ignored.

```ini
# ChatGPT CLI Configuration
# Generated on 2024-01-31 14:30:05

OPENAI_API_KEY=sk-your-api-key-here
OPENAI_API_URL=https://api.openai.com/v1/chat/completions
OPENAI_MODEL=gpt-4
OPENAI_TIMEOUT=90s
OPENAI_MAX_TOKENS=2000
OPENAI_TEMPERATURE=0.7
```

### Managing the Config File

The config file is managed through the `config set` command:

```bash
# Set a value (creates the file if it doesn't exist)
chatgpt-cli config set OPENAI_MODEL gpt-4

# Set multiple values
chatgpt-cli config set OPENAI_MAX_TOKENS 2000
chatgpt-cli config set OPENAI_TEMPERATURE 1.0
```

Each `config set` call merges the new value into the existing config file, preserving all other values. The config file is written with permissions `0600` (owner read/write only) for security.

You can also view the current configuration:

```bash
# List all configuration
chatgpt-cli config list

# Get a specific value
chatgpt-cli config get OPENAI_MODEL
```

---

## Setting Configuration via Environment Variables

### For the Current Session

```bash
export OPENAI_API_KEY="sk-your-api-key-here"
export OPENAI_MODEL="gpt-4"
export OPENAI_MAX_TOKENS="2000"
```

### Permanently (Shell Profile)

Add the exports to your shell profile (`~/.bashrc`, `~/.zshrc`, or `~/.profile`):

```bash
echo 'export OPENAI_API_KEY="sk-your-api-key-here"' >> ~/.bashrc
echo 'export OPENAI_MODEL="gpt-4"' >> ~/.bashrc
source ~/.bashrc
```

### Per-Command Override

You can override any variable for a single command invocation:

```bash
OPENAI_MODEL=gpt-4 chatgpt-cli prompt "Explain quantum computing"
OPENAI_TIMEOUT=120s chatgpt-cli prompt "Write a complete REST API server"
```

---

## File Locations Summary

| File | Path | Description |
|------|------|-------------|
| Config file | `~/.chatgpt-cli/config` | Persisted configuration values |
| Log file | `~/.chatgpt-cli/logs.jsonl` | Application logs in JSONL format |

Both paths are relative to the config directory, which can be changed with `CHATGPT_CLI_CONFIG_DIR`.

---

## Example Configuration

### Minimal Setup

```bash
# Just set your API key — all other values use defaults
export OPENAI_API_KEY="sk-your-api-key-here"
```

### Full Config File Example

```ini
# ChatGPT CLI Configuration

OPENAI_API_KEY=sk-your-api-key-here
OPENAI_API_URL=https://api.openai.com/v1/chat/completions
OPENAI_MODEL=gpt-4
OPENAI_TIMEOUT=90s
OPENAI_MAX_TOKENS=2000
OPENAI_TEMPERATURE=0.5
```

### Using a Custom API Endpoint

```ini
# Use a local proxy or alternative provider
OPENAI_API_KEY=my-custom-key
OPENAI_API_URL=https://my-proxy.example.com/v1/chat/completions
OPENAI_MODEL=my-custom-model
```

For command usage details, see the [Usage](usage.md) page.
