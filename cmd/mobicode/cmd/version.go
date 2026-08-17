package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is overridden at build time with -ldflags.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "mobicode version", Version)
	},
}

func init() { rootCmd.AddCommand(versionCmd) }
