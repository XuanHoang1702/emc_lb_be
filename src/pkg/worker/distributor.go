package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendVerifyEmail(
		ctx context.Context,
		payload *PayloadSendVerifyEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskSendPasswordResetEmail(
		ctx context.Context,
		payload *PayloadSendPasswordResetEmail,
		opts ...asynq.Option,
	) error
	DistributeTaskCancelExpiredOrder(
		ctx context.Context,
		payload *PayloadCancelExpiredOrder,
		opts ...asynq.Option,
	) error
	DistributeTaskSendOrderPaymentSuccessEmail(
		ctx context.Context,
		payload *PayloadSendOrderPaymentSuccessEmail,
		opts ...asynq.Option,
	) error
}

type RedisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RedisTaskDistributor{
		client: client,
	}
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	_, err = distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (distributor *RedisTaskDistributor) DistributeTaskSendPasswordResetEmail(
	ctx context.Context,
	payload *PayloadSendPasswordResetEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendPasswordResetEmail, jsonPayload, opts...)

	_, err = distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (distributor *RedisTaskDistributor) DistributeTaskCancelExpiredOrder(
	ctx context.Context,
	payload *PayloadCancelExpiredOrder,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskCancelExpiredOrder, jsonPayload, opts...)

	_, err = distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (distributor *RedisTaskDistributor) DistributeTaskSendOrderPaymentSuccessEmail(
	ctx context.Context,
	payload *PayloadSendOrderPaymentSuccessEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendOrderPaymentSuccessEmail, jsonPayload, opts...)

	_, err = distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}
