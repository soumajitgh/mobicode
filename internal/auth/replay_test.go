package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/soumajitgh/mobicode/internal/auth"
	"github.com/soumajitgh/mobicode/internal/testutil"
)

func TestOwnerService_ClaimReplay(t *testing.T) {
	repo := auth.NewOwnerRepository(testutil.NewDB(t))
	owner := auth.NewOwnerService(repo)
	expiresAt := time.Now().Add(time.Minute)

	require.NoError(t, owner.ClaimReplay(context.Background(), "first-event", expiresAt))
	require.ErrorIs(t, owner.ClaimReplay(context.Background(), "first-event", expiresAt), auth.ErrReplay)
	require.NoError(t, owner.ClaimReplay(context.Background(), "different-event", expiresAt))
}

func TestOwnerService_ClaimReplayPrunesExpiredEntries(t *testing.T) {
	db := testutil.NewDB(t)
	owner := auth.NewOwnerService(auth.NewOwnerRepository(db))
	require.NoError(t, owner.ClaimReplay(context.Background(), "expired-event", time.Now().Add(-time.Minute)))
	require.NoError(t, owner.ClaimReplay(context.Background(), "fresh-event", time.Now().Add(time.Minute)))

	var count int64
	require.NoError(t, db.Model(&auth.AuthReplay{}).Where("event_id = ?", "expired-event").Count(&count).Error)
	require.EqualValues(t, 0, count)
}
