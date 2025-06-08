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

### Step 1: Define Core Interfaces - DONE

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
    
    // RunWithOutput executes a command and returns interleaved stdout/stderr output
    RunWithOutput(ctx context.Context, command string, args ...string) (output []byte, err error)
}
```

### Step 2: Define Worker Struct - DONE

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

### Step 3: Implement Worker Method - DONE

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
- Collect interleaved output from each command
- Handle any execution errors

#### 3.6: Post Results Comment

- Format lint and test outputs into a comment body
- Convert []byte output to string for display
- Add success/failure indicators
- Call `w.GitHub.PostComment(prNumber, commentBody)`

#### 3.7: Commit and Push Changes

- Call `w.Git.CommitAndPush("Automated changes from kratt worker")`
- Handle any git operation errors

### Step 4: Implement Concrete Types - DONE

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
- Use `exec.Cmd.CombinedOutput()` to get interleaved stdout/stderr
- Return captured output as []byte and errors

### Step 5: Error Handling Strategy - DONE

- Wrap all errors with context using `fmt.Errorf`
- Use specific error types for different failure modes
- Ensure cleanup of resources (worktrees, processes)
- Log important operations for debugging

### Step 6: Testing Approach - DONE

Use fakes (not mocks) that maintain internal state and uphold invariants:

#### FakeLocalGit

- Maintains a map of existing worktrees
- `CreateWorktree()` adds to the worktrees map
- `CheckWorktreeExists()` checks the worktrees map
- `ChangeDirectory()` tracks current directory
- `CommitAndPush()` records commits made
- All operations respect the internal state

#### FakeGitHub  

- Stores PR data and comments in memory
- `GetPRInfo()` returns stored PR information
- `PostComment()` adds comments to internal storage
- Allows verification of posted comments

#### FakeCommandRunner

- Maps command patterns to predefined responses
- `RunWithStdin()` records stdin input for verification
- `RunWithOutput()` returns configured []byte responses
- Simulates command execution without actual process spawning
- Can simulate timeouts and errors

#### Test Strategy

- Test each step of the workflow independently using fakes
- Verify state changes in fakes (worktrees created, comments posted)
- Test timeout and cancellation behavior with controlled fakes
- Verify error handling and cleanup with failing fake operations
- Integration tests use fakes to simulate complete workflows

## File Structure

```
worker/
├── worker.go          # Main Worker struct and ProcessPR method
├── git.go            # LocalGit interface and GitRunner and fake implementation
├── github.go         # Github interface and GitHubCLI and fake implementation  
├── exec.go           # CommandRunner interface and ExecRunner and fake implementation
└── worker_test.go    # Unit and integration tests
```

## Usage Example

```go
worker := &Worker{
    Instructions: "You are an AI assistant helping with code review.",
    AgentCommand: []string{"amp", "--stdin"},
    LintCommand:  []string{"goimports", "-w", "./..."},
    TestCommand:  []string{"go", "test", "./..."},
    Deadline:     30 * time.Minute,
    Git:          &GitRunner{},
    GitHub:       &GitHubCLI{},
    Runner:       &ExecRunner{},
}

err := worker.ProcessPR(123)
if err != nil {
    log.Fatal(err)
}
```
