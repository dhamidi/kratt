# Worker Implementation

## Specification

The Worker implements an automated pull request processing system that:

1. Retrieves pull request information using GitHub CLI
2. Creates or switches to a git worktree for the PR branch
3. Generates a prompt combining PR data with custom instructions
4. Runs an AI agent with the prompt as input
5. Executes lint and test commands after agent completion
6. Posts results as a PR comment
7. Commits and pushes any changes made by the agent

## Implementation Steps

### Step 1: Define Core Interfaces

Create three interfaces to encapsulate external dependencies:

#### LocalGit Interface
```go
type LocalGit interface {
    // CheckWorktreeExists checks if a worktree exists for the given branch
    CheckWorktreeExists(branch string) (bool, error)
    
    // CreateWorktree creates a new worktree for the given branch at the specified path
    CreateWorktree(branch, path string) error
    
    // ChangeDirectory changes to the specified worktree directory
    ChangeDirectory(path string) error
    
    // CommitAndPush commits all changes and pushes to the remote branch
    CommitAndPush(message string) error
    
    // GetWorktreePath returns the path to the worktree for the given branch
    GetWorktreePath(branch string) (string, error)
}
```

#### GitHub Interface
```go
type GitHub interface {
    // GetPRInfo retrieves pull request information including comments
    GetPRInfo(prNumber int) (string, error)
    
    // PostComment posts a comment to the specified pull request
    PostComment(prNumber int, body string) error
}
```

#### CommandRunner Interface
```go
type CommandRunner interface {
    // RunWithStdin executes a command with the given stdin input
    RunWithStdin(ctx context.Context, stdin string, command string, args ...string) error
    
    // RunWithOutput executes a command and returns stdout and stderr
    RunWithOutput(ctx context.Context, command string, args ...string) (stdout, stderr string, err error)
}
```

### Step 2: Define Worker Struct

```go
type Worker struct {
    Instructions string   // Prefix for the agent prompt
    AgentCommand []string // Command to run the AI agent
    LintCommand  []string // Command to run linting
    TestCommand  []string // Command to run tests
    Deadline     time.Duration // Maximum time for agent execution
    
    // Dependencies (injected for testability)
    Git    LocalGit
    GitHub GitHub  
    Runner CommandRunner
}
```

### Step 3: Implement Worker Method

Create the main `ProcessPR(prNumber int) error` method:

#### 3.1: Get PR Information
- Call `w.GitHub.GetPRInfo(prNumber)` to retrieve PR data
- Handle any errors returned from GitHub API

#### 3.2: Handle Git Worktree
- Extract branch name from PR information
- Call `w.Git.CheckWorktreeExists(branch)` to check if worktree exists
- If worktree doesn't exist:
  - Call `w.Git.CreateWorktree(branch, path)` to create it
- Call `w.Git.ChangeDirectory(path)` to switch to worktree

#### 3.3: Generate Agent Prompt
- Wrap PR info in `<pull-request>...</pull-request>` tags
- Prefix with `w.Instructions`
- Create final prompt string

#### 3.4: Execute Agent with Timeout
- Create context with timeout using `w.Deadline`
- Call `w.Runner.RunWithStdin(ctx, prompt, w.AgentCommand[0], w.AgentCommand[1:]...)`
- Handle timeout/cancellation gracefully

#### 3.5: Run Lint and Test Commands
- Call `w.Runner.RunWithOutput(ctx, w.LintCommand[0], w.LintCommand[1:]...)` 
- Call `w.Runner.RunWithOutput(ctx, w.TestCommand[0], w.TestCommand[1:]...)`
- Collect both stdout and stderr from each command
- Handle any execution errors

#### 3.6: Post Results Comment
- Format lint and test outputs into a comment body
- Include both stdout and stderr
- Add success/failure indicators
- Call `w.GitHub.PostComment(prNumber, commentBody)`

#### 3.7: Commit and Push Changes
- Call `w.Git.CommitAndPush("Automated changes from kratt worker")`
- Handle any git operation errors

### Step 4: Implement Concrete Types

#### GitRunner (implements LocalGit)
- Use `os/exec` to run git commands
- Handle worktree operations using `git worktree` subcommands
- Implement directory changes using `os.Chdir` or command working directory

#### GitHubCLI (implements GitHub)  
- Use `os/exec` to run `gh` commands
- Parse `gh pr view` output for PR information
- Use `gh pr comment` for posting comments

#### ExecRunner (implements CommandRunner)
- Use `os/exec.CommandContext` for timeout support
- Handle stdin/stdout/stderr piping
- Return captured output and errors

### Step 5: Error Handling Strategy

- Wrap all errors with context using `fmt.Errorf`
- Use specific error types for different failure modes
- Ensure cleanup of resources (worktrees, processes)
- Log important operations for debugging

### Step 6: Testing Approach

- Create mock implementations of all interfaces
- Test each step of the workflow independently
- Include integration tests with real git repositories
- Test timeout and cancellation behavior
- Verify error handling and cleanup

## File Structure

```
worker/
├── worker.go          # Main Worker struct and ProcessPR method
├── interfaces.go      # Interface definitions
├── git.go            # GitRunner implementation
├── github.go         # GitHubCLI implementation  
├── exec.go           # ExecRunner implementation
└── worker_test.go    # Unit and integration tests
```

## Usage Example

```go
worker := &Worker{
    Instructions: "You are an AI assistant helping with code review.",
    AgentCommand: []string{"amp", "--stdin"},
    LintCommand:  []string{"golangci-lint", "run"},
    TestCommand:  []string{"go", "test", "./..."},
    Deadline:     5 * time.Minute,
    Git:          &GitRunner{},
    GitHub:       &GitHubCLI{},
    Runner:       &ExecRunner{},
}

err := worker.ProcessPR(123)
if err != nil {
    log.Fatal(err)
}
```
