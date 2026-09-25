package cache

import (
	"context"
	"testing"
	"time"

	"emc_lb/src/pkg/entities"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductCacheStore(t *testing.T) {
	ctx := context.Background()

	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := NewRedisProductCacheStore(client, 15*time.Minute)

	productID := "prod-123"
	expectedProduct := entities.ProductResponse{
		ID:   productID,
		Name: "Test Product",
	}

	t.Run("Cache MISS", func(t *testing.T) {
		mr.FlushAll()
		_, err := store.GetByID(ctx, productID)
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("Cache SET and HIT", func(t *testing.T) {
		err := store.SetByID(ctx, productID, expectedProduct)
		assert.NoError(t, err)

		cached, err := store.GetByID(ctx, productID)
		assert.NoError(t, err)
		assert.Equal(t, expectedProduct.ID, cached.ID)
		assert.Equal(t, expectedProduct.Name, cached.Name)
	})

	t.Run("Invalidate", func(t *testing.T) {
		err := store.Invalidate(ctx, productID)
		assert.NoError(t, err)

		_, err = store.GetByID(ctx, productID)
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("Invalid Payload", func(t *testing.T) {
		errSet := mr.Set(buildProductItemKey(productID), "{invalid json}")
		require.NoError(t, errSet)

		_, err := store.GetByID(ctx, productID)
		assert.Error(t, err)

		// ensure it got deleted
		exists := mr.Exists(buildProductItemKey(productID))
		assert.False(t, exists)
	})

	t.Run("Redis Unavailable", func(t *testing.T) {
		mr.Close() // simulate failure
		
		err := store.SetByID(ctx, productID, expectedProduct)
		assert.Error(t, err)

		_, err = store.GetByID(ctx, productID)
		assert.Error(t, err)
	})
}
