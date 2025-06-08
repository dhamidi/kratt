package worker

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Worker implements an automated pull request processing system
type Worker struct {
	Instructions string        // Prefix for the agent prompt
	AgentCommand []string      // Command to run the AI agent
	LintCommand  []string      // Command to run linting
	TestCommand  []string      // Command to run tests
	Deadline     time.Duration // Maximum time for agent execution

	// Dependencies (injected for testability)
	Git    LocalGit
	GitHub GitHub
	Runner CommandRunner
}

// ProcessPR processes a pull request by running the agent and posting results
func (w *Worker) ProcessPR(prNumber int) error {
	// 3.1: Get PR Information
	prInfo, err := w.GitHub.GetPRInfo(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR info: %w", err)
	}

	// 3.2: Handle Git Worktree
	branch, err := w.extractBranchFromPRInfo(prInfo)
	if err != nil {
		return fmt.Errorf("failed to extract branch from PR info: %w", err)
	}

	exists, err := w.Git.CheckWorktreeExists(branch)
	if err != nil {
		return fmt.Errorf("failed to check worktree existence: %w", err)
	}

	if !exists {
		path, err := w.Git.GetWorktreePath(branch)
		if err != nil {
			return fmt.Errorf("failed to get worktree path: %w", err)
		}
		
		err = w.Git.CreateWorktree(branch, path)
		if err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	path, err := w.Git.GetWorktreePath(branch)
	if err != nil {
		return fmt.Errorf("failed to get worktree path: %w", err)
	}

	err = w.Git.ChangeDirectory(path)
	if err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// 3.3: Generate Agent Prompt
	prompt := w.generatePrompt(prInfo)

	// 3.4: Execute Agent with Timeout
	ctx, cancel := context.WithTimeout(context.Background(), w.Deadline)
	defer cancel()

	err = w.Runner.RunWithStdin(ctx, prompt, w.AgentCommand[0], w.AgentCommand[1:]...)
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	// 3.5: Run Lint and Test Commands
	lintOutput, lintErr := w.Runner.RunWithOutput(ctx, w.LintCommand[0], w.LintCommand[1:]...)
	testOutput, testErr := w.Runner.RunWithOutput(ctx, w.TestCommand[0], w.TestCommand[1:]...)

	// 3.6: Post Results Comment
	commentBody := w.formatResultsComment(lintOutput, lintErr, testOutput, testErr)
	err = w.GitHub.PostComment(prNumber, commentBody)
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}

	// 3.7: Commit and Push Changes
	err = w.Git.CommitAndPush("Automated changes from kratt worker")
	if err != nil {
		return fmt.Errorf("failed to commit and push: %w", err)
	}

	return nil
}

// extractBranchFromPRInfo extracts the branch name from PR information
func (w *Worker) extractBranchFromPRInfo(prInfo string) (string, error) {
	// Look for headRefName in JSON format returned by gh CLI
	re := regexp.MustCompile(`"headRefName":\s*"([^"]+)"`)
	matches := re.FindStringSubmatch(prInfo)
	if len(matches) > 1 {
		return matches[1], nil
	}
	
	// Fallback: look for common patterns in PR info that indicate the branch name
	re = regexp.MustCompile(`(?i)branch:\s*([^\s\n]+)`)
	matches = re.FindStringSubmatch(prInfo)
	if len(matches) > 1 {
		return matches[1], nil
	}
	
	// Fallback: look for "head:" pattern
	re = regexp.MustCompile(`(?i)head:\s*([^\s\n]+)`)
	matches = re.FindStringSubmatch(prInfo)
	if len(matches) > 1 {
		return matches[1], nil
	}
	
	return "", fmt.Errorf("could not extract branch name from PR info")
}

// generatePrompt creates the prompt for the agent
func (w *Worker) generatePrompt(prInfo string) string {
	var prompt strings.Builder
	prompt.WriteString(w.Instructions)
	prompt.WriteString("\n\n")
	prompt.WriteString("<pull-request>\n")
	prompt.WriteString(prInfo)
	prompt.WriteString("\n</pull-request>")
	return prompt.String()
}

// formatResultsComment formats the lint and test results into a comment
func (w *Worker) formatResultsComment(lintOutput []byte, lintErr error, testOutput []byte, testErr error) string {
	var comment strings.Builder
	
	comment.WriteString("## Kratt Worker Results\n\n")
	
	// Lint results
	comment.WriteString("### Lint Results\n")
	if lintErr != nil {
		comment.WriteString("❌ **Failed**\n")
		comment.WriteString("```\n")
		comment.WriteString(lintErr.Error())
		comment.WriteString("\n```\n")
	} else {
		comment.WriteString("✅ **Passed**\n")
	}
	
	if len(lintOutput) > 0 {
		comment.WriteString("```\n")
		comment.WriteString(string(lintOutput))
		comment.WriteString("\n```\n")
	}
	
	comment.WriteString("\n")
	
	// Test results
	comment.WriteString("### Test Results\n")
	if testErr != nil {
		comment.WriteString("❌ **Failed**\n")
		comment.WriteString("```\n")
		comment.WriteString(testErr.Error())
		comment.WriteString("\n```\n")
	} else {
		comment.WriteString("✅ **Passed**\n")
	}
	
	if len(testOutput) > 0 {
		comment.WriteString("```\n")
		comment.WriteString(string(testOutput))
		comment.WriteString("\n```\n")
	}
	
	return comment.String()
}

// Start creates a new branch and pull request with instructions
func (w *Worker) Start(branchName string, instruction string) error {
	// 8.1: Create and Switch to New Branch
	err := w.Git.CreateBranch(branchName)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	// 8.2: Write Instructions File
	filePath := fmt.Sprintf("docs/%s-instructions.md", branchName)
	err = w.Git.WriteFile(filePath, instruction)
	if err != nil {
		return fmt.Errorf("failed to write instructions file: %w", err)
	}

	// 8.3: Commit Instructions File
	err = w.Git.CommitAndPush("Add instructions for " + branchName)
	if err != nil {
		return fmt.Errorf("failed to commit instructions file: %w", err)
	}

	// 8.4: Push Branch Upstream
	err = w.Git.PushBranchUpstream(branchName)
	if err != nil {
		return fmt.Errorf("failed to push branch upstream: %w", err)
	}

	// 8.5: Create Pull Request
	title := "Implement " + branchName
	description := fmt.Sprintf("Study docs/%s-instructions.md and make a list of necessary implementation steps in docs/%s-implementation-status.md", branchName, branchName)
	err = w.GitHub.CreatePR(title, description)
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}

	return nil
}
