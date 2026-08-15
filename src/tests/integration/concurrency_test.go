//go:build integration

package integration

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// The actual script used in order_service (copied here for testing the concurrency logic)
var reserveStockScript = redis.NewScript(`
local stock_key = KEYS[1]
local qty = tonumber(ARGV[1])
local current = redis.call('GET', stock_key)

if current == false then
	return -1 -- Not found in Redis
end

if tonumber(current) >= qty then
	redis.call('DECRBY', stock_key, qty)
	return 1 -- Success
else
	return -2 -- Insufficient stock
end
`)

func runStockScript(ctx context.Context, client *redis.Client, key string, qty int) (int, error) {
	return reserveStockScript.Run(ctx, client, []string{key}, qty).Int()
}

// TestConcurrentStockReservation verifies that under high concurrent load,
// the stock reservation script never oversells products.
func TestConcurrentStockReservation(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer client.Close()

	ctx := context.Background()
	productKey := "product_stock:test_prod_1"

	// Initial stock is exactly 5
	err := client.Set(ctx, productKey, 5, 0).Err()
	if err != nil {
		t.Fatalf("failed to set stock: %v", err)
	}

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32

	// 100 concurrent requests trying to buy 1 item each
	numRequests := 100
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := runStockScript(ctx, client, productKey, 1)
			if err != nil {
				return
			}
			if res == 1 {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}()
	}
	wg.Wait()

	if successCount != 5 {
		t.Errorf("Expected exactly 5 successful reservations, got %d", successCount)
	}
	if failCount != int32(numRequests-5) {
		t.Errorf("Expected exactly %d failed reservations, got %d", numRequests-5, failCount)
	}

	// Verify remaining stock is 0
	remaining := client.Get(ctx, productKey).Val()
	if remaining != "0" {
		t.Errorf("Expected 0 stock remaining, got %s", remaining)
	}
}

// TestConcurrentCouponUsage simulates a coupon with a max usage limit being applied concurrently
func TestConcurrentCouponUsage(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer client.Close()

	ctx := context.Background()
	couponKey := "coupon_usage:SPRING2026"
	usageLimit := 10

	var wg sync.WaitGroup
	var successCount int32

	// 50 users trying to apply a coupon that is limited to 10 uses
	numRequests := 50
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate an INCR transaction or atomic check for coupon limits
			// E.g., usage = INCR coupon_key, if usage > limit then DECR coupon_key and fail.
			usage, err := client.Incr(ctx, couponKey).Result()
			if err != nil {
				return
			}

			if usage <= int64(usageLimit) {
				atomic.AddInt32(&successCount, 1)
			} else {
				// Revert if over limit
				client.Decr(ctx, couponKey)
			}
		}()
	}
	wg.Wait()

	if successCount != int32(usageLimit) {
		t.Errorf("Expected exactly %d successful coupon usages, got %d", usageLimit, successCount)
	}
}
