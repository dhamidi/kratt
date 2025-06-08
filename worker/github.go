package worker

import (
	"fmt"
	"os/exec"
	"strconv"
)

// GitHub interface encapsulates GitHub operations
type GitHub interface {
	// GetPRInfo retrieves pull request information including comments
	GetPRInfo(prNumber int) (string, error)

	// PostComment posts a comment to the specified pull request
	PostComment(prNumber int, body string) error
}

// GitHubCLI implements GitHub interface using GitHub CLI
type GitHubCLI struct{}

// GetPRInfo retrieves pull request information using gh CLI
func (g *GitHubCLI) GetPRInfo(prNumber int) (string, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber), "--json", "title,body,headRefName,comments")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get PR info for #%d: %w", prNumber, err)
	}
	return string(output), nil
}

// PostComment posts a comment to the specified pull request using gh CLI
func (g *GitHubCLI) PostComment(prNumber int, body string) error {
	cmd := exec.Command("gh", "pr", "comment", strconv.Itoa(prNumber), "--body", body)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to post comment to PR #%d: %w", prNumber, err)
	}
	return nil
}

// FakeGitHub implements GitHub interface for testing
type FakeGitHub struct {
	prData   map[int]string            // prNumber -> PR info
	comments map[int][]string          // prNumber -> list of comments
}

// NewFakeGitHub creates a new FakeGitHub instance
func NewFakeGitHub() *FakeGitHub {
	return &FakeGitHub{
		prData:   make(map[int]string),
		comments: make(map[int][]string),
	}
}

// SetPRInfo sets the PR information for testing
func (f *FakeGitHub) SetPRInfo(prNumber int, info string) {
	f.prData[prNumber] = info
}

// GetPRInfo returns stored PR information
func (f *FakeGitHub) GetPRInfo(prNumber int) (string, error) {
	if info, exists := f.prData[prNumber]; exists {
		return info, nil
	}
	return "", fmt.Errorf("PR #%d not found", prNumber)
}

// PostComment adds a comment to the fake storage
func (f *FakeGitHub) PostComment(prNumber int, body string) error {
	if _, exists := f.comments[prNumber]; !exists {
		f.comments[prNumber] = []string{}
	}
	f.comments[prNumber] = append(f.comments[prNumber], body)
	return nil
}

// GetComments returns all comments for a PR (for testing)
func (f *FakeGitHub) GetComments(prNumber int) []string {
	return f.comments[prNumber]
}
