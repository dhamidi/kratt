# Kratt CLI Implementation

## Overview

The `kratt` CLI provides a command-line interface for running the automated PR processing worker. It uses `github.com/spf13/cobra` for command parsing and structure.

## Commands

### `kratt worker run <pr-number>`

Runs the worker to process a specific pull request in the current repository.

**Usage:**
```bash
kratt worker run 1        # Process PR #1
kratt worker run 42       # Process PR #42
```

**Behavior:**
1. Detects the current git repository and validates it's a valid git repo
2. Determines the GitHub repository from git remotes
3. Configures and runs the Worker with default settings
4. Processes the specified pull request number
5. Exits with status 0 on success, 1 on error

**Error Conditions:**
- Not in a git repository: "Error: current directory is not a git repository"
- No GitHub remote found: "Error: no GitHub remote found in current repository"
- Invalid PR number: "Error: invalid pull request number: must be a positive integer"
- GitHub API errors: "Error: failed to access PR #X: <details>"
- Git operation errors: "Error: git operation failed: <details>"

## Configuration

The CLI uses default configuration that can be customized via flags:

### Global Flags
- `--timeout duration`: Maximum time for agent execution (default: 30m)
- `--instructions file`: Path to file containing agent instructions (default: built-in instructions)
- `--agent command`: Command to run the AI agent (default: ["amp", "--stdin"])
- `--lint command`: Command to run linting (default: ["go", "fmt", "./..."])
- `--test command`: Command to run tests (default: ["go", "test", "./..."])

### Example with Flags
```bash
kratt worker run 1 --timeout 45m --instructions ./custom-instructions.txt
kratt worker run 1 --agent "claude-dev --stdin" --lint "golangci-lint run"
```

## Implementation Structure

```
cmd/
├── root.go          # Root command setup and global flags
├── worker.go        # Worker subcommand group
└── worker_run.go    # worker run subcommand implementation

main.go              # CLI entry point
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management (optional)
- Existing `worker` package

## Repository Detection

The CLI must detect:
1. Current directory is a git repository
2. GitHub repository from git remotes (origin or upstream)
3. Repository owner and name for GitHub API calls

This requires extending the `LocalGit` interface with repository detection methods.

## Default Configuration

```go
defaultWorker := &worker.Worker{
    Instructions: "You are an AI assistant helping with code review. Please analyze the pull request and make any necessary improvements to the code.",
    AgentCommand: []string{"amp", "--stdin"},
    LintCommand:  []string{"go", "fmt", "./..."},
    TestCommand:  []string{"go", "test", "./..."},
    Deadline:     30 * time.Minute,
    Git:          &worker.GitRunner{},
    GitHub:       &worker.GitHubCLI{},
    Runner:       &worker.ExecRunner{},
}
```

## Error Handling

- All errors include context about the operation that failed
- Exit codes: 0 = success, 1 = error
- Error messages are user-friendly and actionable
- Debug information available via `--verbose` flag

## Future Extensions

- `kratt config` - Manage configuration
- `kratt worker status` - Check worker status
- `kratt worker list` - List recent worker runs
- Configuration file support in `.kratt.yaml`
