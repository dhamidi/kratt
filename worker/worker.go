package worker

import "time"

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
