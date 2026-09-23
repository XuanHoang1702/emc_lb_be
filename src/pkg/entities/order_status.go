package entities

// Order status constants.
const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"

	PaymentStatusUnpaid   = "unpaid"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// IsTerminalStatus returns true if the order status is a final state
func IsTerminalStatus(status string) bool {
	return status == OrderStatusDelivered || status == OrderStatusCancelled
}

// ValidTransitions defines allowed order status transitions.
// Terminal states (delivered, cancelled) have no valid next states.
var ValidTransitions = map[string][]string{
	OrderStatusPending:    {OrderStatusProcessing, OrderStatusCancelled},
	OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
	OrderStatusShipped:    {OrderStatusDelivered},
	OrderStatusDelivered:  {}, // terminal
	OrderStatusCancelled:  {}, // terminal
}

// CanTransition checks if transitioning from current to next is a valid order state change.
func CanTransition(current, next string) bool {
	allowed, ok := ValidTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// ValidPaymentTransitions defines allowed payment status transitions.
var ValidPaymentTransitions = map[string][]string{
	PaymentStatusUnpaid:   {PaymentStatusPaid, PaymentStatusFailed},
	PaymentStatusFailed:   {PaymentStatusPaid}, // Allow a retry to succeed
	PaymentStatusPaid:     {PaymentStatusRefunded}, // Paid is terminal except for refunds
	PaymentStatusRefunded: {}, // terminal
}

// CanTransitionPayment checks if transitioning from current to next is a valid payment state change.
func CanTransitionPayment(current, next string) bool {
	allowed, ok := ValidPaymentTransitions[current]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}
