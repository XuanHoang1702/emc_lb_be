package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) ([]entities.OrderResponse, error)
	ListOrders(ctx context.Context, userID string) ([]entities.OrderResponse, error)
	GetOrder(ctx context.Context, id string) (entities.OrderResponse, error)
	GetOrderByInvoiceNumber(ctx context.Context, invoiceNumber string) (entities.OrderResponse, error)
	MarkAsPaidByInvoice(ctx context.Context, invoiceNumber string, paidAmount float64) error
	UpdateOrderStatus(ctx context.Context, id string, status string) error
}

type orderService struct {
	orderRepository   repository.OrderRepository
	productRepository repository.ProductRepository
	couponService     CouponService
	redisClient       *redis.Client
	productCache      cache.ProductCacheStore
	mongoClient       *mongo.Client
}

func NewOrderService(orderRepository repository.OrderRepository, productRepository repository.ProductRepository, couponSvc CouponService, redisClient *redis.Client, productCache cache.ProductCacheStore, mongoClient *mongo.Client) OrderService {
	return &orderService{
		orderRepository:   orderRepository,
		productRepository: productRepository,
		couponService:     couponSvc,
		redisClient:       redisClient,
		productCache:      productCache,
		mongoClient:       mongoClient,
	}
}

//nolint:gocyclo
func (s *orderService) CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) ([]entities.OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, &res.AppError{
			Message:    "Order must have at least one item",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	var createdOrders []entities.OrderResponse
	// Core order creation logic, called within or without a transaction.
	coreLogic := func(opCtx context.Context) error {
		createdOrders = nil

		shopItems := make(map[string][]entities.OrderItem)
		shopTotals := make(map[string]float64)

		for _, item := range req.Items {
			product, err := s.productRepository.GetByID(opCtx, item.ProductID)
			if err != nil {
				return &res.AppError{
					Message:    fmt.Sprintf("Product %s not found", item.ProductID),
					Code:       erres.CommonBadRequest,
					StatusCode: http.StatusBadRequest,
				}
			}

			// Atomic stock deduction: check-and-decrement in one MongoDB operation.
			// For backorder products, use UpdateStock (no precondition).
			if product.AllowBackorder {
				if err := s.productRepository.UpdateStock(opCtx, item.ProductID, -item.Quantity, item.Quantity); err != nil {
					return res.WrapError(err, "Failed to update stock in DB", erres.CommonInternal)
				}
			} else {
				if err := s.productRepository.DeductStock(opCtx, item.ProductID, item.Quantity); err != nil {
					return &res.AppError{
						Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
						Code:       erres.CommonBadRequest,
						StatusCode: http.StatusBadRequest,
					}
				}
			}

			itemPrice := product.Price
			shopID := product.ShopID

			shopItems[shopID] = append(shopItems[shopID], entities.OrderItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     itemPrice,
			})
			shopTotals[shopID] += itemPrice * float64(item.Quantity)
		}

		// Process coupon
		var fixedDiscountRemaining float64
		isFixedDiscount := false

		totalAllShops := 0.0
		for _, t := range shopTotals {
			totalAllShops += t
		}

		if req.CouponCode != "" {
			// Atomic validate + increment: prevents TOCTOU race where
			// concurrent requests all pass validation before any increments usage.
			coupon, err := s.couponService.ValidateAndIncrementUsage(opCtx, req.CouponCode, totalAllShops)
			if err != nil {
				return err
			}

			switch coupon.Type {
			case "percentage":
				totalDiscount := totalAllShops * (coupon.Value / 100.0)
				if coupon.MaxDiscountAmount > 0 && totalDiscount > coupon.MaxDiscountAmount {
					totalDiscount = coupon.MaxDiscountAmount
				}
				isFixedDiscount = true
				fixedDiscountRemaining = totalDiscount
			case "fixed_amount":
				isFixedDiscount = true
				fixedDiscountRemaining = coupon.Value
				if fixedDiscountRemaining > totalAllShops {
					fixedDiscountRemaining = totalAllShops
				}
			}
		}

		paymentGroupID := fmt.Sprintf("PG-%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:16])

		for shopID, items := range shopItems {
			subTotal := shopTotals[shopID]
			discountAmount := 0.0

			if isFixedDiscount {
				proportion := subTotal / totalAllShops
				discountAmount = fixedDiscountRemaining * proportion
			}

			amountAfterDiscount := subTotal - discountAmount
			taxAmount := amountAfterDiscount * 0.10 // 10% VAT
			finalTotal := amountAfterDiscount + taxAmount

			invoiceNumber := fmt.Sprintf("INV-%s", strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))[:16])
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

			createdOrder, err := s.orderRepository.Create(opCtx, order)
			if err != nil {
				return res.WrapError(err, "Can not create order", erres.CommonInternal)
			}

			createdOrders = append(createdOrders, mapping.ToOrderResponse(createdOrder))
		}

		// Coupon usage already incremented atomically in ValidateAndIncrementUsage above.

		return nil
	}

	var execErr error
	if s.mongoClient != nil {
		// Production path: use MongoDB transaction for atomicity
		session, err := s.mongoClient.StartSession()
		if err != nil {
			return nil, res.WrapError(err, "Failed to start database session", erres.CommonInternal)
		}
		defer session.EndSession(ctx)

		_, execErr = session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
			if err := coreLogic(sessCtx); err != nil {
				return nil, err
			}
			return nil, nil
		})
	} else {
		// Non-transactional fallback (unit tests or standalone mode)
		execErr = coreLogic(ctx)
	}

	if execErr != nil {
		return nil, execErr
	}

	// Invalidate product cache since stock has changed
	if s.productCache != nil {
		if cacheErr := s.productCache.InvalidateAll(ctx); cacheErr != nil {
			logs.WithContext(ctx).Warn("product cache invalidation failed after order creation",
				"error", cacheErr)
		}
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

func (s *orderService) GetOrderByInvoiceNumber(ctx context.Context, invoiceNumber string) (entities.OrderResponse, error) {
	order, err := s.orderRepository.GetByInvoiceNumber(ctx, invoiceNumber)
	if err != nil {
		return entities.OrderResponse{}, res.WrapError(err, "Order not found", erres.CommonNotFound)
	}
	return mapping.ToOrderResponse(order), nil
}

func (s *orderService) MarkAsPaidByInvoice(ctx context.Context, invoiceNumber string, paidAmount float64) error {
	order, err := s.orderRepository.GetByInvoiceNumber(ctx, invoiceNumber)
	if err != nil {
		return fmt.Errorf("order with invoice number %s not found: %w", invoiceNumber, err)
	}

	if !utils.MoneyEqual(order.TotalAmount, paidAmount) {
		return &res.AppError{
			Message:    fmt.Sprintf("Payment amount mismatch. Expected: %.2f, Got: %.2f", order.TotalAmount, paidAmount),
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	return s.orderRepository.UpdatePaymentStatus(ctx, order.ID, "paid")
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id string, status string) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate state transition using the order status state machine.
	if !entities.CanTransition(order.Status, status) {
		return &res.AppError{
			Message:    fmt.Sprintf("Cannot transition order from '%s' to '%s'", order.Status, status),
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if err := s.orderRepository.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	// If transitioning to cancelled, restore stock and log any failures.
	if status == entities.OrderStatusCancelled {
		var restoreErrors []string
		for _, item := range order.Items {
			if restoreErr := s.productRepository.UpdateStock(ctx, item.ProductID, item.Quantity, -item.Quantity); restoreErr != nil {
				logs.WithContext(ctx).Error("failed to restore stock on cancellation",
					"order_id", id,
					"product_id", item.ProductID,
					"quantity", item.Quantity,
					"error", restoreErr,
				)
				restoreErrors = append(restoreErrors, item.ProductID)
			}
		}
		if s.productCache != nil {
			if cacheErr := s.productCache.InvalidateAll(ctx); cacheErr != nil {
				logs.WithContext(ctx).Warn("product cache invalidation failed after cancellation",
					"order_id", id, "error", cacheErr)
			}
		}
		if len(restoreErrors) > 0 {
			return &res.AppError{
				Message:    fmt.Sprintf("Order cancelled but failed to restore stock for products: %s", strings.Join(restoreErrors, ", ")),
				Code:       erres.CommonInternal,
				StatusCode: http.StatusInternalServerError,
			}
		}
	}

	return nil
}
