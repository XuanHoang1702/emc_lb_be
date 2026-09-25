package worker

import (
	"context"

	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mail"

	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
}

type OrderManager interface {
	UpdateOrderStatus(ctx context.Context, id string, status string) error
	ExpireOrder(ctx context.Context, id string) error
}

type RedisTaskProcessor struct {
	server       *asynq.Server
	mailer       mail.Mailer
	orderManager OrderManager
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, mailer mail.Mailer, orderManager OrderManager) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				logs.LogError("worker", "process_task_failed", err, map[string]any{
					"task_type": task.Type(),
					"payload":   string(task.Payload()),
				})
			}),
			Logger: logs.NewAsynqLogger(),
		},
	)

	return &RedisTaskProcessor{
		server:       server,
		mailer:       mailer,
		orderManager: orderManager,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()

	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	mux.HandleFunc(TaskSendPasswordResetEmail, processor.ProcessTaskSendPasswordResetEmail)
	mux.HandleFunc(TaskCancelExpiredOrder, processor.ProcessTaskCancelExpiredOrder)
	mux.HandleFunc(TaskSendOrderPaymentSuccessEmail, processor.ProcessTaskSendOrderPaymentSuccessEmail)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) Shutdown() {
	processor.server.Shutdown()
}
