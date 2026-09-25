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

func TestCategoryCacheStore(t *testing.T) {
	ctx := context.Background()

	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	store := NewRedisCategoryCacheStore(client, 24*time.Hour)

	expectedCategories := []entities.CategoryResponse{
		{ID: "cat-1", Name: "Category 1"},
		{ID: "cat-2", Name: "Category 2"},
	}

	t.Run("Cache MISS", func(t *testing.T) {
		mr.FlushAll()
		_, err := store.GetList(ctx)
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("Cache SET and HIT", func(t *testing.T) {
		err := store.SetList(ctx, expectedCategories)
		assert.NoError(t, err)

		cached, err := store.GetList(ctx)
		assert.NoError(t, err)
		assert.Len(t, cached, 2)
		assert.Equal(t, expectedCategories[0].Name, cached[0].Name)
	})

	t.Run("Invalidate", func(t *testing.T) {
		err := store.InvalidateList(ctx)
		assert.NoError(t, err)

		_, err = store.GetList(ctx)
		assert.ErrorIs(t, err, redis.Nil)
	})

	t.Run("Redis Unavailable", func(t *testing.T) {
		mr.Close() // simulate failure
		
		err := store.SetList(ctx, expectedCategories)
		assert.Error(t, err)

		_, err = store.GetList(ctx)
		assert.Error(t, err)
	})
}
