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

### `kratt worker start <branch-name> <instructions>`

Creates a new branch with instructions and opens a pull request for implementation.

**Usage:**

```bash
kratt worker start feature/auth "Implement JWT authentication"
kratt worker start bugfix/login-error "Fix login validation bug"
```

**Behavior:**

1. Detects the current git repository and validates it's a valid git repo
2. Determines the GitHub repository from git remotes
3. Creates a new branch with the specified name
4. Writes instructions to `docs/<branch-name>-instructions.md`
5. Commits the instructions file
6. Pushes the branch upstream with `git push -u origin <branch-name>`
7. Creates a pull request with title "Implement <branch-name>" and description asking for implementation steps
8. Exits with status 0 on success, 1 on error

**Error Conditions:**

- Not in a git repository: "Error: current directory is not a git repository"
- No GitHub remote found: "Error: no GitHub remote found in current repository"
- Invalid branch name: "Error: invalid branch name: must not contain invalid characters"
- Branch already exists: "Error: branch already exists: <branch-name>"
- GitHub API errors: "Error: failed to create PR: <details>"
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
kratt worker start feature/auth "Implement auth" --timeout 45m
```

## Implementation Structure

```
cmd/
├── root.go          # Root command setup and global flags
├── worker.go        # Worker subcommand group
├── worker_run.go    # worker run subcommand implementation
└── worker_start.go  # worker start subcommand implementation

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

## Implementation Status

**CLI Implementation: DONE ✅**

- ✅ Root command setup with global flags (cmd/root.go)
- ✅ Worker subcommand group (cmd/worker.go)  
- ✅ Worker run subcommand implementation (cmd/worker_run.go)
- ✅ Main CLI entry point (main.go)
- ✅ Repository detection and validation
- ✅ Configuration flag handling
- ✅ Error handling with proper exit codes
- ✅ Integration with worker package

**Worker Start Command: DONE ✅**

- ✅ Worker start subcommand implementation (cmd/worker_start.go)
- ✅ Branch name validation logic
- ✅ Integration with Worker.Start method
- ✅ Error handling for branch creation and PR creation
- ✅ Tests for start command functionality
