package entities

// Order status constants.
const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

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
