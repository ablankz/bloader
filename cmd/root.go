/*
Copyright © 2024 hayashi kenta <k.hayashi@cresplanex.com>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ctr = container.NewContainer()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bloader",
	Short: "The tool for load testing",
	Long: `This tool is used to perform load testing.
It sends requests to the specified server and measures the response time.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	defer ctr.Close()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is ./bloaderconfig.yaml, $HOME/configs/config.yaml, or /etc/bloader/config.yaml)")
	if err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")); err != nil {
		fmt.Printf("Error binding flag: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	ctr.Ctx = context.Background()

	configFile := viper.GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Failed to get home directory: %v\n", err)
			os.Exit(1)
		}
		viper.AddConfigPath("./bloader")
		viper.AddConfigPath(homeDir + "/configs")
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
				f, err := os.Open(override.Path)
				if err != nil {
					fmt.Printf("failed to load file: %v\n", err)
					os.Exit(1)
				}
				defer f.Close()
				overrideMap := make(map[string]any)
				switch override.FileType {
				case config.OverrideFileTypesYAML:
					decoder := yaml.NewDecoder(f)
					if err := decoder.Decode(&overrideMap); err != nil {
						fmt.Printf("failed to decode file: %v\n", err)
						os.Exit(1)
					}
				}
				for _, v := range override.Vars {
					value := config.GetNestedValueFromMap(overrideMap, v.Value)
					config.SetNestedValue(viper.GetViper(), v.Key, value)
				}
			} else {
				f, err := os.Open(override.Path)
				if err != nil {
					fmt.Printf("failed to load file: %v\n", err)
					os.Exit(1)
				}
				defer f.Close()
				overrideMap := make(map[string]any)
				switch override.FileType {
				case config.OverrideFileTypesYAML:
					decoder := yaml.NewDecoder(f)
					if err := decoder.Decode(&overrideMap); err != nil {
						fmt.Printf("failed to decode file: %v\n", err)
						os.Exit(1)
					}
				}
				viper.MergeConfigMap(overrideMap)
			}
		}
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Printf("Error unmarshalling config: %v\n", err)
		os.Exit(1)
	}
	validCfg, err := cfg.Validate()
	if err != nil {
		fmt.Printf("Error validating config: %v\n", err)
		os.Exit(1)
	}

	if err := ctr.Init(validCfg); err != nil {
		fmt.Printf("Error initializing container: %v\n", err)
		os.Exit(1)
	}

	// for k, v := range ctr.AuthenticatorContainer.Container {
	// 	if expired := (*v).IsExpired(ctr.Ctx, ctr.Store); expired {
	// 		yellow := color.New(color.FgYellow).SprintFunc()
	// 		fmt.Printf(yellow("Token for %s has expired. Refreshing token...\n"), k)
	// 		if err := (*v).Refresh(ctr.Ctx, ctr.Store); err != nil {
	// 			red := color.New(color.FgRed).SprintFunc()
	// 			fmt.Printf(red("Failed to refresh token: %v\n"), err)
	// 			fmt.Printf("You may need to re-authenticate, if want to access the credential API.\n")
	// 		} else {
	// 			green := color.New(color.FgGreen).SprintFunc()
	// 			fmt.Printf(green("Successfully refreshed token for %s\n"), k)
	// 		}
	// 	}
	// }
}
