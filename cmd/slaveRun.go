/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// slaveRunCmd represents the slaveRun command
var slaveRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the worker node",
	Long: `This command is used to start a worker node.
A worker node is a node that is responsible for running the tasks assigned by the master node.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("slaveRun called")
	},
}

func init() {
	slaveCmd.AddCommand(slaveRunCmd)
}
