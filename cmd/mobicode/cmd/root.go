package cmd

import "github.com/spf13/cobra"

var configPath string

var rootCmd = &cobra.Command{
	Use:   "mobicode",
	Short: "Run the Mobicode API and manage its database",
}

// Execute runs the root CLI command.
func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "path to .env configuration file")
}
