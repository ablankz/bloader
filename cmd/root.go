/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// var ctr = container.NewContainer()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bloader",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/configs/config.yaml, /etc/bloader/config.yaml, or ./config.yaml.)")
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	// ctr.Ctx = context.Background()

	configFile := viper.GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		viper.AddConfigPath(homeDir + "/configs")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/bloader")
		viper.SetConfigName("config")
		// viper.SetConfigType("yaml")
	}

	// Load environment variables
	viper.AutomaticEnv()
	// Prefix for environment variables
	viper.SetEnvPrefix("BLOADER")
	// ex. "BLOADER_SERVER_PORT" -> "server.port"
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load config file
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		os.Exit(1)
	}

	// var baseConfig map[string]any
	// if err := viper.Unmarshal(&baseConfig); err != nil {
	// 	fmt.Printf("Error unmarshalling base config: %v\n", err)
	// 	os.Exit(1)
	// }

	var cfgForOverride config.ConfigForOverride
	if err := viper.Unmarshal(&cfgForOverride, func(m *mapstructure.DecoderConfig) {
		m.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
		)
	}); err != nil {
		fmt.Printf("Error unmarshalling config: %v\n", err)
		os.Exit(1)
	}
	validForOverride, err := cfgForOverride.Validate()
	if err != nil {
		fmt.Printf("Error validating config: %v\n", err)
		os.Exit(1)
	}
	for _, override := range validForOverride.Override {
		if !override.EnabledEnv.All && !utils.Contains(override.EnabledEnv.Values, validForOverride.Env) {
			continue
		}
		switch override.Type {
		case config.OverrideTypeStatic:
			config.SetNestedValue(viper.GetViper(), override.Key, override.Value)
		case config.OverrideTypeFile:
			if override.Partial {
				// Load the file
				// Merge the file with the base config
			} else {
				f, err := os.Open(override.Path)
				if err != nil {
					fmt.Printf("failed to load file: %v\n", err)
					os.Exit(1)
				}
				defer f.Close()
				overrideFile := make(map[string]any)
				switch override.FileType {
				case config.OverrideFileTypesYAML:
					decoder := yaml.NewDecoder(f)
					if err := decoder.Decode(&overrideFile); err != nil {
						fmt.Printf("failed to decode file: %v\n", err)
						os.Exit(1)
					}
				}
				viper.MergeConfigMap(overrideFile)
			}
		case config.OverrideTypeStore:
		}
		// config.SetNestedValue(viper.GetViper(), "auth[0].oauth2.authUrl", "invalid")
		// viper.Set("auth.[0].oauth2.authUrl", "https://accounts.google.com/o/oauth2/auth")
	}
	// var key string
	// var value any
	// key = "auth[0].oauth2.auth_url"
	// value = "https://accounts.google.com/o/oauth2/auth"
	// config.SetNestedValue(viper.GetViper(), key, value)

	// key = "auth[0].oauth2.scope[1]"
	// value = "aaaa"
	// config.SetNestedValue(viper.GetViper(), key, value)

	// key = "env"
	// value = "bbb"
	// config.SetNestedValue(viper.GetViper(), key, value)

	// key = "logging.output[0].format"
	// value = "json"
	// config.SetNestedValue(viper.GetViper(), key, value)

	// key = "logging.outpu[12].format"
	// value = "js"
	// config.SetNestedValue(viper.GetViper(), key, value)
	// var currentValue any

	// if err := container.Init(cfg); err != nil {
	// 	fmt.Printf("Error initializing container: %v\n", err)
	// 	os.Exit(1)
	// }

	// if expire := container.AuthToken.IsExpired(); expire {
	// 	yellow := color.New(color.FgYellow).SprintFunc()
	// 	fmt.Println(yellow("Token has expired. Refreshing token..."))
	// 	if err := container.AuthToken.Refresh(container.Ctx, createOAuthConfig(), container.Config.Credential.Path); err != nil {
	// 		red := color.New(color.FgRed).SprintFunc()
	// 		fmt.Println(red(fmt.Sprintf("Failed to refresh token: %v", err)))
	// 		fmt.Println("You may need to re-authenticate, if want to access the credential API.")
	// 	} else {
	// 		green := color.New(color.FgGreen).SprintFunc()
	// 		fmt.Println(green("Successfully refreshed token"))
	// 	}
	// }
}
