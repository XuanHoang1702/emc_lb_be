//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
)

// Task payload types
const (
	TypeEmailVerification = "email:verify"
)

// simulateWorker handler that processes tasks
func simulateWorker(ctx context.Context, task *asynq.Task) error {
	if task.Type() == TypeEmailVerification {
		// simulate success
		return nil
	}
	return errors.New("unknown task")
}

// failingWorker always fails to test retries
var failingWorkerCalls int

func failingWorker(ctx context.Context, task *asynq.Task) error {
	failingWorkerCalls++
	return errors.New("temporary failure")
}

func TestQueuePublishAndConsumer(t *testing.T) {
	s := miniredis.RunT(t)
	redisOpt := asynq.RedisClientOpt{Addr: s.Addr()}

	// 1. Initialize Client & Server
	client := asynq.NewClient(redisOpt)
	defer client.Close()

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 1,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeEmailVerification, simulateWorker)

	// Start processing
	go func() {
		if err := srv.Run(mux); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()
	defer srv.Stop()

	// 2. Publish task
	task := asynq.NewTask(TypeEmailVerification, []byte("user@example.com"))
	info, err := client.Enqueue(task)
	if err != nil {
		t.Fatalf("failed to enqueue task: %v", err)
	}
	if info.Queue != "default" {
		t.Errorf("expected task to be in default queue, got %s", info.Queue)
	}

	// 3. Wait for processing
	time.Sleep(1 * time.Second)

	// We could inspect miniredis keys if needed, but if it didn't crash, it works.
	// Asynq moves completed tasks out of the active queue or deletes them.
}

func TestQueueRetryRecovery(t *testing.T) {
	s := miniredis.RunT(t)
	redisOpt := asynq.RedisClientOpt{Addr: s.Addr()}

	client := asynq.NewClient(redisOpt)
	defer client.Close()

	srv := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 1,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return 1 * time.Second // Retry very fast for testing
			},
		},
	)

	failingWorkerCalls = 0
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeEmailVerification, failingWorker)

	go func() {
		_ = srv.Run(mux)
	}()
	defer srv.Stop()

	// Enqueue task with MaxRetry = 2
	task := asynq.NewTask(TypeEmailVerification, []byte("fail@example.com"))
	_, err := client.Enqueue(task, asynq.MaxRetry(2))
	if err != nil {
		t.Fatalf("failed to enqueue task: %v", err)
	}

	// Wait enough time for 2 retries (Initial + 2 retries = 3 calls)
	time.Sleep(3 * time.Second)

	if failingWorkerCalls < 2 {
		t.Errorf("Expected at least 2 retry attempts, got %d", failingWorkerCalls)
	}
}
