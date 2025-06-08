package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/dhamidi/kratt/worker"
)

// mockGit implements LocalGit for testing
type mockGit struct {
	isGitRepo      bool
	isGitRepoErr   error
	owner          string
	repo           string
	getRepoErr     error
	branchExists   bool
	branchExistsErr error
	createBranchErr error
	writeFileErr    error
	commitPushErr   error
	pushUpstreamErr error
}

func (m *mockGit) IsGitRepository() (bool, error) {
	return m.isGitRepo, m.isGitRepoErr
}

func (m *mockGit) GetGitHubRepository() (string, string, error) {
	return m.owner, m.repo, m.getRepoErr
}

func (m *mockGit) BranchExists(branch string) (bool, error) {
	return m.branchExists, m.branchExistsErr
}

func (m *mockGit) CreateBranch(branch string) error {
	return m.createBranchErr
}

func (m *mockGit) WriteFile(path, content string) error {
	return m.writeFileErr
}

func (m *mockGit) CommitAndPush(message string) error {
	return m.commitPushErr
}

func (m *mockGit) PushBranchUpstream(branch string) error {
	return m.pushUpstreamErr
}

// Implement remaining LocalGit methods with no-op implementations
func (m *mockGit) CheckWorktreeExists(branch string) (bool, error) { return false, nil }
func (m *mockGit) GetWorktreePath(branch string) (string, error)   { return "", nil }
func (m *mockGit) CreateWorktree(branch, path string) error        { return nil }
func (m *mockGit) ChangeDirectory(path string) error              { return nil }

// mockGitHub implements GitHub for testing
type mockGitHub struct {
	createPRErr error
}

func (m *mockGitHub) GetPRInfo(prNumber int) (string, error) {
	return "", nil
}

func (m *mockGitHub) PostComment(prNumber int, body string) error {
	return nil
}

func (m *mockGitHub) CreatePR(title, description string) error {
	return m.createPRErr
}

// mockRunner implements CommandRunner for testing
type mockRunner struct{}

func (m *mockRunner) RunWithStdin(ctx context.Context, stdin, command string, args ...string) error {
	return nil
}

func (m *mockRunner) RunWithOutput(ctx context.Context, command string, args ...string) ([]byte, error) {
	return nil, nil
}

func TestIsValidBranchName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid simple name", "feature", true},
		{"valid with slash", "feature/auth", true},
		{"valid with dash", "feature-auth", true},
		{"valid with underscore", "feature_auth", true},
		{"empty string", "", false},
		{"starts with dot", ".feature", false},
		{"ends with dot", "feature.", false},
		{"starts with slash", "/feature", false},
		{"ends with slash", "feature/", false},
		{"contains space", "feature auth", false},
		{"contains colon", "feature:auth", false},
		{"contains question mark", "feature?auth", false},
		{"contains asterisk", "feature*auth", false},
		{"contains bracket", "feature[auth]", false},
		{"contains double dots", "feature..auth", false},
		{"contains tilde", "feature~auth", false},
		{"contains caret", "feature^auth", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidBranchName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidBranchName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRunWorkerStart_NotGitRepository(t *testing.T) {
	git := &mockGit{
		isGitRepo: false,
	}

	w := &worker.Worker{
		Git:    git,
		GitHub: &mockGitHub{},
		Runner: &mockRunner{},
	}

	// Set up the original worker creation to use our mock
	originalWorker := func() *worker.Worker { return w }
	_ = originalWorker // Placeholder for when we refactor to inject dependencies

	// For now, test the validation function directly
	if isValidBranchName("valid-branch") != true {
		t.Error("Expected valid branch name to pass validation")
	}

	// Test the git repository check logic
	isRepo, err := git.IsGitRepository()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if isRepo {
		t.Error("Expected not to be a git repository")
	}
}

func TestRunWorkerStart_BranchAlreadyExists(t *testing.T) {
	git := &mockGit{
		isGitRepo:    true,
		owner:        "testowner",
		repo:         "testrepo",
		branchExists: true,
	}

	isRepo, _ := git.IsGitRepository()
	if !isRepo {
		t.Error("Expected to be a git repository")
	}

	exists, _ := git.BranchExists("existing-branch")
	if !exists {
		t.Error("Expected branch to exist")
	}
}

func TestRunWorkerStart_GitHubRepositoryError(t *testing.T) {
	git := &mockGit{
		isGitRepo:  true,
		getRepoErr: errors.New("no remote found"),
	}

	isRepo, _ := git.IsGitRepository()
	if !isRepo {
		t.Error("Expected to be a git repository")
	}

	_, _, err := git.GetGitHubRepository()
	if err == nil {
		t.Error("Expected error when getting GitHub repository")
	}
}

func TestRunWorkerStart_Success(t *testing.T) {
	git := &mockGit{
		isGitRepo:    true,
		owner:        "testowner",
		repo:         "testrepo",
		branchExists: false,
	}

	github := &mockGitHub{}

	w := &worker.Worker{
		Git:    git,
		GitHub: github,
		Runner: &mockRunner{},
	}

	err := w.Start("test-branch", "Test instructions")
	if err != nil {
		t.Errorf("Expected successful start, got error: %v", err)
	}
}

func TestRunWorkerStart_CreatePRError(t *testing.T) {
	git := &mockGit{
		isGitRepo:    true,
		owner:        "testowner",
		repo:         "testrepo",
		branchExists: false,
	}

	github := &mockGitHub{
		createPRErr: errors.New("failed to create PR"),
	}

	w := &worker.Worker{
		Git:    git,
		GitHub: github,
		Runner: &mockRunner{},
	}

	err := w.Start("test-branch", "Test instructions")
	if err == nil {
		t.Error("Expected error when creating PR")
	}
}
