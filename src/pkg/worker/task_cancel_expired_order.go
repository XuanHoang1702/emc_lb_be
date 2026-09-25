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

	err := processor.orderManager.ExpireOrder(ctx, payload.OrderID)
	if err != nil {
		logs.WithContext(ctx).Warn("Failed to expire order", "order_id", payload.OrderID, "error", err)
		OrderCancellationFailure.Inc()
		return err // Retry if it fails due to DB issue, state rules are handled inside
	}

	OrderCancellationSuccess.Inc()
	logs.WithContext(ctx).Info("Successfully cancelled expired order", "order_id", payload.OrderID)
	return nil
}
