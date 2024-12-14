/*
Copyright © 2024 hayashi kenta <k.hayashi@cresplanex.com>
*/
package cmd

import (
	"github.com/ablankz/bloader/internal/runner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var runnerFile string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the load test",
	Long: `This command runs the load test.
It sends requests to the specified server and measures the response time.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runner.Run(ctr, runnerFile); err != nil {
			color.Red("Failed to run the load test: %v\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&runnerFile, "file", "f", "", "The file to run the load test")
}
