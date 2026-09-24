package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var reserveStockScript = redis.NewScript(`
local stock_key = KEYS[1]
local qty = tonumber(ARGV[1])
local current = redis.call('GET', stock_key)

if current == false then
	return -1 -- Not found in Redis (cache miss)
end

if tonumber(current) >= qty then
	redis.call('DECRBY', stock_key, qty)
	return 1 -- Success
else
	return -2 -- Insufficient stock
end
`)

type InventoryService interface {
	// ReserveStock atomically deducts stock in Redis using a LUA script.
	// Returns a boolean indicating success, and an error if it was a system error or cache miss.
	ReserveStock(ctx context.Context, productID string, quantity int64) (bool, error)

	// RestoreStock atomically increments stock in Redis (useful for rollbacks/cancellations).
	RestoreStock(ctx context.Context, productID string, quantity int64) error

	// SyncStockToRedis is used to warm up the cache from MongoDB to Redis.
	SyncStockToRedis(ctx context.Context, productID string, stock int64) error
}

type inventoryService struct {
	redisClient *redis.Client
}

func NewInventoryService(redisClient *redis.Client) InventoryService {
	return &inventoryService{
		redisClient: redisClient,
	}
}

func (s *inventoryService) getStockKey(productID string) string {
	return fmt.Sprintf("product_stock:%s", productID)
}

func (s *inventoryService) ReserveStock(ctx context.Context, productID string, quantity int64) (bool, error) {
	key := s.getStockKey(productID)
	res, err := reserveStockScript.Run(ctx, s.redisClient, []string{key}, quantity).Int()
	if err != nil {
		return false, fmt.Errorf("failed to execute reserve script: %w", err)
	}

	switch res {
	case 1:
		return true, nil // Success
	case -1:
		return false, fmt.Errorf("stock not found in redis for product %s", productID)
	case -2:
		return false, nil // Insufficient stock
	default:
		return false, fmt.Errorf("unknown response from lua script: %d", res)
	}
}

func (s *inventoryService) RestoreStock(ctx context.Context, productID string, quantity int64) error {
	key := s.getStockKey(productID)
	// If the key exists, INCRBY. If it doesn't exist, we might not want to create it,
	// but INCRBY creates it if it doesn't exist. To avoid creating ghost keys, we could check if it exists.
	// For simplicity, we just use INCRBY. If it's a ghost key, it will just have a value and might be overwritten by sync.
	// Better approach:
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return s.redisClient.IncrBy(ctx, key, quantity).Err()
	}
	// If it doesn't exist in Redis, it means it wasn't cached anyway. Do nothing.
	return nil
}

func (s *inventoryService) SyncStockToRedis(ctx context.Context, productID string, stock int64) error {
	key := s.getStockKey(productID)
	return s.redisClient.Set(ctx, key, stock, 0).Err() // Set with no expiration for active products
}
