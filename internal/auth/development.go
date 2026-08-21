package auth

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/soumajitgh/mobicode/internal/config"
)

// BootstrapDevelopmentIdentity configures the configured DEV_NSEC as owner on a fresh dev database.
func BootstrapDevelopmentIdentity(cfg *config.Config, owner *OwnerService, log *zap.Logger) error {
	if cfg.DevNsec == "" {
		return nil
	}
	publicKey, err := PublicKeyFromNsec(cfg.DevNsec)
	if err != nil {
		return err
	}
	existing, err := owner.Owner(context.Background())
	if errors.Is(err, ErrSetupRequired) {
		if err := owner.Configure(context.Background(), publicKey); err != nil {
			return fmt.Errorf("configure development identity: %w", err)
		}
		log.Warn("development Nostr identity configured", zap.String("public_key", publicKey))
		return nil
	}
	if err != nil {
		return err
	}
	if existing.PublicKey != publicKey {
		return fmt.Errorf("configured owner does not match DEV_NSEC; run `mobicode identity reset --confirm` against the development database")
	}
	return nil
}
