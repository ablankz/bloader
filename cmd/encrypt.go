/*
Copyright © 2024 hayashi kenta <k.hayashi@cresplanex.com>
*/
package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var decrypt bool
var encryptID string

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt text",
	Long: `This command encrypts text.
It encrypts the text using the encryption key.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			color.Red("Text is required")
			return
		}
		target := args[0]
		encryper, ok := ctr.Encypter[encryptID]
		if !ok {
			color.Red("Encrypt setting not found")
			return
		}

		if decrypt {
			b, err := encryper.Decrypt(target)
			if err != nil {
				color.Red("Failed to decrypt: %v", err)
				return
			}
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Println(green("Decrypted text:"), string(b))
			return
		} else {
			s, err := encryper.Encrypt([]byte(target))
			if err != nil {
				color.Red("Failed to encrypt: %v", err)
				return
			}
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Println(green("Encrypted text:"), s)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().BoolVarP(&decrypt, "decrypt", "d", false, "Switch to decrypt mode")
	encryptCmd.Flags().StringVarP(&encryptID, "id", "i", "", `ID of the encrypt setting. This is required.`)
	encryptCmd.MarkFlagRequired("id")
}
