package cmd

import (
	"context"
	"fmt"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
	"github.com/soumajitgh/mobicode/internal/database"
	"github.com/soumajitgh/mobicode/internal/setup"
	"github.com/spf13/cobra"
)

var resetIdentityConfirm bool

var identityCmd = &cobra.Command{Use: "identity", Short: "Manage the single owner identity"}

var resetIdentityCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear the paired owner identity and reopen browser setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !resetIdentityConfirm {
			return fmt.Errorf("refusing to reset identity without --confirm")
		}
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		db, err := database.Open(cfg.DatabasePath)
		if err != nil {
			return err
		}
		owner := auth.NewOwnerService(auth.NewOwnerRepository(db))
		service := setup.NewService(setup.NewRepository(db), owner)
		if err := service.Reset(context.Background()); err != nil {
			return fmt.Errorf("reset identity: %w", err)
		}
		cmd.Println("Owner identity cleared. Start the server, then visit /setup to pair a mobile.")
		return nil
	},
}

func init() {
	resetIdentityCmd.Flags().BoolVar(&resetIdentityConfirm, "confirm", false, "confirm identity reset")
	identityCmd.AddCommand(resetIdentityCmd)
	rootCmd.AddCommand(identityCmd)
}
