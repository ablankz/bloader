/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// storeObjectDeleteCmd represents the storeObjectDelete command
var storeObjectDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an object from the specified bucket",
	Long: `This command deletes an object from the specified bucket.
It sends a request to the storage server to delete the object.
For example:

bloader store object delete --bucket 1234 objectKey`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide the object key")
			return
		}
		objKey := args[0]
		if err := ctr.Store.DeleteObject(bucketID, objKey); err != nil {
			color.Red("Failed to delete object: %v", err)
			return
		}
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Println(green("Object deleted"))
	},
}

func init() {
	storeObjectCmd.AddCommand(storeObjectDeleteCmd)
}
