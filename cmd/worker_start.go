package cmd

import (
	"fmt"
	"regexp"

	"github.com/dhamidi/kratt/worker"
	"github.com/spf13/cobra"
)

var workerStartCmd = &cobra.Command{
	Use:   "start <branch-name> <instructions>",
	Short: "Create a new branch with instructions and open a pull request",
	Long:  "Creates a new branch with instructions and opens a pull request for implementation.",
	Args:  cobra.ExactArgs(2),
	RunE:  runWorkerStart,
}

func init() {
	workerCmd.AddCommand(workerStartCmd)
}

func runWorkerStart(cmd *cobra.Command, args []string) error {
	branchName := args[0]
	instructions := args[1]

	// Validate branch name
	if !isValidBranchName(branchName) {
		return fmt.Errorf("invalid branch name: must not contain invalid characters")
	}

	// Create git runner and check if we're in a git repository
	gitRunner := &worker.GitRunner{}
	isGitRepo, err := gitRunner.IsGitRepository()
	if err != nil {
		return fmt.Errorf("error checking git repository: %w", err)
	}
	if !isGitRepo {
		return fmt.Errorf("current directory is not a git repository")
	}

	// Get GitHub repository information
	owner, repo, err := gitRunner.GetGitHubRepository()
	if err != nil {
		return fmt.Errorf("no GitHub remote found in current repository: %w", err)
	}

	if verbose {
		fmt.Printf("Creating branch %s in repository %s/%s\n", branchName, owner, repo)
	}

	// Check if branch already exists
	exists, err := gitRunner.BranchExists(branchName)
	if err != nil {
		return fmt.Errorf("error checking if branch exists: %w", err)
	}
	if exists {
		return fmt.Errorf("branch already exists: %s", branchName)
	}

	// Create worker with configuration
	w := &worker.Worker{
		Instructions: "You are an AI assistant helping with implementation. Please analyze the instructions and implement the requested feature.",
		AgentCommand: agentCommand,
		LintCommand:  lintCommand,
		TestCommand:  testCommand,
		Deadline:     timeout,
		Git:          gitRunner,
		GitHub:       &worker.GitHubCLI{},
		Runner:       &worker.ExecRunner{},
	}

	// Start the new branch and create PR
	if err := w.Start(branchName, instructions); err != nil {
		return fmt.Errorf("failed to start branch %s: %w", branchName, err)
	}

	if verbose {
		fmt.Printf("Successfully created branch %s and opened pull request\n", branchName)
	}

	return nil
}

// isValidBranchName validates that a branch name doesn't contain invalid characters
func isValidBranchName(name string) bool {
	// Git branch names cannot contain certain characters
	// This is a basic validation - can be enhanced based on git's actual rules
	invalidChars := regexp.MustCompile(`[~^:?*\[\]\s\\]`)
	if invalidChars.MatchString(name) {
		return false
	}
	
	// Cannot start or end with dot or slash
	if len(name) == 0 || name[0] == '.' || name[0] == '/' || 
		name[len(name)-1] == '.' || name[len(name)-1] == '/' {
		return false
	}
	
	// Cannot contain double dots
	if regexp.MustCompile(`\.\.`).MatchString(name) {
		return false
	}
	
	return true
}
