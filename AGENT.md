# AGENT.md

This file contains guidelines for AI coding agents working on the Kratt project.

## Build/Test Commands
- Build: `go build ./...`
- Test all: `go test ./...`
- Test single: `go test ./path/to/package -run TestName`
- Lint: `golangci-lint run`
- Format: `go fmt ./...`

## Code Style Guidelines
- Language: Go 1.24 (module: github.com/dhamidi/kratt)
- Formatting: Use `gofmt`/`go fmt`
- Imports: Group stdlib, external, local with blank lines between groups
- Naming: Follow Go conventions (PascalCase for exported, camelCase for unexported)
- Error handling: Return errors explicitly, wrap with context using `fmt.Errorf`
- Types: Use `any` instead of `interface{}`
- Comments: Follow Go doc comment conventions, start with function/type name
- Packages: Use short, lowercase names without underscores

## Project Structure
- `docs/` - Documentation
- `worker/` - Worker implementation (empty)

Note: This file should be updated as the project develops and conventions are established.
