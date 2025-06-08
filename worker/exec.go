package worker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CommandRunner interface encapsulates command execution
type CommandRunner interface {
	// RunWithStdin executes a command with the given stdin input
	RunWithStdin(ctx context.Context, stdin string, command string, args ...string) error

	// RunWithOutput executes a command and returns interleaved stdout/stderr output
	RunWithOutput(ctx context.Context, command string, args ...string) (output []byte, err error)
}

// ExecRunner implements CommandRunner interface using os/exec
type ExecRunner struct{}

// RunWithStdin executes a command with the given stdin input
func (e *ExecRunner) RunWithStdin(ctx context.Context, stdin string, command string, args ...string) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdin = strings.NewReader(stdin)
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", command, args, err)
	}
	return nil
}

// RunWithOutput executes a command and returns interleaved stdout/stderr output
func (e *ExecRunner) RunWithOutput(ctx context.Context, command string, args ...string) (output []byte, err error) {
	cmd := exec.CommandContext(ctx, command, args...)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("command %s %v failed: %w", command, args, err)
	}
	return output, nil
}

// FakeCommandRunner implements CommandRunner interface for testing
type FakeCommandRunner struct {
	stdinInputs map[string]string         // command -> stdin input (for verification)
	responses   map[string][]byte         // command -> output response
	errors      map[string]error          // command -> error to return
}

// NewFakeCommandRunner creates a new FakeCommandRunner instance
func NewFakeCommandRunner() *FakeCommandRunner {
	return &FakeCommandRunner{
		stdinInputs: make(map[string]string),
		responses:   make(map[string][]byte),
		errors:      make(map[string]error),
	}
}

// SetResponse configures the response for a command pattern
func (f *FakeCommandRunner) SetResponse(commandPattern string, output []byte, err error) {
	f.responses[commandPattern] = output
	if err != nil {
		f.errors[commandPattern] = err
	}
}

// RunWithStdin records stdin input and returns configured response
func (f *FakeCommandRunner) RunWithStdin(ctx context.Context, stdin string, command string, args ...string) error {
	cmdKey := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	f.stdinInputs[cmdKey] = stdin
	
	if err, exists := f.errors[cmdKey]; exists {
		return err
	}
	return nil
}

// RunWithOutput returns configured output and error
func (f *FakeCommandRunner) RunWithOutput(ctx context.Context, command string, args ...string) (output []byte, err error) {
	cmdKey := fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	
	if output, exists := f.responses[cmdKey]; exists {
		if err, hasErr := f.errors[cmdKey]; hasErr {
			return output, err
		}
		return output, nil
	}
	
	// Default response if not configured
	return []byte("fake output"), nil
}

// GetStdinInput returns recorded stdin input for verification (for testing)
func (f *FakeCommandRunner) GetStdinInput(command string) string {
	return f.stdinInputs[command]
}
