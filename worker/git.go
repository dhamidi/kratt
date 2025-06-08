package worker

import (
	"fmt"
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

	// Push changes
	pushCmd := exec.Command("git", "push")
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

// FakeLocalGit implements LocalGit interface for testing
type FakeLocalGit struct {
	worktrees map[string]string // branch -> path mapping
	currentDir string
	commits []string
}

// NewFakeLocalGit creates a new FakeLocalGit instance
func NewFakeLocalGit() *FakeLocalGit {
	return &FakeLocalGit{
		worktrees: make(map[string]string),
		currentDir: "/fake/repo",
		commits: []string{},
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
