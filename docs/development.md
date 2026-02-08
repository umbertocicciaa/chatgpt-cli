# Development

## Prerequisites

- Go 1.19 or later
- Make (optional, for using Makefile targets)

## Getting Started

```bash
git clone https://github.com/umbertocicciaa/chatgpt-cli.git
cd chatgpt-cli
go mod download
```

## Building

```bash
go build -v ./...
```

## Testing

```bash
go test -v ./...
```

## Linting

```bash
golangci-lint run
```

## Project Structure

```
chatgpt-cli/
├── main.go          # Application entry point
├── main_test.go     # Tests
├── go.mod           # Go module definition
├── Makefile         # Build automation
├── docs/            # Documentation (MkDocs)
└── mkdocs.yml       # MkDocs configuration
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -am 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request
