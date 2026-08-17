package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/database"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the data directory and run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if err := database.RunMigrations(cfg.DatabasePath); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
		cmd.Println("Initialization complete. Run `mobicode serve` to start the server.")
		return nil
	},
}

func init() { rootCmd.AddCommand(initCmd) }
