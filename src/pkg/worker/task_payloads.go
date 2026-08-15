package worker

const (
	TaskSendVerifyEmail = "task:send_verify_email"
)

type PayloadSendVerifyEmail struct {
	Email    string `json:"email"`
	UserName string `json:"user_name"`
	OTP      string `json:"otp"`
	TTL      int    `json:"ttl"`
}
