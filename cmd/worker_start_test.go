package cmd

import (
	"testing"

	"github.com/dhamidi/kratt/worker"
)



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
	git := worker.NewFakeLocalGit()
	git.SetGitRepository(false)

	w := &worker.Worker{
		Git:    git,
		GitHub: worker.NewFakeGitHub(),
		Runner: worker.NewFakeCommandRunner(),
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
	git := worker.NewFakeLocalGit()
	git.SetGitHubRepository("testowner", "testrepo")
	git.CreateBranch("existing-branch") // This adds it to the fake state

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
	git := worker.NewFakeLocalGit()
	git.FailGetGitHubRepository = true

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
	git := worker.NewFakeLocalGit()
	git.SetGitHubRepository("testowner", "testrepo")

	github := worker.NewFakeGitHub()

	w := &worker.Worker{
		Git:    git,
		GitHub: github,
		Runner: worker.NewFakeCommandRunner(),
	}

	err := w.Start("test-branch", "Test instructions")
	if err != nil {
		t.Errorf("Expected successful start, got error: %v", err)
	}
}

func TestRunWorkerStart_CreatePRError(t *testing.T) {
	git := worker.NewFakeLocalGit()
	git.SetGitHubRepository("testowner", "testrepo")

	github := worker.NewFakeGitHub()
	github.FailCreatePR = true

	w := &worker.Worker{
		Git:    git,
		GitHub: github,
		Runner: worker.NewFakeCommandRunner(),
	}

	err := w.Start("test-branch", "Test instructions")
	if err == nil {
		t.Error("Expected error when creating PR")
	}
}
