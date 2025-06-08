package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	timeout      time.Duration
	instructions string
	agentCommand []string
	lintCommand  []string
	testCommand  []string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "kratt",
	Short: "Automated PR processing worker",
	Long:  "Kratt provides a command-line interface for running automated PR processing with AI agents.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 30*time.Minute, "Maximum time for agent execution")
	rootCmd.PersistentFlags().StringVar(&instructions, "instructions", "", "Path to file containing agent instructions")
	rootCmd.PersistentFlags().StringSliceVar(&agentCommand, "agent", []string{"amp", "--stdin"}, "Command to run the AI agent")
	rootCmd.PersistentFlags().StringSliceVar(&lintCommand, "lint", []string{"go", "fmt", "./..."}, "Command to run linting")
	rootCmd.PersistentFlags().StringSliceVar(&testCommand, "test", []string{"go", "test", "./..."}, "Command to run tests")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
}
