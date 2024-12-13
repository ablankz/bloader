/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// storeListCmd represents the storeList command
var storeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a list of buckets in the store.",
	Long:  `You can view a list of buckets currently held in the store.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("storeList called")
	},
}

func init() {
	rootCmd.AddCommand(storeListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storeListCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storeListCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
