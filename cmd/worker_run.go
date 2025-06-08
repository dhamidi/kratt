package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/dhamidi/kratt/worker"
	"github.com/spf13/cobra"
)

var workerRunCmd = &cobra.Command{
	Use:   "run <pr-number>",
	Short: "Process a specific pull request",
	Long:  "Runs the worker to process a specific pull request in the current repository.",
	Args:  cobra.ExactArgs(1),
	RunE:  runWorkerRun,
}

func init() {
	workerCmd.AddCommand(workerRunCmd)
}

func runWorkerRun(cmd *cobra.Command, args []string) error {
	// Parse PR number
	prNumber, err := strconv.Atoi(args[0])
	if err != nil || prNumber <= 0 {
		return fmt.Errorf("invalid pull request number: must be a positive integer")
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
		fmt.Printf("Processing PR #%d in repository %s/%s\n", prNumber, owner, repo)
	}

	// Load custom instructions if specified
	var instructionsText string
	if instructions != "" {
		file, err := os.Open(instructions)
		if err != nil {
			return fmt.Errorf("failed to open instructions file: %w", err)
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read instructions file: %w", err)
		}
		instructionsText = string(content)
	} else {
		instructionsText = "You are an AI assistant helping with code review. Please analyze the pull request and make any necessary improvements to the code."
	}

	// Create worker with configuration
	w := &worker.Worker{
		Instructions: instructionsText,
		AgentCommand: agentCommand,
		LintCommand:  lintCommand,
		TestCommand:  testCommand,
		Deadline:     timeout,
		Git:          gitRunner,
		GitHub:       &worker.GitHubCLI{},
		Runner:       &worker.ExecRunner{},
	}

	// Process the pull request
	if err := w.ProcessPR(prNumber); err != nil {
		return fmt.Errorf("failed to process PR #%d: %w", prNumber, err)
	}

	if verbose {
		fmt.Printf("Successfully processed PR #%d\n", prNumber)
	}

	return nil
}
