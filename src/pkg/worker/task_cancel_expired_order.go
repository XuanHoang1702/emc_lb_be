package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"emc_lb/src/pkg/logs"
	"github.com/hibiken/asynq"
)

func (processor *RedisTaskProcessor) ProcessTaskCancelExpiredOrder(ctx context.Context, task *asynq.Task) error {
	var payload PayloadCancelExpiredOrder
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	logs.WithContext(ctx).Info("Processing expired order cancellation task", "order_id", payload.OrderID)

	// Update order status to "cancelled".
	// The OrderService (via OrderManager interface) will handle stock restoration.
	err := processor.orderManager.UpdateOrderStatus(ctx, payload.OrderID, "cancelled")
	if err != nil {
		// Log the error. If the order was already paid or cancelled, the service will return an error.
		// We might not want to retry if the state transition is invalid.
		logs.WithContext(ctx).Warn("Failed to cancel expired order", "order_id", payload.OrderID, "error", err)
		return nil // Don't retry if it fails due to state rules
	}

	logs.WithContext(ctx).Info("Successfully cancelled expired order", "order_id", payload.OrderID)
	return nil
}
