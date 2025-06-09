package worker

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LocalGit interface encapsulates git worktree operations
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

	// BranchExists checks if a branch exists
	BranchExists(branchName string) (bool, error)
}

// GitRunner implements LocalGit interface using git commands
type GitRunner struct{}

// CheckWorktreeExists checks if a worktree exists for the given branch
func (g *GitRunner) CheckWorktreeExists(branch string) (bool, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list worktrees: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "branch ") && strings.Contains(line, branch) {
			return true, nil
		}
	}
	return false, nil
}

// CreateWorktree creates a new worktree for the given branch at the specified path
func (g *GitRunner) CreateWorktree(branch, path string) error {
	cmd := exec.Command("git", "worktree", "add", path, branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create worktree for branch %s at %s: %w", branch, path, err)
	}
	return nil
}

// ChangeDirectory changes to the specified worktree directory
func (g *GitRunner) ChangeDirectory(path string) error {
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", path, err)
	}
	return nil
}

// CommitAndPush commits all changes and pushes to the remote branch
func (g *GitRunner) CommitAndPush(message string) error {
	// Add all changes
	addCmd := exec.Command("git", "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Check if there are any changes to commit
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	// If no changes, skip commit and push
	if len(strings.TrimSpace(string(statusOutput))) == 0 {
		return nil
	}

	// Commit changes
	commitCmd := exec.Command("git", "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	// Get current branch name
	branchCmd := exec.Command("git", "branch", "--show-current")
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branchName := strings.TrimSpace(string(branchOutput))

	// Push changes with upstream
	pushCmd := exec.Command("git", "push", "-u", "origin", branchName)
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}

// GetWorktreePath returns the path to the worktree for the given branch
func (g *GitRunner) GetWorktreePath(branch string) (string, error) {
	// Get the current repository root
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}

	repoRoot := strings.TrimSpace(string(output))
	// Create worktree path as ../repo-branch
	repoName := filepath.Base(repoRoot)
	worktreePath := filepath.Join(filepath.Dir(repoRoot), fmt.Sprintf("%s-%s", repoName, branch))

	return worktreePath, nil
}

// IsGitRepository checks if the current directory is a git repository
func (g *GitRunner) IsGitRepository() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	if err != nil {
		// If git command fails, we're not in a git repository
		return false, nil
	}
	return true, nil
}

// GetGitHubRepository extracts GitHub owner/repo from git remotes
func (g *GitRunner) GetGitHubRepository() (owner, repo string, err error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get remote origin URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(remoteURL, "git@github.com:") {
		path := strings.TrimPrefix(remoteURL, "git@github.com:")
		return parseGitHubPath(path)
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse remote URL: %w", err)
	}

	if parsedURL.Host != "github.com" {
		return "", "", fmt.Errorf("not a GitHub repository: %s", remoteURL)
	}

	return parseGitHubPath(parsedURL.Path)
}

// parseGitHubPath extracts owner and repo from a GitHub path
func parseGitHubPath(path string) (owner, repo string, err error) {
	// Remove leading slash and .git suffix
	path = strings.TrimPrefix(path, "/")
	if strings.HasSuffix(path, ".git") {
		path = path[:len(path)-4]
	}

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub path format: %s", path)
	}

	return parts[0], parts[1], nil
}

// CreateBranch creates a new branch and switches to it
func (g *GitRunner) CreateBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create and switch to branch %s: %w", branchName, err)
	}
	return nil
}

// WriteFile writes content to a file at the specified path
func (g *GitRunner) WriteFile(path, content string) error {
	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// PushBranchUpstream pushes a new branch upstream with git push -u origin
func (g *GitRunner) PushBranchUpstream(branchName string) error {
	cmd := exec.Command("git", "push", "-u", "origin", branchName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %s upstream: %w", branchName, err)
	}
	return nil
}

// BranchExists checks if a branch exists
func (g *GitRunner) BranchExists(branchName string) (bool, error) {
	cmd := exec.Command("git", "branch", "--list", branchName)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list branches: %w", err)
	}

	// If output is not empty, the branch exists
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// FakeLocalGit implements LocalGit interface for testing
type FakeLocalGit struct {
	worktrees       map[string]string // branch -> path mapping
	currentDir      string
	commits         []string
	isGitRepo       bool
	githubOwner     string
	githubRepo      string
	createdBranches []string          // track created branches
	writtenFiles    map[string]string // path -> content mapping
	pushedBranches  []string          // track pushed branches

	// Error simulation flags
	FailCreateBranch        bool
	FailWriteFile           bool
	FailCommitAndPush       bool
	FailPushBranchUpstream  bool
	FailGetGitHubRepository bool
}

// NewFakeLocalGit creates a new FakeLocalGit instance
func NewFakeLocalGit() *FakeLocalGit {
	return &FakeLocalGit{
		worktrees:       make(map[string]string),
		currentDir:      "/fake/repo",
		commits:         []string{},
		isGitRepo:       true,
		githubOwner:     "owner",
		githubRepo:      "repo",
		createdBranches: []string{},
		writtenFiles:    make(map[string]string),
		pushedBranches:  []string{},
	}
}

// CheckWorktreeExists checks if a worktree exists in the fake state
func (f *FakeLocalGit) CheckWorktreeExists(branch string) (bool, error) {
	_, exists := f.worktrees[branch]
	return exists, nil
}

// CreateWorktree adds a worktree to the fake state
func (f *FakeLocalGit) CreateWorktree(branch, path string) error {
	f.worktrees[branch] = path
	return nil
}

// ChangeDirectory updates the fake current directory
func (f *FakeLocalGit) ChangeDirectory(path string) error {
	f.currentDir = path
	return nil
}

// CommitAndPush records a commit in the fake state
func (f *FakeLocalGit) CommitAndPush(message string) error {
	if f.FailCommitAndPush {
		return fmt.Errorf("fake commit and push failure")
	}
	f.commits = append(f.commits, message)
	return nil
}

// GetWorktreePath returns the path for a branch or generates one
func (f *FakeLocalGit) GetWorktreePath(branch string) (string, error) {
	if path, exists := f.worktrees[branch]; exists {
		return path, nil
	}
	return fmt.Sprintf("/fake/repo-%s", branch), nil
}

// GetCommits returns all recorded commits (for testing)
func (f *FakeLocalGit) GetCommits() []string {
	return f.commits
}

// GetCurrentDir returns the current directory (for testing)
func (f *FakeLocalGit) GetCurrentDir() string {
	return f.currentDir
}

// IsGitRepository returns the configured git repository status (for testing)
func (f *FakeLocalGit) IsGitRepository() (bool, error) {
	return f.isGitRepo, nil
}

// GetGitHubRepository returns the configured GitHub owner/repo (for testing)
func (f *FakeLocalGit) GetGitHubRepository() (owner, repo string, err error) {
	if f.FailGetGitHubRepository {
		return "", "", fmt.Errorf("fake get github repository failure")
	}
	return f.githubOwner, f.githubRepo, nil
}

// SetGitRepository sets whether this should report as a git repository (for testing)
func (f *FakeLocalGit) SetGitRepository(isRepo bool) {
	f.isGitRepo = isRepo
}

// SetGitHubRepository sets the GitHub owner/repo values (for testing)
func (f *FakeLocalGit) SetGitHubRepository(owner, repo string) {
	f.githubOwner = owner
	f.githubRepo = repo
}

// CreateBranch records a created branch in the fake state
func (f *FakeLocalGit) CreateBranch(branchName string) error {
	if f.FailCreateBranch {
		return fmt.Errorf("fake create branch failure")
	}
	f.createdBranches = append(f.createdBranches, branchName)
	return nil
}

// WriteFile stores file content in the fake state
func (f *FakeLocalGit) WriteFile(path, content string) error {
	if f.FailWriteFile {
		return fmt.Errorf("fake write file failure")
	}
	f.writtenFiles[path] = content
	return nil
}

// PushBranchUpstream records a pushed branch in the fake state
func (f *FakeLocalGit) PushBranchUpstream(branchName string) error {
	if f.FailPushBranchUpstream {
		return fmt.Errorf("fake push branch upstream failure")
	}
	f.pushedBranches = append(f.pushedBranches, branchName)
	return nil
}

// GetCreatedBranches returns all created branches (for testing)
func (f *FakeLocalGit) GetCreatedBranches() []string {
	return f.createdBranches
}

// GetWrittenFiles returns all written files (for testing)
func (f *FakeLocalGit) GetWrittenFiles() map[string]string {
	return f.writtenFiles
}

// GetPushedBranches returns all pushed branches (for testing)
func (f *FakeLocalGit) GetPushedBranches() []string {
	return f.pushedBranches
}

// BranchExists checks if a branch exists in the fake state
func (f *FakeLocalGit) BranchExists(branchName string) (bool, error) {
	for _, branch := range f.createdBranches {
		if branch == branchName {
			return true, nil
		}
	}
	return false, nil
}
