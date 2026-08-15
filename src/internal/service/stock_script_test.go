package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestRedis spins up a miniredis instance and returns a client pointing to it.
func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

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

func runStockScript(t *testing.T, client *redis.Client, key string, qty int) (int, error) {
	t.Helper()
	return reserveStockScript.Run(context.Background(), client, []string{key}, qty).Int()
}

func assertScriptResult(t *testing.T, res, expected int) {
	t.Helper()
	if res != expected {
		t.Fatalf("expected %d, got %d", expected, res)
	}
}

func TestReserveStockScript_MissingKey(t *testing.T) {
	client := newTestRedis(t)

	res, err := runStockScript(t, client, "product_stock:p1", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScriptResult(t, res, -1)
}

func TestReserveStockScript_SufficientStock(t *testing.T) {
	client := newTestRedis(t)
	key := "product_stock:p2"
	if err := client.Set(context.Background(), key, 10, 0).Err(); err != nil {
		t.Fatalf("set stock: %v", err)
	}

	res, err := runStockScript(t, client, key, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScriptResult(t, res, 1)

	remaining := client.Get(context.Background(), key).Val()
	if remaining != "6" {
		t.Fatalf("expected remaining stock 6, got %s", remaining)
	}
}

func TestReserveStockScript_InsufficientStock(t *testing.T) {
	client := newTestRedis(t)
	key := "product_stock:p3"
	if err := client.Set(context.Background(), key, 3, 0).Err(); err != nil {
		t.Fatalf("set stock: %v", err)
	}

	res, err := runStockScript(t, client, key, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScriptResult(t, res, -2)

	// Stock must not be mutated on failure.
	remaining := client.Get(context.Background(), key).Val()
	if remaining != "3" {
		t.Fatalf("expected unchanged stock 3, got %s", remaining)
	}
}

func TestReserveStockScript_ExactStock(t *testing.T) {
	client := newTestRedis(t)
	key := "product_stock:p4"
	if err := client.Set(context.Background(), key, 5, 0).Err(); err != nil {
		t.Fatalf("set stock: %v", err)
	}

	res, err := runStockScript(t, client, key, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertScriptResult(t, res, 1)

	remaining := client.Get(context.Background(), key).Val()
	if remaining != "0" {
		t.Fatalf("expected remaining stock 0, got %s", remaining)
	}
}
