package mapping

import "emc_lb/src/pkg/entities"

func ToOrderResponse(order entities.Order) entities.OrderResponse {
	return entities.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		InvoiceNumber:   order.InvoiceNumber,
		Items:           order.Items,
		TotalAmount:     order.TotalAmount,
		Status:          order.Status,
		PaymentStatus:   order.PaymentStatus,
		PaymentMethod:   order.PaymentMethod,
		ShippingAddress: order.ShippingAddress,
		ContactPhone:    order.ContactPhone,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}
