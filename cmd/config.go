/*
Copyright © 2024 hayashi kenta <k.hayashi@cresplanex.com>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the current configuration",
	Long: `This command prints the current configuration.
It reads the configuration from the configuration file and prints it in YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		exConfigSettings := map[string]any{}
		for key, value := range settings {
			if key != "config" {
				exConfigSettings[key] = value
			}
		}
		yamlData, err := yaml.Marshal(exConfigSettings)
		if err != nil {
			log.Fatalf("Error converting configuration to YAML: %v", err)
		}
		fmt.Println("Current configuration:")
		fmt.Println(string(yamlData))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
