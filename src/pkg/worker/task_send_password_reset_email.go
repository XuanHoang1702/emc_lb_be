package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

func (processor *RedisTaskProcessor) ProcessTaskSendPasswordResetEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendPasswordResetEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w, asynq.SkipRetry", err)
	}

	err := processor.mailer.SendPasswordResetOTP(
		ctx,
		payload.Email,
		payload.UserName,
		payload.OTP,
		payload.TTL,
	)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}
