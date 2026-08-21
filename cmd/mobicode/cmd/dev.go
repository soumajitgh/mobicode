package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/config"
)

var (
	devRequestURL    string
	devRequestBody   string
	devRequestMethod string
	devNsecValueOnly bool
)

var devCmd = &cobra.Command{Use: "dev", Short: "Development-only authentication helpers"}

var signRequestCmd = &cobra.Command{
	Use:   "sign-request",
	Short: "Print a NIP-98 Authorization header for an exact request body",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg.Env != "development" || cfg.DevNsec == "" {
			return fmt.Errorf("sign-request requires ENV=development and DEV_NSEC")
		}
		requestURL := strings.TrimSpace(devRequestURL)
		if requestURL == "" {
			requestURL = cfg.PublicBaseURL + "/graphql"
		}
		if devRequestBody == "" {
			return fmt.Errorf("--body is required because NIP-98 signs the exact request body")
		}
		header, err := auth.CreateNIP98Authorization(cfg.DevNsec, requestURL, devRequestMethod, devRequestBody, time.Now())
		if err != nil {
			return err
		}
		cmd.Println("Authorization: " + header)
		return nil
	},
}

var generateNsecCmd = &cobra.Command{
	Use:   "generate-nsec",
	Short: "Generate an intentionally local-only Nostr development key",
	RunE: func(cmd *cobra.Command, args []string) error {
		nsec, publicKey, err := auth.GenerateNsec()
		if err != nil {
			return err
		}
		if devNsecValueOnly {
			cmd.Println(nsec)
			return nil
		}
		cmd.Printf("DEV_NSEC=%s\n", nsec)
		cmd.Printf("EXPO_PUBLIC_DEV_NSEC=%s\n", nsec)
		cmd.Printf("Public key: %s\n", publicKey)
		return nil
	},
}

func init() {
	signRequestCmd.Flags().StringVar(&devRequestURL, "url", "", "absolute request URL (defaults to PUBLIC_BASE_URL/graphql)")
	signRequestCmd.Flags().StringVar(&devRequestMethod, "method", "POST", "HTTP method")
	signRequestCmd.Flags().StringVar(&devRequestBody, "body", "", "exact raw request body")
	generateNsecCmd.Flags().BoolVar(&devNsecValueOnly, "value-only", false, "print only the nsec value")
	devCmd.AddCommand(signRequestCmd)
	devCmd.AddCommand(generateNsecCmd)
	rootCmd.AddCommand(devCmd)
}
