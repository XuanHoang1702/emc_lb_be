package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	err := processor.mailer.SendEmailVerificationOTP(
		ctx,
		payload.Email,
		payload.UserName,
		payload.OTP,
		payload.TTL,
	)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	return nil
}
