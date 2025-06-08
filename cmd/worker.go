package cmd

import (
	"github.com/spf13/cobra"
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker commands for processing pull requests",
	Long:  "Commands for running the automated PR processing worker.",
}

func init() {
	rootCmd.AddCommand(workerCmd)
}
