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

func TestBrandCacheStore(t *testing.T) {
	ctx := context.Background()

	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := NewRedisBrandCacheStore(client, 24*time.Hour)

	expectedBrands := []entities.BrandResponse{
		{ID: "brand-1", Name: "Brand 1"},
		{ID: "brand-2", Name: "Brand 2"},
	}

	t.Run("Cache MISS", func(t *testing.T) {
		mr.FlushAll()
		_, err := store.GetAll(ctx)
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("Cache SET and HIT", func(t *testing.T) {
		err := store.SetAll(ctx, expectedBrands)
		assert.NoError(t, err)

		cached, err := store.GetAll(ctx)
		assert.NoError(t, err)
		assert.Len(t, cached, 2)
		assert.Equal(t, expectedBrands[0].Name, cached[0].Name)
	})

	t.Run("Invalidate", func(t *testing.T) {
		err := store.InvalidateAll(ctx)
		assert.NoError(t, err)

		_, err = store.GetAll(ctx)
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("Redis Unavailable", func(t *testing.T) {
		mr.Close() // simulate failure
		
		err := store.SetAll(ctx, expectedBrands)
		assert.Error(t, err)

		_, err = store.GetAll(ctx)
		assert.Error(t, err)
	})
}
