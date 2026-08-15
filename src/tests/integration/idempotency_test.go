//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// TestDuplicatePaymentWebhook simulates receiving the same payment webhook twice.
// A common approach is caching the processed transaction ID in Redis or checking DB status.
func TestDuplicatePaymentWebhook(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer client.Close()

	ctx := context.Background()
	transactionID := "txn_sepay_123456"
	cacheKey := "processed_txn:" + transactionID

	// First Webhook Delivery
	// SetNX (Set if Not eXists) is a classic pattern for idempotency
	isNew, err := client.SetNX(ctx, cacheKey, "processed", 24*time.Hour).Result()
	if err != nil {
		t.Fatalf("first webhook failed: %v", err)
	}
	if !isNew {
		t.Errorf("First webhook should be processed as new")
	}

	// Processing logic here (e.g. update order to PAID)

	// Second Webhook Delivery (Duplicate)
	isNewDuplicate, err := client.SetNX(ctx, cacheKey, "processed", 24*time.Hour).Result()
	if err != nil {
		t.Fatalf("second webhook failed: %v", err)
	}
	if isNewDuplicate {
		t.Errorf("Second webhook should be detected as duplicate")
	}

	// Logic should gracefully ignore the duplicate since isNewDuplicate == false
}

// TestDuplicateOrderRequest verifies client-provided idempotency keys prevent duplicate orders.
func TestDuplicateOrderRequest(t *testing.T) {
	s := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer client.Close()

	ctx := context.Background()
	idempotencyKey := "idem_key_uuid_87654321"
	cacheKey := "idem_order:" + idempotencyKey

	// Attempt 1
	attempt1, _ := client.SetNX(ctx, cacheKey, "creating", 10*time.Minute).Result()
	if !attempt1 {
		t.Fatalf("First request should acquire the lock")
	}

	// Attempt 2 (while Attempt 1 is processing or finished)
	attempt2, _ := client.SetNX(ctx, cacheKey, "creating", 10*time.Minute).Result()
	if attempt2 {
		t.Fatalf("Second request with same idempotency key should be blocked")
	}
}
