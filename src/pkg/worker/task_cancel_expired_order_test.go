package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
)

type mockOrderManager struct {
	updateOrderStatusFunc func(ctx context.Context, id string, status string) error
	expireOrderFunc       func(ctx context.Context, id string) error
}

func (m *mockOrderManager) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	if m.updateOrderStatusFunc != nil {
		return m.updateOrderStatusFunc(ctx, id, status)
	}
	return nil
}

func (m *mockOrderManager) ExpireOrder(ctx context.Context, id string) error {
	if m.expireOrderFunc != nil {
		return m.expireOrderFunc(ctx, id)
	}
	return nil
}

func TestProcessTaskCancelExpiredOrder(t *testing.T) {
	tests := []struct {
		name    string
		payload interface{}
		mockErr error
		wantErr bool
	}{
		{
			name: "successful cancellation",
			payload: PayloadCancelExpiredOrder{
				OrderID: "order-123",
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "cancellation failure",
			payload: PayloadCancelExpiredOrder{
				OrderID: "order-123",
			},
			mockErr: errors.New("db error"),
			wantErr: true,
		},
		{
			name:    "invalid payload",
			payload: "invalid-string-not-json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payloadData []byte
			if str, ok := tt.payload.(string); ok {
				payloadData = []byte(str)
			} else {
				payloadData, _ = json.Marshal(tt.payload)
			}

			task := asynq.NewTask(TaskCancelExpiredOrder, payloadData)

			orderMgr := &mockOrderManager{
				expireOrderFunc: func(ctx context.Context, id string) error {
					return tt.mockErr
				},
			}

			processor := &RedisTaskProcessor{
				orderManager: orderMgr,
			}

			err := processor.ProcessTaskCancelExpiredOrder(context.Background(), task)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessTaskCancelExpiredOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
