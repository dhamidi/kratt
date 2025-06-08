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

### Step 1: Define Core Interfaces - DONE ✅

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
    
    // Repository detection methods (added for CLI support)
    // IsGitRepository checks if the current directory is a git repository
    IsGitRepository() (bool, error)
    
    // GetGitHubRepository extracts GitHub owner/repo from git remotes
    GetGitHubRepository() (owner, repo string, err error)
    
    // Start method support (added for Worker.Start)
    // CreateBranch creates a new branch and switches to it
    CreateBranch(branchName string) error
    
    // WriteFile writes content to a file at the specified path
    WriteFile(path, content string) error
    
    // PushBranchUpstream pushes a new branch upstream with git push -u origin
    PushBranchUpstream(branchName string) error
}
```

#### GitHub Interface

```go
type GitHub interface {
    // GetPRInfo retrieves pull request information including comments
    GetPRInfo(prNumber int) (string, error)
    
    // PostComment posts a comment to the specified pull request
    PostComment(prNumber int, body string) error
    
    // CreatePR creates a new pull request with the given title and description
    CreatePR(title, description string) error
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

### Step 2: Define Worker Struct - DONE ✅

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

#### Worker Methods

The Worker provides two main methods:

1. **ProcessPR(prNumber int) error** - Processes an existing pull request
2. **Start(branchName string, instruction string) error** - Creates a new branch and pull request with instructions

### Step 3: Implement Worker Method - DONE ✅

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

### Step 8: Implement Worker.Start Method - NEW

Create the `Start(branchName string, instruction string) error` method:

#### 8.1: Create and Switch to New Branch

- Call `w.Git.CreateBranch(branchName)` to create a new branch and switch to it
- Handle any git operation errors

#### 8.2: Write Instructions File

- Create file path as `docs/<branchName>-instructions.md`
- Call `w.Git.WriteFile(path, instruction)` to write the instructions
- Handle any file writing errors

#### 8.3: Commit Instructions File

- Call `w.Git.CommitAndPush("Add instructions for " + branchName)`
- Handle any git operation errors

#### 8.4: Push Branch Upstream

- Call `w.Git.PushBranchUpstream(branchName)` to push with `git push -u origin`
- Handle any git operation errors

#### 8.5: Create Pull Request

- Create PR title as "Implement " + branchName
- Create PR description as "Study docs/<branchName>-instructions.md and make a list of necessary implementation steps in docs/<branchName>-implementation-status.md"
- Call `w.GitHub.CreatePR(title, description)` to create the pull request
- Handle any GitHub API errors

### Step 4: Implement Concrete Types - DONE ✅

#### GitRunner (implements LocalGit)

- Use `os/exec` to run git commands
- Handle worktree operations using `git worktree` subcommands
- Implement directory changes using `os.Chdir` or command working directory
- Repository detection using `git rev-parse --is-inside-work-tree`
- GitHub repository extraction using `git remote get-url origin` and URL parsing
- Branch creation using `git checkout -b <branchName>`
- File writing using `os.WriteFile` or equivalent
- Upstream push using `git push -u origin <branchName>`

#### GitHubCLI (implements GitHub)  

- Use `os/exec` to run `gh` commands
- Parse `gh pr view` output for PR information
- Use `gh pr comment` for posting comments
- Use `gh pr create --title <title> --body <description>` for creating pull requests

#### ExecRunner (implements CommandRunner)

- Use `os/exec.CommandContext` for timeout support
- Handle stdin/stdout/stderr piping
- Use `exec.Cmd.CombinedOutput()` to get interleaved stdout/stderr
- Return captured output as []byte and errors

### Step 5: Error Handling Strategy - DONE ✅

- Wrap all errors with context using `fmt.Errorf`
- Use specific error types for different failure modes
- Ensure cleanup of resources (worktrees, processes)
- Log important operations for debugging

### Step 6: Testing Approach - DONE ✅

Use fakes (not mocks) that maintain internal state and uphold invariants:

#### FakeLocalGit

- Maintains a map of existing worktrees
- `CreateWorktree()` adds to the worktrees map
- `CheckWorktreeExists()` checks the worktrees map
- `ChangeDirectory()` tracks current directory
- `CommitAndPush()` records commits made
- `IsGitRepository()` returns configurable boolean (default: true)
- `GetGitHubRepository()` returns configurable owner/repo (default: "owner/repo")
- `CreateBranch()` records created branches
- `WriteFile()` stores file content in memory
- `PushBranchUpstream()` records upstream pushes
- All operations respect the internal state

#### FakeGitHub  

- Stores PR data and comments in memory
- `GetPRInfo()` returns stored PR information
- `PostComment()` adds comments to internal storage
- `CreatePR()` records created pull requests with title and description
- Allows verification of posted comments and created PRs

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

## File Structure - COMPLETED ✅

```
worker/
├── worker.go          # Main Worker struct and ProcessPR method - DONE ✅
├── git.go            # LocalGit interface and GitRunner and fake implementation - DONE ✅
├── github.go         # Github interface and GitHubCLI and fake implementation - DONE ✅
├── exec.go           # CommandRunner interface and ExecRunner and fake implementation - DONE ✅
└── worker_test.go    # Unit and integration tests - DONE ✅
```

## Implementation Status

Core worker implementation completed:
- ✅ All interfaces defined with proper method signatures
- ✅ Concrete implementations using exec commands for git, GitHub CLI, and system commands
- ✅ Comprehensive fake implementations for testing
- ✅ Full Worker.ProcessPR implementation with all 7 steps
- ✅ Complete test coverage with unit and integration tests
- ✅ Error handling with proper context wrapping
- ✅ All tests passing

**CLI Integration Requirements:**
- ✅ Add repository detection methods to LocalGit interface - DONE
- ✅ Implement IsGitRepository() and GetGitHubRepository() in GitRunner - DONE
- ✅ Update FakeLocalGit with new methods for testing - DONE

**Worker.Start Method Requirements:**
- ⏳ Extend LocalGit interface with CreateBranch, WriteFile, PushBranchUpstream methods
- ⏳ Extend GitHub interface with CreatePR method
- ⏳ Implement Worker.Start method with 5 steps (8.1-8.5)
- ⏳ Update GitRunner with new LocalGit methods
- ⏳ Update GitHubCLI with CreatePR method
- ⏳ Update FakeLocalGit and FakeGitHub with new methods
- ⏳ Add comprehensive tests for Worker.Start method

## Usage Examples

### Processing an Existing PR

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

### Creating a New Branch and PR

```go
worker := &Worker{
    // ... same configuration as above
}

instruction := "Implement user authentication system with JWT tokens and role-based access control"
err := worker.Start("feature/auth-system", instruction)
if err != nil {
    log.Fatal(err)
}
```
