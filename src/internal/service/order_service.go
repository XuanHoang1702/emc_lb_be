package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) (entities.OrderResponse, error)
	ListOrders(ctx context.Context, userID string) ([]entities.OrderResponse, error)
	GetOrder(ctx context.Context, id string) (entities.OrderResponse, error)
	MarkAsPaidByInvoice(ctx context.Context, invoiceNumber string) error
	UpdateOrderStatus(ctx context.Context, id string, status string) error
}

type orderService struct {
	orderRepository   repository.OrderRepository
	productRepository repository.ProductRepository
}

func NewOrderService(orderRepository repository.OrderRepository, productRepository repository.ProductRepository) OrderService {
	return &orderService{
		orderRepository:   orderRepository,
		productRepository: productRepository,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) (entities.OrderResponse, error) {
	if len(req.Items) == 0 {
		return entities.OrderResponse{}, &res.AppError{
			Message:    "Order must have at least one item",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	var totalAmount float64
	var items []entities.OrderItem

	for _, item := range req.Items {
		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			return entities.OrderResponse{}, &res.AppError{
				Message:    fmt.Sprintf("Product %s not found", item.ProductID),
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}

		if product.Stock < item.Quantity && !product.AllowBackorder {
			return entities.OrderResponse{}, &res.AppError{
				Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}

		// Deduct stock
		if err := s.productRepository.UpdateStock(ctx, item.ProductID, -item.Quantity, item.Quantity); err != nil {
			return entities.OrderResponse{}, res.WrapError(err, "Failed to update stock", erres.CommonInternal)
		}

		// Use the actual current price of the product
		itemPrice := product.Price
		totalAmount += itemPrice * float64(item.Quantity)

		items = append(items, entities.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     itemPrice,
		})
	}

	invoiceNumber := fmt.Sprintf("INV-%s", uuid.New().String()[:8])
	now := time.Now().UTC()

	order := entities.Order{
		UserID:          userID,
		InvoiceNumber:   invoiceNumber,
		Items:           items,
		TotalAmount:     totalAmount,
		Status:          "pending",
		PaymentStatus:   "unpaid",
		PaymentMethod:   req.PaymentMethod,
		ShippingAddress: req.ShippingAddress,
		ContactPhone:    req.ContactPhone,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	createdOrder, err := s.orderRepository.Create(ctx, order)
	if err != nil {
		return entities.OrderResponse{}, res.WrapError(err, "Can not create order", erres.CommonInternal)
	}

	return mapping.ToOrderResponse(createdOrder), nil
}

func (s *orderService) ListOrders(ctx context.Context, userID string) ([]entities.OrderResponse, error) {
	orders, err := s.orderRepository.List(ctx, userID)
	if err != nil {
		return nil, res.WrapError(err, "Can not list orders", erres.CommonInternal)
	}

	var responses []entities.OrderResponse
	for _, order := range orders {
		responses = append(responses, mapping.ToOrderResponse(order))
	}
	return responses, nil
}

func (s *orderService) GetOrder(ctx context.Context, id string) (entities.OrderResponse, error) {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return entities.OrderResponse{}, res.WrapError(err, "Order not found", erres.CommonNotFound)
	}
	return mapping.ToOrderResponse(order), nil
}

func (s *orderService) MarkAsPaidByInvoice(ctx context.Context, invoiceNumber string) error {
	// Find order by invoice number
	// For simplicity in this scaffold, let's assume we list and filter by invoice number.
	// In production, you would add an index and a repository method GetByInvoiceNumber
	orders, err := s.orderRepository.List(ctx, "")
	if err != nil {
		return err
	}

	for _, order := range orders {
		if order.InvoiceNumber == invoiceNumber {
			return s.orderRepository.UpdatePaymentStatus(ctx, order.ID, "paid")
		}
	}
	return fmt.Errorf("order with invoice number %s not found", invoiceNumber)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	validStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
		"shipped":    true,
		"delivered":  true,
		"cancelled":  true,
	}

	if !validStatuses[status] {
		return &res.AppError{
			Message:    "Invalid order status",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	// If order is already cancelled, we shouldn't change it back or process it further
	if order.Status == "cancelled" {
		return &res.AppError{
			Message:    "Order is already cancelled",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if err := s.orderRepository.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	// If transitioning to cancelled, restore stock
	if status == "cancelled" {
		for _, item := range order.Items {
			_ = s.productRepository.UpdateStock(ctx, item.ProductID, item.Quantity, -item.Quantity)
		}
	}

	return nil
}
