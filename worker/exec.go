package worker

import "context"

// CommandRunner interface encapsulates command execution
type CommandRunner interface {
	// RunWithStdin executes a command with the given stdin input
	RunWithStdin(ctx context.Context, stdin string, command string, args ...string) error

	// RunWithOutput executes a command and returns interleaved stdout/stderr output
	RunWithOutput(ctx context.Context, command string, args ...string) (output []byte, err error)
}
