package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/config"
	"emc_lb/src/pkg/cache"
	"emc_lb/src/pkg/entities"
	erres "emc_lb/src/pkg/errors"
	"emc_lb/src/pkg/logs"
	"emc_lb/src/pkg/mapping"
	"emc_lb/src/pkg/res"
	"emc_lb/src/pkg/utils"
	"emc_lb/src/pkg/worker"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID string, req entities.CreateOrderRequest) ([]entities.OrderResponse, error)
	CreateOrderFromCheckout(ctx context.Context, userID string, req entities.CheckoutRequest) ([]entities.OrderResponse, error)
	ListOrders(ctx context.Context, userID string) ([]entities.OrderResponse, error)
	GetOrder(ctx context.Context, id string) (entities.OrderResponse, error)
	GetOrderByInvoiceNumber(ctx context.Context, invoiceNumber string) (entities.OrderResponse, error)
	MarkAsPaidByInvoice(ctx context.Context, invoiceNumber string, paidAmount float64) error
	ConfirmPayment(ctx context.Context, invoiceNumber string, paidAmount float64, transactionID string) error
	UpdateOrderStatus(ctx context.Context, id string, status string) error
	CancelOrder(ctx context.Context, userID string, orderID string) error
	ExpireOrder(ctx context.Context, id string) error
}

type orderService struct {
	orderRepository   repository.OrderRepository
	productRepository repository.ProductRepository
	cartRepository    repository.CartRepository
	couponService     CouponService
	inventoryService  InventoryService
	productCache      cache.ProductCacheStore
	pgxpool           *pgxpool.Pool
	taskDistributor   worker.TaskDistributor
}

func NewOrderService(orderRepository repository.OrderRepository, productRepository repository.ProductRepository, cartRepo repository.CartRepository, couponSvc CouponService, inventoryService InventoryService, productCache cache.ProductCacheStore, pgxpool *pgxpool.Pool, taskDistributor worker.TaskDistributor) OrderService {
	return &orderService{
		orderRepository:   orderRepository,
		productRepository: productRepository,
		cartRepository:    cartRepo,
		couponService:     couponSvc,
		inventoryService:  inventoryService,
		productCache:      productCache,
		pgxpool:           pgxpool,
		taskDistributor:   taskDistributor,
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
	coreLogic := func(opCtx context.Context, repo repository.OrderRepository) error {
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

			// Atomic stock deduction: check-and-decrement.
			// For backorder products, use UpdateStock (no precondition).
			if product.AllowBackorder {
				if err := s.productRepository.UpdateStock(opCtx, item.ProductID, -item.Quantity, item.Quantity); err != nil {
					return res.WrapError(err, "Failed to update stock in DB", erres.CommonInternal)
				}
			} else {
				if s.inventoryService != nil {
					ok, err := s.inventoryService.ReserveStock(opCtx, item.ProductID, item.Quantity)
					if err != nil {
						// Fallback to MongoDB if Redis fails or stock not cached
						if dbErr := s.productRepository.DeductStock(opCtx, item.ProductID, item.Quantity); dbErr != nil {
							return &res.AppError{
								Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
								Code:       erres.CommonBadRequest,
								StatusCode: http.StatusBadRequest,
							}
						}
					} else if !ok {
						return &res.AppError{
							Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
							Code:       erres.CommonBadRequest,
							StatusCode: http.StatusBadRequest,
						}
					} else {
						// Reserved in Redis successfully, now deduct in MongoDB
						if dbErr := s.productRepository.DeductStock(opCtx, item.ProductID, item.Quantity); dbErr != nil {
							_ = s.inventoryService.RestoreStock(context.Background(), item.ProductID, item.Quantity)
							return res.WrapError(dbErr, "Failed to deduct stock in DB", erres.CommonInternal)
						}
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
			ttl := config.Get().App.OrderPaymentTTL
			expiresAt := now.Add(ttl)

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
				ExpiresAt:       &expiresAt,
			}

			createdOrder, err := repo.Create(opCtx, order)
			if err != nil {
				return res.WrapError(err, "Can not create order", erres.CommonInternal)
			}

			// Schedule order cancellation (15 minutes TTL)
			if s.taskDistributor != nil {
				payload := &worker.PayloadCancelExpiredOrder{
					OrderID: createdOrder.ID,
				}
				if enqueueErr := s.taskDistributor.DistributeTaskCancelExpiredOrder(ctx, payload, asynq.ProcessAt(expiresAt)); enqueueErr != nil {
					logs.WithContext(ctx).Error("failed to enqueue cancel expired order task", "order_id", createdOrder.ID, "error", enqueueErr)
				}
			}

			createdOrders = append(createdOrders, mapping.ToOrderResponse(createdOrder))
		}

		// Coupon usage already incremented atomically in ValidateAndIncrementUsage above.

		return nil
	}

	var execErr error
	if s.pgxpool != nil {
		tx, err := s.pgxpool.Begin(ctx)
		if err != nil {
			return nil, res.WrapError(err, "Failed to start database transaction", erres.CommonInternal)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		txRepo := s.orderRepository.WithTx(tx)
		if err := coreLogic(ctx, txRepo); err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, res.WrapError(err, "Failed to commit database transaction", erres.CommonInternal)
		}
	} else {
		// Non-transactional fallback (unit tests or standalone mode)
		execErr = coreLogic(ctx, s.orderRepository)
	}

	if execErr != nil {
		return nil, execErr
	}

	// Invalidate product cache since stock has changed
	if s.productCache != nil {
		for _, item := range req.Items {
			if cacheErr := s.productCache.Invalidate(ctx, item.ProductID); cacheErr != nil {
				logs.WithContext(ctx).Warn("product cache invalidation failed after order creation",
					"product_id", item.ProductID, "error", cacheErr)
			}
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

	// Fallback/Legacy if MarkAsPaidByInvoice is used (no idempotency checks here, use ConfirmPayment)
	// We map it to ConfirmPayment using empty transaction ID
	return s.ConfirmPayment(ctx, invoiceNumber, paidAmount, "")
}

func (s *orderService) ConfirmPayment(ctx context.Context, invoiceNumber string, paidAmount float64, transactionID string) error {
	order, err := s.orderRepository.GetByInvoiceNumber(ctx, invoiceNumber)
	if err != nil {
		return fmt.Errorf("order with invoice number %s not found: %w", invoiceNumber, err)
	}

	// Idempotency check
	if order.PaymentTransactionID != nil && *order.PaymentTransactionID == transactionID && order.PaymentStatus == entities.PaymentStatusPaid {
		// Already processed successfully, return nil
		logs.WithContext(ctx).Info("webhook idempotency: already processed payment", "invoice", invoiceNumber, "tx_id", transactionID)
		return nil
	}

	if !utils.MoneyEqual(order.TotalAmount, paidAmount) {
		return &res.AppError{
			Message:    fmt.Sprintf("Payment amount mismatch. Expected: %.2f, Got: %.2f", order.TotalAmount, paidAmount),
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	if order.Status == entities.OrderStatusCancelled {
		// Refund needed scenario
		logs.WithContext(ctx).Warn("payment received for cancelled order (refund needed)", "invoice", invoiceNumber)
		// We can mark payment as paid, but leave order status as cancelled
		_, err := s.orderRepository.ConfirmPaymentAtomic(ctx, invoiceNumber, entities.PaymentStatusUnpaid, entities.PaymentStatusPaid, entities.OrderStatusCancelled, transactionID)
		return err
	}

	newStatus := entities.OrderStatusProcessing
	if order.Status != entities.OrderStatusPending {
		newStatus = order.Status // keep current status if not pending (e.g. already shipped)
	}

	rows, err := s.orderRepository.ConfirmPaymentAtomic(ctx, invoiceNumber, entities.PaymentStatusUnpaid, entities.PaymentStatusPaid, newStatus, transactionID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("could not confirm payment atomically, maybe state changed")
	}

	// Fetch customer info and dispatch email
	if s.taskDistributor != nil {
		email, name, errInfo := s.orderRepository.GetCustomerInfoByUserID(ctx, order.UserID)
		if errInfo != nil {
			logs.WithContext(ctx).Warn("could not get customer info for payment email", "user_id", order.UserID, "error", errInfo)
		} else if email != "" {
			err = s.taskDistributor.DistributeTaskSendOrderPaymentSuccessEmail(ctx, &worker.PayloadSendOrderPaymentSuccessEmail{
				InvoiceNumber: invoiceNumber,
				CustomerEmail: email,
				CustomerName:  name,
				AmountPaid:    paidAmount,
			})
			if err != nil {
				logs.WithContext(ctx).Warn("failed to enqueue payment success email task", "invoice", invoiceNumber, "error", err)
			}
		}
	}

	return nil
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

	// If transitioning to cancelled, restore stock and log any failures.
	if status == entities.OrderStatusCancelled {
		return s.cancelOrderAtomic(ctx, order)
	}

	if err := s.orderRepository.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	return nil
}

func (s *orderService) ExpireOrder(ctx context.Context, id string) error {
	order, err := s.orderRepository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if already cancelled but inventory not returned
	if order.Status == entities.OrderStatusCancelled && !order.InventoryReturned {
		return s.releaseInventory(ctx, order)
	}

	// Only expire if the order is still pending and unpaid.
	if order.Status != entities.OrderStatusPending || order.PaymentStatus != entities.PaymentStatusUnpaid {
		logs.WithContext(ctx).Info("order is no longer pending/unpaid, skipping expiration", "order_id", id, "status", order.Status, "payment_status", order.PaymentStatus)
		return nil
	}

	if order.ExpiresAt != nil && time.Now().UTC().Before(*order.ExpiresAt) {
		logs.WithContext(ctx).Info("order is not yet expired, skipping", "order_id", id)
		return nil
	}

	return s.cancelOrderAtomic(ctx, order)
}

func (s *orderService) cancelOrderAtomic(ctx context.Context, order entities.Order) error {
	if err := s.orderRepository.UpdateStatusAtomic(ctx, order.ID, order.Status, entities.OrderStatusCancelled); err != nil {
		if err == pgx.ErrNoRows {
			// It might be already cancelled. Check if we need to release inventory.
			currentOrder, fetchErr := s.orderRepository.GetByID(ctx, order.ID)
			if fetchErr == nil && currentOrder.Status == entities.OrderStatusCancelled && !currentOrder.InventoryReturned {
				return s.releaseInventory(ctx, currentOrder)
			}
			return &res.AppError{
				Message:    "Order is no longer in a state that can be cancelled",
				Code:       erres.CommonBadRequest,
				StatusCode: http.StatusBadRequest,
			}
		}
		return res.WrapError(err, "Failed to update order status", erres.CommonInternal)
	}

	return s.releaseInventory(ctx, order)
}

func (s *orderService) releaseInventory(ctx context.Context, order entities.Order) error {
	if order.InventoryReturned {
		return nil
	}

	for _, item := range order.Items {
		// Idempotent Mongo restore
		if restoreErr := s.productRepository.RestoreStockIdempotent(ctx, item.ProductID, item.Quantity, order.ID); restoreErr != nil {
			worker.OrderCancellationInventoryReleaseFailure.Inc()
			return res.WrapError(restoreErr, "Failed to restore stock on cancellation", erres.CommonInternal)
		}

		// Idempotent Redis restore
		if s.inventoryService != nil {
			if restoreErr := s.inventoryService.RestoreStockIdempotent(context.Background(), item.ProductID, item.Quantity, order.ID); restoreErr != nil {
				worker.OrderCancellationInventoryReleaseFailure.Inc()
				logs.WithContext(ctx).Warn("Failed to restore stock in redis", "error", restoreErr)
			}
		}

		if s.productCache != nil {
			if cacheErr := s.productCache.Invalidate(ctx, item.ProductID); cacheErr != nil {
				logs.WithContext(ctx).Warn("product cache invalidation failed after cancellation", "order_id", order.ID, "product_id", item.ProductID, "error", cacheErr)
			}
		}
	}

	if err := s.orderRepository.MarkInventoryReturned(ctx, order.ID); err != nil {
		if err != pgx.ErrNoRows {
			worker.OrderCancellationInventoryReleaseFailure.Inc()
			return res.WrapError(err, "Failed to mark inventory as returned", erres.CommonInternal)
		}
	}

	return nil
}

func (s *orderService) CancelOrder(ctx context.Context, userID string, orderID string) error {
	order, err := s.orderRepository.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.UserID != userID {
		return &res.AppError{
			Message:    "You do not have permission to cancel this order",
			Code:       erres.CommonForbidden,
			StatusCode: http.StatusForbidden,
		}
	}

	if order.Status != entities.OrderStatusPending && order.Status != entities.OrderStatusProcessing {
		return &res.AppError{
			Message:    "Only pending or processing orders can be cancelled",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	return s.cancelOrderAtomic(ctx, order)
}

//nolint:gocyclo
func (s *orderService) CreateOrderFromCheckout(ctx context.Context, userID string, req entities.CheckoutRequest) ([]entities.OrderResponse, error) {
	if len(req.Items) == 0 {
		return nil, &res.AppError{
			Message:    "Order must have at least one item",
			Code:       erres.CommonBadRequest,
			StatusCode: http.StatusBadRequest,
		}
	}

	var createdOrders []entities.OrderResponse
	var productIDsToClearFromCart []string

	// Core order creation logic, called within or without a transaction.
	coreLogic := func(opCtx context.Context, repo repository.OrderRepository) error {
		createdOrders = nil
		productIDsToClearFromCart = nil

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

			if product.Status != "active" {
				return &res.AppError{
					Message:    fmt.Sprintf("Product %s is not active", product.Name),
					Code:       erres.CommonBadRequest,
					StatusCode: http.StatusBadRequest,
				}
			}

			productIDsToClearFromCart = append(productIDsToClearFromCart, item.ProductID)

			// Atomic stock deduction: check-and-decrement.
			// For backorder products, use UpdateStock (no precondition).
			if product.AllowBackorder {
				if err := s.productRepository.UpdateStock(opCtx, item.ProductID, -item.Quantity, item.Quantity); err != nil {
					return res.WrapError(err, "Failed to update stock in DB", erres.CommonInternal)
				}
			} else {
				if s.inventoryService != nil {
					ok, err := s.inventoryService.ReserveStock(opCtx, item.ProductID, item.Quantity)
					if err != nil {
						// Fallback to MongoDB if Redis fails or stock not cached
						if dbErr := s.productRepository.DeductStock(opCtx, item.ProductID, item.Quantity); dbErr != nil {
							return &res.AppError{
								Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
								Code:       erres.CommonBadRequest,
								StatusCode: http.StatusBadRequest,
							}
						}
					} else if !ok {
						return &res.AppError{
							Message:    fmt.Sprintf("Not enough stock for product %s", product.Name),
							Code:       erres.CommonBadRequest,
							StatusCode: http.StatusBadRequest,
						}
					} else {
						// Reserved in Redis successfully, now deduct in MongoDB
						if dbErr := s.productRepository.DeductStock(opCtx, item.ProductID, item.Quantity); dbErr != nil {
							_ = s.inventoryService.RestoreStock(context.Background(), item.ProductID, item.Quantity)
							return res.WrapError(dbErr, "Failed to deduct stock in DB", erres.CommonInternal)
						}
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
			}

			itemPrice := product.Price
			shopID := product.ShopID
			subTotal := itemPrice * float64(item.Quantity)

			shopItems[shopID] = append(shopItems[shopID], entities.OrderItem{
				ProductID:   item.ProductID,
				ProductName: product.Name,
				SKU:         product.SKU,
				Thumbnail:   product.Thumbnail,
				Quantity:    item.Quantity,
				Price:       itemPrice,
				SubTotal:    subTotal,
			})
			shopTotals[shopID] += subTotal
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
			ttl := config.Get().App.OrderPaymentTTL
			expiresAt := now.Add(ttl)

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
				ExpiresAt:       &expiresAt,
			}

			createdOrder, err := repo.Create(opCtx, order)
			if err != nil {
				return res.WrapError(err, "Can not create order", erres.CommonInternal)
			}

			// Schedule order cancellation (15 minutes TTL)
			if s.taskDistributor != nil {
				payload := &worker.PayloadCancelExpiredOrder{
					OrderID: createdOrder.ID,
				}
				if enqueueErr := s.taskDistributor.DistributeTaskCancelExpiredOrder(ctx, payload, asynq.ProcessAt(expiresAt)); enqueueErr != nil {
					logs.WithContext(ctx).Error("failed to enqueue cancel expired order task", "order_id", createdOrder.ID, "error", enqueueErr)
				}
			}

			createdOrders = append(createdOrders, mapping.ToOrderResponse(createdOrder))
		}

		return nil
	}

	var execErr error
	if s.pgxpool != nil {
		tx, err := s.pgxpool.Begin(ctx)
		if err != nil {
			return nil, res.WrapError(err, "Failed to start database transaction", erres.CommonInternal)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		txRepo := s.orderRepository.WithTx(tx)
		if err := coreLogic(ctx, txRepo); err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, res.WrapError(err, "Failed to commit database transaction", erres.CommonInternal)
		}
	} else {
		// Non-transactional fallback (unit tests or standalone mode)
		execErr = coreLogic(ctx, s.orderRepository)
	}

	if execErr != nil {
		return nil, execErr
	}

	// Post-commit tasks
	// 1. Invalidate product cache since stock has changed
	if s.productCache != nil {
		for _, item := range req.Items {
			if cacheErr := s.productCache.Invalidate(ctx, item.ProductID); cacheErr != nil {
				logs.WithContext(ctx).Warn("product cache invalidation failed after order creation",
					"product_id", item.ProductID, "error", cacheErr)
			}
		}
	}

	// 2. Clear cart
	if s.cartRepository != nil && len(productIDsToClearFromCart) > 0 {
		if err := s.cartRepository.RemoveItems(ctx, userID, productIDsToClearFromCart); err != nil {
			logs.WithContext(ctx).Warn("failed to clear cart after order creation", "user_id", userID, "error", err)
		}
	}

	return createdOrders, nil
}
