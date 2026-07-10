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
	CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) ([]entities.OrderResponse, error)
	ListOrders(ctx context.Context, userID string) ([]entities.OrderResponse, error)
	GetOrder(ctx context.Context, id string) (entities.OrderResponse, error)
	MarkAsPaidByInvoice(ctx context.Context, invoiceNumber string) error
	UpdateOrderStatus(ctx context.Context, id string, status string) error
}

type orderService struct {
	orderRepository   repository.OrderRepository
	productRepository repository.ProductRepository
	couponService     CouponService
}

func NewOrderService(orderRepository repository.OrderRepository, productRepository repository.ProductRepository, couponSvc CouponService) OrderService {
	return &orderService{
		orderRepository:   orderRepository,
		productRepository: productRepository,
		couponService:     couponSvc,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) ([]entities.OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, &res.AppError{
			Message:    "Order must have at least one item",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	var deductedItems []entities.OrderItem
	rollback := func() {
		for _, di := range deductedItems {
			_ = s.productRepository.UpdateStock(context.Background(), di.ProductID, di.Quantity, -di.Quantity)
		}
	}

	// Group items by ShopID
	shopItems := make(map[string][]entities.OrderItem)
	shopTotals := make(map[string]float64)

	for _, item := range req.Items {
		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			rollback()
			return nil, &res.AppError{
				Message:    fmt.Sprintf("Product %s not found", item.ProductID),
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}

		if product.Stock < item.Quantity && !product.AllowBackorder {
			rollback()
			return nil, &res.AppError{
				Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}

		// Deduct stock (Optimistic lock ideally, but here we just update MongoDB. Redis lock is recommended for high concurrency)
		if err := s.productRepository.UpdateStock(ctx, item.ProductID, -item.Quantity, item.Quantity); err != nil {
			rollback()
			return nil, res.WrapError(err, "Failed to update stock", erres.CommonInternal)
		}

		deductedItems = append(deductedItems, item)

		itemPrice := product.Price
		shopID := product.ShopID
		
		shopItems[shopID] = append(shopItems[shopID], entities.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     itemPrice,
		})
		shopTotals[shopID] += itemPrice * float64(item.Quantity)
	}

	// Process Global Coupon (simple proportional split or just apply to total)
	// For simplicity, we apply a percentage to each sub-order if it's a percentage,
	// or proportionally if it's a fixed amount. 
	var globalDiscountPercentage float64
	var fixedDiscountRemaining float64
	isFixedDiscount := false

	totalAllShops := 0.0
	for _, t := range shopTotals {
		totalAllShops += t
	}

	if req.CouponCode != "" {
		coupon, err := s.couponService.ValidateCouponForAmount(ctx, req.CouponCode, totalAllShops)
		if err == nil {
			if coupon.Type == "percentage" {
				globalDiscountPercentage = coupon.Value / 100.0
			} else if coupon.Type == "fixed_amount" {
				isFixedDiscount = true
				fixedDiscountRemaining = coupon.Value
				if fixedDiscountRemaining > totalAllShops {
					fixedDiscountRemaining = totalAllShops
				}
			}
			go s.couponService.IncrementUsage(context.Background(), req.CouponCode, 1)
		} else {
			rollback()
			return nil, err
		}
	}

	paymentGroupID := fmt.Sprintf("PG-%s", uuid.New().String()[:8])
	var createdOrders []entities.OrderResponse

	for shopID, items := range shopItems {
		subTotal := shopTotals[shopID]
		discountAmount := 0.0

		if globalDiscountPercentage > 0 {
			discountAmount = subTotal * globalDiscountPercentage
		} else if isFixedDiscount {
			proportion := subTotal / totalAllShops
			discountAmount = fixedDiscountRemaining * proportion
		}

		amountAfterDiscount := subTotal - discountAmount
		taxAmount := amountAfterDiscount * 0.10 // 10% VAT
		finalTotal := amountAfterDiscount + taxAmount

		invoiceNumber := fmt.Sprintf("INV-%s", uuid.New().String()[:8])
		now := time.Now().UTC()

		order := entities.Order{
			PaymentGroupID:  paymentGroupID,
			ShopID:          shopID,
			UserID:          userID,
			InvoiceNumber:   invoiceNumber,
			Items:           items,
			SubTotal:        subTotal,
			CouponCode:      req.CouponCode,
			DiscountAmount:  discountAmount,
			TaxAmount:       taxAmount,
			TotalAmount:     finalTotal,
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
			rollback() // Ideally we should rollback previously inserted orders too
			return nil, res.WrapError(err, "Can not create order", erres.CommonInternal)
		}

		createdOrders = append(createdOrders, mapping.ToOrderResponse(createdOrder))
	}

	return createdOrders, nil
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
			// Use context.Background() to avoid cancellation mid-rollback
			_ = s.productRepository.UpdateStock(context.Background(), item.ProductID, item.Quantity, -item.Quantity)
		}
	}

	return nil
}
