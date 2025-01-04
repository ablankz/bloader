/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ablankz/bloader/internal/slave"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// slaveRunCmd represents the slaveRun command
var slaveRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the worker node",
	Long: `This command is used to start a worker node.
A worker node is a node that is responsible for running the tasks assigned by the master node.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(ctr.Ctx)
		defer cancel()

		ctr.Ctx = ctx

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

		go func() {
			<-signalChan
			fmt.Println("\nSignal received: cleaning up...")
			cancel()
		}()

		if err := slave.SlaveRun(ctr); err != nil {
			color.Red("Failed to run the worker node: %v", err)
			return
		}
	},
}

func init() {
	slaveCmd.AddCommand(slaveRunCmd)
}
