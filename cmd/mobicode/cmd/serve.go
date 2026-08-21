package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/database"
	"github.com/soumajitgh/mobicode/internal/graphql"
	"github.com/soumajitgh/mobicode/internal/health"
	"github.com/soumajitgh/mobicode/internal/logger"
	"github.com/soumajitgh/mobicode/internal/server"
	"github.com/soumajitgh/mobicode/internal/setup"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		app := fx.New(
			fx.Supply(cfg),
			auth.Module,
			logger.Module,
			database.Module,
			server.Module,
			health.Module,
			graphql.Module,
			setup.Module,
			fx.WithLogger(logger.NewFxLogger),
		)
		if err := app.Err(); err != nil {
			return err
		}
		app.Run()
		return nil
	},
}

func init() { rootCmd.AddCommand(serveCmd) }
