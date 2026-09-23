package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
)

func (processor *RedisTaskProcessor) ProcessTaskSendOrderPaymentSuccessEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendOrderPaymentSuccessEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	err := processor.mailer.SendOrderPaymentSuccessEmail(ctx, payload.CustomerEmail, payload.CustomerName, payload.InvoiceNumber, payload.AmountPaid)
	if err != nil {
		return fmt.Errorf("failed to send order payment email: %w", err)
	}

	return nil
}
