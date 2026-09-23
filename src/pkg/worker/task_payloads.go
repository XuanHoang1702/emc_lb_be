package worker

const (
	TaskSendVerifyEmail    = "task:send_verify_email"
	TaskCancelExpiredOrder = "task:cancel_expired_order"
)

type PayloadSendVerifyEmail struct {
	Email    string `json:"email"`
	UserName string `json:"user_name"`
	OTP      string `json:"otp"`
	TTL      int    `json:"ttl"`
}

type PayloadCancelExpiredOrder struct {
	OrderID string `json:"order_id"`
}
