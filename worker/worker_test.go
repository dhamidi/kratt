package worker

import (
	"context"
	"testing"
	"time"
)

func TestWorkerProcessPR(t *testing.T) {
	// Setup fakes
	fakeGit := NewFakeLocalGit()
	fakeGitHub := NewFakeGitHub()
	fakeRunner := NewFakeCommandRunner()

	// Configure test data
	prInfo := `{
		"title": "Test PR",
		"body": "This is a test PR",
		"headRefName": "feature-branch",
		"comments": []
	}`
	fakeGitHub.SetPRInfo(123, prInfo)

	// Configure command responses
	fakeRunner.SetResponse("goimports -w ./...", []byte("goimports output"), nil)
	fakeRunner.SetResponse("go test ./...", []byte("PASS\nok"), nil)

	// Create worker
	worker := &Worker{
		Instructions: "You are a helpful AI assistant.",
		AgentCommand: []string{"echo", "agent-output"},
		LintCommand:  []string{"goimports", "-w", "./..."},
		TestCommand:  []string{"go", "test", "./..."},
		Deadline:     5 * time.Second,
		Git:          fakeGit,
		GitHub:       fakeGitHub,
		Runner:       fakeRunner,
	}

	// Test ProcessPR
	err := worker.ProcessPR(123)
	if err != nil {
		t.Fatalf("ProcessPR failed: %v", err)
	}

	// Verify worktree was created
	exists, _ := fakeGit.CheckWorktreeExists("feature-branch")
	if !exists {
		t.Error("Expected worktree to be created for feature-branch")
	}

	// Verify comment was posted
	comments := fakeGitHub.GetComments(123)
	if len(comments) == 0 {
		t.Error("Expected comment to be posted")
	}

	// Verify changes were committed
	commits := fakeGit.GetCommits()
	if len(commits) == 0 {
		t.Error("Expected changes to be committed")
	}
}

func TestFakeLocalGit(t *testing.T) {
	fake := NewFakeLocalGit()

	// Test worktree operations
	exists, err := fake.CheckWorktreeExists("test-branch")
	if err != nil || exists {
		t.Error("Expected worktree not to exist initially")
	}

	err = fake.CreateWorktree("test-branch", "/fake/path")
	if err != nil {
		t.Fatalf("CreateWorktree failed: %v", err)
	}

	exists, err = fake.CheckWorktreeExists("test-branch")
	if err != nil || !exists {
		t.Error("Expected worktree to exist after creation")
	}

	// Test directory change
	err = fake.ChangeDirectory("/new/path")
	if err != nil {
		t.Fatalf("ChangeDirectory failed: %v", err)
	}

	if fake.GetCurrentDir() != "/new/path" {
		t.Error("Expected current directory to be updated")
	}

	// Test commit
	err = fake.CommitAndPush("test commit")
	if err != nil {
		t.Fatalf("CommitAndPush failed: %v", err)
	}

	commits := fake.GetCommits()
	if len(commits) != 1 || commits[0] != "test commit" {
		t.Error("Expected commit to be recorded")
	}

	// Test repository detection
	isRepo, err := fake.IsGitRepository()
	if err != nil {
		t.Fatalf("IsGitRepository failed: %v", err)
	}
	if !isRepo {
		t.Error("Expected default to be a git repository")
	}

	// Test GitHub repository detection
	owner, repo, err := fake.GetGitHubRepository()
	if err != nil {
		t.Fatalf("GetGitHubRepository failed: %v", err)
	}
	if owner != "owner" || repo != "repo" {
		t.Errorf("Expected owner='owner' repo='repo', got owner='%s' repo='%s'", owner, repo)
	}

	// Test setting repository status
	fake.SetGitRepository(false)
	isRepo, err = fake.IsGitRepository()
	if err != nil {
		t.Fatalf("IsGitRepository failed: %v", err)
	}
	if isRepo {
		t.Error("Expected repository status to be false after setting")
	}

	// Test setting GitHub repository
	fake.SetGitHubRepository("testowner", "testrepo")
	owner, repo, err = fake.GetGitHubRepository()
	if err != nil {
		t.Fatalf("GetGitHubRepository failed: %v", err)
	}
	if owner != "testowner" || repo != "testrepo" {
		t.Errorf("Expected owner='testowner' repo='testrepo', got owner='%s' repo='%s'", owner, repo)
	}
}

func TestFakeGitHub(t *testing.T) {
	fake := NewFakeGitHub()

	// Test PR info
	_, err := fake.GetPRInfo(123)
	if err == nil {
		t.Error("Expected error when getting non-existent PR")
	}

	fake.SetPRInfo(123, "test pr info")
	info, err := fake.GetPRInfo(123)
	if err != nil || info != "test pr info" {
		t.Error("Expected to get stored PR info")
	}

	// Test comments
	err = fake.PostComment(123, "test comment")
	if err != nil {
		t.Fatalf("PostComment failed: %v", err)
	}

	comments := fake.GetComments(123)
	if len(comments) != 1 || comments[0] != "test comment" {
		t.Error("Expected comment to be stored")
	}
}

func TestFakeCommandRunner(t *testing.T) {
	fake := NewFakeCommandRunner()
	ctx := context.Background()

	// Test RunWithStdin
	err := fake.RunWithStdin(ctx, "test input", "echo", "hello")
	if err != nil {
		t.Fatalf("RunWithStdin failed: %v", err)
	}

	input := fake.GetStdinInput("echo hello")
	if input != "test input" {
		t.Error("Expected stdin input to be recorded")
	}

	// Test RunWithOutput
	fake.SetResponse("ls -la", []byte("test output"), nil)
	output, err := fake.RunWithOutput(ctx, "ls", "-la")
	if err != nil || string(output) != "test output" {
		t.Error("Expected configured output to be returned")
	}
}
