package cmd

import (
	"watchdog/main.go/internal/configuration"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Displays the current application configuration",
	Long: `Displays the current Watchdog HTTP configuration.

The command reads the application's environment variables with the WATCHDOG_
prefix and prints them in alphabetical order.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println("config called")
		return configuration.NewDisplay().Run()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
