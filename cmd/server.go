package cmd

import (
	"watchdog/main.go/internal/app"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		application, err := app.NewApplication(cmd.OutOrStdout())
		if err != nil {
			return err
		}

		return application.Run(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
