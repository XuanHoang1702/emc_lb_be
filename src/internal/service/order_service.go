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
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var reserveStockScript = redis.NewScript(`
local stock_key = KEYS[1]
local qty = tonumber(ARGV[1])
local current = redis.call('GET', stock_key)

if current == false then
	return -1 -- Not found in Redis (cache miss)
end

if tonumber(current) >= qty then
	redis.call('DECRBY', stock_key, qty)
	return 1 -- Success
else
	return -2 -- Insufficient stock
end
`)

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

			if !product.AllowBackorder {
				if product.Stock < item.Quantity {
					return &res.AppError{
						Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
						Code:       erres.CommonBadRequest,
						StatusCode: http.StatusBadRequest,
					}
				}
			}

			if err := s.productRepository.UpdateStock(opCtx, item.ProductID, -item.Quantity, item.Quantity); err != nil {
				return res.WrapError(err, "Failed to update stock in DB", erres.CommonInternal)
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
			coupon, err := s.couponService.ValidateCouponForAmount(opCtx, req.CouponCode, totalAllShops)
			if err == nil {
				if coupon.Type == "percentage" {
					totalDiscount := totalAllShops * (coupon.Value / 100.0)
					if coupon.MaxDiscountAmount > 0 && totalDiscount > coupon.MaxDiscountAmount {
						totalDiscount = coupon.MaxDiscountAmount
					}
					isFixedDiscount = true
					fixedDiscountRemaining = totalDiscount
				} else if coupon.Type == "fixed_amount" {
					isFixedDiscount = true
					fixedDiscountRemaining = coupon.Value
					if fixedDiscountRemaining > totalAllShops {
						fixedDiscountRemaining = totalAllShops
					}
				}
				// We will increment coupon usage at the end after successful commit
			} else {
				return err
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

		if req.CouponCode != "" {
			// Increment synchronously in the transaction path if it succeeds
			s.couponService.IncrementUsage(opCtx, req.CouponCode, 1)
		}

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
		_ = s.productCache.InvalidateAll(ctx)
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

	if order.TotalAmount != paidAmount {
		return &res.AppError{
			Message:    fmt.Sprintf("Payment amount mismatch. Expected: %f, Got: %f", order.TotalAmount, paidAmount),
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
		if s.productCache != nil {
			_ = s.productCache.InvalidateAll(context.Background())
		}
	}

	return nil
}
