package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenStore_SaveAndGetSession(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := NewRedisRefreshTokenStore(client)

	ctx := context.Background()
	session := SessionData{UserID: "user-1", SessionVersion: 2}
	tokenA := "token-A"

	// 1. Save token
	err = store.Save(ctx, session, tokenA, 1*time.Hour)
	assert.NoError(t, err)

	// 2. Get session
	retrievedSession, err := store.GetSession(ctx, tokenA)
	assert.NoError(t, err)
	assert.Equal(t, session.UserID, retrievedSession.UserID)
	assert.Equal(t, session.SessionVersion, retrievedSession.SessionVersion)

	// 3. Delete token
	err = store.Delete(ctx, tokenA)
	assert.NoError(t, err)

	// 4. Get after delete
	_, err = store.GetSession(ctx, tokenA)
	assert.Error(t, err)
}

func TestRefreshTokenStore_FallbackOldTokens(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := NewRedisRefreshTokenStore(client)

	ctx := context.Background()
	token := "token-old"

	// Manually insert an old-format string (just UserID)
	client.Set(ctx, buildRefreshTokenKey(token), "user-legacy", 1*time.Hour)

	retrievedSession, err := store.GetSession(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, "user-legacy", retrievedSession.UserID)
	assert.Equal(t, int32(1), retrievedSession.SessionVersion)
}
