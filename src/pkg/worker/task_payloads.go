package worker

const (
	TaskSendVerifyEmail                  = "task:send_verify_email"
	TaskCancelExpiredOrder               = "task:cancel_expired_order"
	TaskSendOrderPaymentSuccessEmail     = "task:send_order_payment_success_email"
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

type PayloadSendOrderPaymentSuccessEmail struct {
	InvoiceNumber string  `json:"invoice_number"`
	CustomerEmail string  `json:"customer_email"`
	CustomerName  string  `json:"customer_name"`
	AmountPaid    float64 `json:"amount_paid"`
}
