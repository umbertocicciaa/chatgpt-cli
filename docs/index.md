# ChatGPT CLI

A modern, extensible command-line interface for ChatGPT with subcommand support, configuration management, and comprehensive logging.

## Features

- Subcommand-based architecture (`prompt`, `logs`, `config`, `help`)
- Configuration management via environment variables and config file
- Comprehensive logging of prompts, responses, and errors
- API key masking for security
- Cross-platform support (Linux, macOS, Windows)

## Quick Start

```bash
# 1. Set your API key
export OPENAI_API_KEY="sk-your-api-key-here"

# 2. Build
go build -o chatgpt-cli main.go

# 3. Send a prompt
chatgpt-cli prompt "Explain Go interfaces"
```

## Documentation

- [Installation](installation.md) - How to install ChatGPT CLI
- [Usage](usage.md) - All commands, syntax, and examples
- [Configuration](configuration.md) - Configuration variables, config file, and precedence
- [Development](development.md) - Contributing and development guide
