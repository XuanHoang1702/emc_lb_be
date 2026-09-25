package service_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/internal/service"
	"emc_lb/src/pkg/entities"
	"emc_lb/src/pkg/res"
	"sync"

	"github.com/jackc/pgx/v5"
	"os"
)

func init() {
	os.Setenv("PORT", "8080")
	os.Setenv("ENVIRONMENT", "test")
	os.Setenv("SYSTEM_SECRET", "test-secret")
	os.Setenv("POSTGRES_URI", "postgres://user:pass@localhost:5432/db")
	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("REDIS_URI", "redis://localhost:6379")
	os.Setenv("SEPAY_API_KEY", "test")
	os.Setenv("RESEND_API_KEY", "test")
	os.Setenv("AWS_REGION", "test")
	os.Setenv("AWS_ACCESS_KEY_ID", "test")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	os.Setenv("AWS_S3_ENDPOINT", "test")
	os.Setenv("AWS_S3_BUCKET_NAME", "test")
	os.Setenv("ACCESS_TOKEN_SECRET", "test")
	os.Setenv("REFRESH_TOKEN_SECRET", "test")
}

// stubOrderRepository for testing
type stubOrderRepository struct {
	mu             sync.Mutex
	orders         []entities.Order
	createFn       func(entities.Order) (entities.Order, error)
	updateStatusFn func(string, string) error
}

func newStubOrderRepo() *stubOrderRepository {
	return &stubOrderRepository{}
}

func (r *stubOrderRepository) Create(_ context.Context, o entities.Order) (entities.Order, error) {
	if r.createFn != nil {
		return r.createFn(o)
	}
	o.ID = fmt.Sprintf("order-%d", len(r.orders)+1)
	r.orders = append(r.orders, o)
	return o, nil
}

func (r *stubOrderRepository) List(_ context.Context, _ string) ([]entities.Order, error) {
	return r.orders, nil
}

func (r *stubOrderRepository) GetByID(_ context.Context, id string) (entities.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, o := range r.orders {
		if o.ID == id {
			return o, nil
		}
	}
	return entities.Order{}, errors.New("not found")
}

func (r *stubOrderRepository) UpdateStatusAtomic(_ context.Context, id string, expectedCurrent string, newStatus string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, o := range r.orders {
		if o.ID == id {
			if o.Status != expectedCurrent {
				return fmt.Errorf("state mismatch: expected %s, got %s", expectedCurrent, o.Status)
			}
			r.orders[i].Status = newStatus
			return nil
		}
	}
	return fmt.Errorf("not found order with id %s", id)
}

func (r *stubOrderRepository) ConfirmPaymentAtomic(_ context.Context, invoiceNumber, expectedCurrentPaymentStatus, newPaymentStatus, newOrderStatus, transactionID string) (int64, error) {
	for i, o := range r.orders {
		if o.InvoiceNumber == invoiceNumber {
			if o.PaymentStatus != expectedCurrentPaymentStatus {
				return 0, nil
			}
			r.orders[i].PaymentStatus = newPaymentStatus
			r.orders[i].Status = newOrderStatus
			tx := transactionID
			r.orders[i].PaymentTransactionID = &tx
			return 1, nil
		}
	}
	return 0, nil
}

func (r *stubOrderRepository) MarkInventoryReturned(ctx context.Context, id string) error {
	return nil
}

func (r *stubOrderRepository) GetCustomerInfoByUserID(ctx context.Context, userID string) (email string, name string, err error) {
	return "test@example.com", "Test User", nil
}

func (r *stubOrderRepository) WithTx(tx pgx.Tx) repository.OrderRepository {
	return r
}

func (r *stubOrderRepository) UpdateStatus(_ context.Context, id string, status string) error {
	if r.updateStatusFn != nil {
		return r.updateStatusFn(id, status)
	}
	for i, o := range r.orders {
		if o.ID == id {
			r.orders[i].Status = status
			return nil
		}
	}
	return errors.New("not found")
}

func (r *stubOrderRepository) UpdatePaymentStatus(_ context.Context, id string, status string) error {
	for i, o := range r.orders {
		if o.ID == id {
			r.orders[i].PaymentStatus = status
			return nil
		}
	}
	return errors.New("not found")
}

func (r *stubOrderRepository) GetByInvoiceNumber(_ context.Context, invoiceNumber string) (entities.Order, error) {
	for _, o := range r.orders {
		if o.InvoiceNumber == invoiceNumber {
			return o, nil
		}
	}
	return entities.Order{}, errors.New("not found")
}

func (r *stubOrderRepository) EnsureIndexes(_ context.Context) error {
	return nil
}

// stubProductRepository for testing
type stubProductRepository struct {
	mu          sync.Mutex
	products    map[string]entities.Product
	stockDeltas map[string]int64
}

func newStubProductRepo() *stubProductRepository {
	return &stubProductRepository{
		products:    make(map[string]entities.Product),
		stockDeltas: make(map[string]int64),
	}
}

func (r *stubProductRepository) Create(_ context.Context, p entities.Product) (entities.Product, error) {
	r.products[p.ID] = p
	return p, nil
}

func (r *stubProductRepository) List(_ context.Context) ([]entities.Product, error) {
	var result []entities.Product
	for _, p := range r.products {
		result = append(result, p)
	}
	return result, nil
}

func (r *stubProductRepository) GetByID(_ context.Context, id string) (entities.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[id]
	if !ok {
		return entities.Product{}, errors.New("not found")
	}
	return p, nil
}

func (r *stubProductRepository) Update(_ context.Context, _ string, _ map[string]any) (entities.Product, error) {
	return entities.Product{}, nil
}

func (r *stubProductRepository) Delete(_ context.Context, _ string, _ map[string]any) error {
	return nil
}

func (r *stubProductRepository) UpdateStock(_ context.Context, id string, stockDelta int64, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stockDeltas[id] += stockDelta
	if p, ok := r.products[id]; ok {
		p.Stock += stockDelta
		r.products[id] = p
	}
	return nil
}

func (r *stubProductRepository) DeductStock(_ context.Context, id string, quantity int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[id]
	if !ok {
		return errors.New("not found")
	}
	if p.Stock < quantity {
		return errors.New("insufficient stock")
	}
	p.Stock -= quantity
	p.SoldCount += quantity
	r.products[id] = p
	r.stockDeltas[id] -= quantity
	return nil
}

func (r *stubProductRepository) RestoreStockIdempotent(ctx context.Context, id string, quantity int64, orderID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stockDeltas[id] += quantity
	if p, ok := r.products[id]; ok {
		p.Stock += quantity
		r.products[id] = p
	}
	return nil
}

// stubCouponSvc for testing
type stubCouponSvc struct {
	validateFn func(code string, amount float64) (entities.Coupon, error)
}

func (s *stubCouponSvc) Create(_ context.Context, _ entities.CreateCouponRequest) (entities.CouponResponse, error) {
	return entities.CouponResponse{}, nil
}

func (s *stubCouponSvc) GetByCode(_ context.Context, _ string) (entities.CouponResponse, error) {
	return entities.CouponResponse{}, nil
}

func (s *stubCouponSvc) List(_ context.Context) ([]entities.CouponResponse, error) {
	return nil, nil
}

func (s *stubCouponSvc) Update(_ context.Context, _ string, _ entities.UpdateCouponRequest) (entities.CouponResponse, error) {
	return entities.CouponResponse{}, nil
}

func (s *stubCouponSvc) ValidateCouponForAmount(_ context.Context, code string, amount float64) (entities.Coupon, error) {
	if s.validateFn != nil {
		return s.validateFn(code, amount)
	}
	return entities.Coupon{}, errors.New("no coupon")
}

func (s *stubCouponSvc) IncrementUsage(_ context.Context, _ string, _ int64) error {
	return nil
}

func (s *stubCouponSvc) ValidateAndIncrementUsage(_ context.Context, code string, amount float64) (entities.Coupon, error) {
	if s.validateFn != nil {
		return s.validateFn(code, amount)
	}
	return entities.Coupon{}, errors.New("no coupon")
}

// stubProductCache for testing
type stubProductCache struct {
	invalidated bool
}

func (c *stubProductCache) GetList(_ context.Context, _ string) ([]entities.ProductResponse, error) {
	return nil, errors.New("miss")
}
func (c *stubProductCache) SetList(_ context.Context, _ string, _ []entities.ProductResponse) error {
	return nil
}
func (c *stubProductCache) GetByID(_ context.Context, _ string) (entities.ProductResponse, error) {
	return entities.ProductResponse{}, errors.New("miss")
}
func (c *stubProductCache) SetByID(_ context.Context, _ string, _ entities.ProductResponse) error {
	return nil
}
func (c *stubProductCache) Invalidate(_ context.Context, _ string) error {
	c.invalidated = true
	return nil
}
func (c *stubProductCache) InvalidateList(_ context.Context) error {
	c.invalidated = true
	return nil
}

func sampleProduct(id string, shopID string, price float64, stock int64) entities.Product {
	return entities.Product{
		ID:     id,
		Name:   "Product " + id,
		ShopID: shopID,
		Price:  price,
		Stock:  stock,
		Status: "active",
	}
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	svc := service.NewOrderService(newStubOrderRepo(), newStubProductRepo(), nil, &stubCouponSvc{}, nil, nil, nil, nil)

	_, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{})
	if err == nil {
		t.Fatal("expected error for empty items")
	}
	appErr, ok := err.(*res.AppError)
	if !ok {
		t.Fatal("expected AppError")
	}
	if appErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", appErr.StatusCode)
	}
}

func TestCreateOrder_Success(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 10)
	orderRepo := newStubOrderRepo()

	// No redis client → fallback to DB stock check (AllowBackorder=false by default)
	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	orders, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{
		Items: []entities.OrderItem{{ProductID: "p1", Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].Status != "pending" {
		t.Fatalf("expected pending status, got %s", orders[0].Status)
	}
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 1) // only 1 in stock
	orderRepo := newStubOrderRepo()

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	_, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{
		Items: []entities.OrderItem{{ProductID: "p1", Quantity: 5}},
	})
	if err == nil {
		t.Fatal("expected error for insufficient stock")
	}
}

func TestCreateOrder_WithPercentageCoupon(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 200, 10)
	orderRepo := newStubOrderRepo()

	couponSvc := &stubCouponSvc{
		validateFn: func(_ string, _ float64) (entities.Coupon, error) {
			return entities.Coupon{
				Code:  "SAVE20",
				Type:  "percentage",
				Value: 20, // 20% off
			}, nil
		},
	}

	svc := service.NewOrderService(orderRepo, productRepo, nil, couponSvc, nil, nil, nil, nil)

	orders, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{
		Items:      []entities.OrderItem{{ProductID: "p1", Quantity: 1}},
		CouponCode: "SAVE20",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// SubTotal=200, Discount=40 (20%), After=160, Tax=16 (10%), Total=176
	order := orders[0]
	if order.DiscountAmount != 40 {
		t.Fatalf("expected discount 40, got %f", order.DiscountAmount)
	}
	expectedTotal := 176.0
	if order.TotalAmount != expectedTotal {
		t.Fatalf("expected total %f, got %f", expectedTotal, order.TotalAmount)
	}
}

func TestCreateOrder_WithFixedCoupon(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 500, 10)
	orderRepo := newStubOrderRepo()

	couponSvc := &stubCouponSvc{
		validateFn: func(_ string, _ float64) (entities.Coupon, error) {
			return entities.Coupon{
				Code:  "FLAT100",
				Type:  "fixed_amount",
				Value: 100,
			}, nil
		},
	}

	svc := service.NewOrderService(orderRepo, productRepo, nil, couponSvc, nil, nil, nil, nil)

	orders, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{
		Items:      []entities.OrderItem{{ProductID: "p1", Quantity: 1}},
		CouponCode: "FLAT100",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// SubTotal=500, Discount=100, After=400, Tax=40 (10%), Total=440
	order := orders[0]
	if order.DiscountAmount != 100 {
		t.Fatalf("expected discount 100, got %f", order.DiscountAmount)
	}
}

func TestCreateOrder_MultiShopSplit(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 10)
	productRepo.products["p2"] = sampleProduct("p2", "shop-B", 200, 10)
	orderRepo := newStubOrderRepo()

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	orders, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{
		Items: []entities.OrderItem{
			{ProductID: "p1", Quantity: 1},
			{ProductID: "p2", Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders (multi-shop split), got %d", len(orders))
	}
}

func TestCreateOrder_InvalidatesProductCache(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 10)
	orderRepo := newStubOrderRepo()
	productCache := &stubProductCache{}

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, productCache, nil, nil)

	_, err := svc.CreateOrder(context.Background(), "user-1", entities.CreateOrderRequest{
		Items: []entities.OrderItem{{ProductID: "p1", Quantity: 1}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !productCache.invalidated {
		t.Fatal("expected product cache to be invalidated after order creation")
	}
}

func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	orderRepo := newStubOrderRepo()
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		Status: "pending",
	})

	svc := service.NewOrderService(orderRepo, newStubProductRepo(), nil, &stubCouponSvc{}, nil, nil, nil, nil)

	err := svc.UpdateOrderStatus(context.Background(), "order-1", "invalid_status")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestUpdateOrderStatus_AlreadyCancelled(t *testing.T) {
	orderRepo := newStubOrderRepo()
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		Status: "cancelled",
	})

	svc := service.NewOrderService(orderRepo, newStubProductRepo(), nil, &stubCouponSvc{}, nil, nil, nil, nil)

	err := svc.UpdateOrderStatus(context.Background(), "order-1", "processing")
	if err == nil {
		t.Fatal("expected error when trying to change cancelled order")
	}
}

func TestUpdateOrderStatus_CancelledRestoresStock(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 8)
	orderRepo := newStubOrderRepo()
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		Status: "pending",
		Items:  []entities.OrderItem{{ProductID: "p1", Quantity: 2}},
	})

	productCache := &stubProductCache{}
	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, productCache, nil, nil)

	err := svc.UpdateOrderStatus(context.Background(), "order-1", "cancelled")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check stock was restored
	if productRepo.stockDeltas["p1"] != 2 {
		t.Fatalf("expected stock delta +2 for p1, got %d", productRepo.stockDeltas["p1"])
	}

	// Check cache invalidated
	if !productCache.invalidated {
		t.Fatal("expected product cache to be invalidated after cancellation")
	}
}

func TestCheckout_SnapshotsProductData(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 10)
	orderRepo := newStubOrderRepo()

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	orders, err := svc.CreateOrderFromCheckout(context.Background(), "user-1", entities.CheckoutRequest{
		Items: []entities.CheckoutItemRequest{{ProductID: "p1", Quantity: 2}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	order := orders[0]
	item := order.Items[0]
	if item.ProductName != "Product p1" {
		t.Fatalf("expected ProductName 'Product p1', got %s", item.ProductName)
	}
	if item.Price != 100 {
		t.Fatalf("expected Price 100, got %f", item.Price)
	}
	if item.SubTotal != 200 {
		t.Fatalf("expected SubTotal 200, got %f", item.SubTotal)
	}
}

func TestCancelOrder_BuyerCancelPending(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 8)
	orderRepo := newStubOrderRepo()
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: "pending",
		Items:  []entities.OrderItem{{ProductID: "p1", Quantity: 2}},
	})

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	err := svc.CancelOrder(context.Background(), "user-1", "order-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if productRepo.stockDeltas["p1"] != 2 {
		t.Fatalf("expected stock delta +2, got %d", productRepo.stockDeltas["p1"])
	}
}

func TestConcurrency_SingleStock_TwoRequests(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 1) // Stock = 1
	orderRepo := newStubOrderRepo()

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	var successCount, failCount int
	done := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func(idx int) {
			_, err := svc.CreateOrderFromCheckout(context.Background(), fmt.Sprintf("user-%d", idx), entities.CheckoutRequest{
				Items: []entities.CheckoutItemRequest{{ProductID: "p1", Quantity: 1}},
			})
			done <- err
		}(i)
	}

	for i := 0; i < 2; i++ {
		err := <-done
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 success, got %d", successCount)
	}
	if failCount != 1 {
		t.Fatalf("expected exactly 1 failure, got %d", failCount)
	}

	if productRepo.products["p1"].Stock != 0 {
		t.Fatalf("expected stock to be 0, got %d", productRepo.products["p1"].Stock)
	}
}

func TestConcurrency_DoubleCancelPrevention(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 8) // Current stock 8
	orderRepo := newStubOrderRepo()

	orderID := "order-cancel-race"
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     orderID,
		UserID: "user-1",
		Status: "pending",
		Items:  []entities.OrderItem{{ProductID: "p1", Quantity: 2}}, // Original stock was 10
	})

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	var successCount, failCount int
	done := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			err := svc.CancelOrder(context.Background(), "user-1", orderID)
			done <- err
		}()
	}

	for i := 0; i < 2; i++ {
		err := <-done
		if err == nil {
			successCount++
		} else {
			failCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 success cancel, got %d", successCount)
	}
	if failCount != 1 {
		t.Fatalf("expected exactly 1 failed cancel, got %d", failCount)
	}

	// Stock should only be restored ONCE (+2), so 8 + 2 = 10
	if productRepo.stockDeltas["p1"] != 2 {
		t.Fatalf("expected stock delta +2, got %d", productRepo.stockDeltas["p1"])
	}
}

func TestExpireOrder_Success(t *testing.T) {
	productRepo := newStubProductRepo()
	productRepo.products["p1"] = sampleProduct("p1", "shop-A", 100, 8)
	orderRepo := newStubOrderRepo()
	
	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)
	
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: "pending",
		PaymentStatus: "unpaid",
		ExpiresAt: &past,
		Items:  []entities.OrderItem{{ProductID: "p1", Quantity: 2}},
	})

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	err := svc.ExpireOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if productRepo.stockDeltas["p1"] != 2 {
		t.Fatalf("expected stock delta +2, got %d", productRepo.stockDeltas["p1"])
	}
}

func TestExpireOrder_NotYetExpired(t *testing.T) {
	productRepo := newStubProductRepo()
	orderRepo := newStubOrderRepo()
	
	now := time.Now().UTC()
	future := now.Add(1 * time.Hour)
	
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: "pending",
		PaymentStatus: "unpaid",
		ExpiresAt: &future,
		Items:  []entities.OrderItem{{ProductID: "p1", Quantity: 2}},
	})

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	err := svc.ExpireOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if productRepo.stockDeltas["p1"] != 0 {
		t.Fatalf("expected stock delta 0, got %d", productRepo.stockDeltas["p1"])
	}
	// order status should still be pending
	if orderRepo.orders[0].Status != "pending" {
		t.Fatalf("expected pending status, got %s", orderRepo.orders[0].Status)
	}
}

func TestExpireOrder_AlreadyCancelledInventoryNotReturned(t *testing.T) {
	productRepo := newStubProductRepo()
	orderRepo := newStubOrderRepo()
	
	orderRepo.orders = append(orderRepo.orders, entities.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: "cancelled",
		PaymentStatus: "unpaid",
		InventoryReturned: false,
		Items:  []entities.OrderItem{{ProductID: "p1", Quantity: 2}},
	})

	svc := service.NewOrderService(orderRepo, productRepo, nil, &stubCouponSvc{}, nil, nil, nil, nil)

	err := svc.ExpireOrder(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if productRepo.stockDeltas["p1"] != 2 {
		t.Fatalf("expected stock delta +2, got %d", productRepo.stockDeltas["p1"])
	}
}

