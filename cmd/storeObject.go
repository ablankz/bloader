/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var bucketID string
var storeObjectEncrypt string

// storeObjectCmd represents the storeObject command
var storeObjectCmd = &cobra.Command{
	Use:   "object",
	Short: "Store an object in the specified bucket",
	Long: `This command stores an object in the specified bucket.
It reads the object from the specified file and sends a request to the storage server to store the object.`,
}

func init() {
	storeCmd.AddCommand(storeObjectCmd)

	storeObjectCmd.PersistentFlags().StringVarP(&bucketID, "bucket", "b", "", "ID of the bucket where the object will be stored")
	storeObjectCmd.MarkPersistentFlagRequired("bucket")
}
