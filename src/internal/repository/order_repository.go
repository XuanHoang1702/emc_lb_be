package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/entities"
)

type OrderRepository interface {
	WithTx(tx pgx.Tx) OrderRepository
	Create(context.Context, entities.Order) (entities.Order, error)
	List(context.Context, string) ([]entities.Order, error)
	GetByID(context.Context, string) (entities.Order, error)
	GetByInvoiceNumber(context.Context, string) (entities.Order, error)
	UpdateStatus(context.Context, string, string) error
	UpdateStatusAtomic(context.Context, string, string, string) error
	ConfirmPaymentAtomic(ctx context.Context, invoiceNumber, expectedCurrentPaymentStatus, newPaymentStatus, newOrderStatus, transactionID string) (int64, error)
	MarkInventoryReturned(ctx context.Context, id string) error
	GetCustomerInfoByUserID(ctx context.Context, userID string) (email string, name string, err error)
	GetExpiredPendingOrders(ctx context.Context) ([]entities.Order, error)
}

type orderRepository struct {
	pool *pgxpool.Pool
	db   *sqlc.Queries
}

func NewOrderRepository(pool *pgxpool.Pool, db sqlc.Querier) OrderRepository {
	return &orderRepository{
		pool: pool,
		db:   db.(*sqlc.Queries),
	}
}

func (r *orderRepository) WithTx(tx pgx.Tx) OrderRepository {
	return &orderRepository{
		pool: r.pool,
		db:   r.db.WithTx(tx),
	}
}

func (r *orderRepository) Create(ctx context.Context, order entities.Order) (entities.Order, error) {
	userUUID, err := uuid.Parse(order.UserID)
	if err != nil {
		return entities.Order{}, err
	}

	params := sqlc.CreateOrderParams{
		PaymentGroupID:  &order.PaymentGroupID,
		ShopID:          &order.ShopID,
		Uuid:            userUUID,
		InvoiceNumber:   order.InvoiceNumber,
		SubTotal:        Float64ToNumeric(order.SubTotal),
		CouponCode:      &order.CouponCode,
		DiscountAmount:  Float64ToNumeric(order.DiscountAmount),
		TaxAmount:       Float64ToNumeric(order.TaxAmount),
		TotalAmount:     Float64ToNumeric(order.TotalAmount),
		Status:          order.Status,
		PaymentStatus:   order.PaymentStatus,
		PaymentMethod:   &order.PaymentMethod,
		ShippingAddress: &order.ShippingAddress,
		ContactPhone:    &order.ContactPhone,
	}

	if order.ExpiresAt != nil {
		params.ExpiresAt = *order.ExpiresAt
	}

	row, err := r.db.CreateOrder(ctx, params)
	if err != nil {
		return entities.Order{}, err
	}

	for _, item := range order.Items {
		if item.Quantity > 2147483647 || item.Quantity < -2147483648 {
			return entities.Order{}, fmt.Errorf("quantity %d overflows int32", item.Quantity)
		}
		sku := item.SKU
		thumb := item.Thumbnail
		_, err = r.db.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID:     row.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Sku:         &sku,
			Thumbnail:   &thumb,
			Quantity:    int32(item.Quantity),
			Price:       Float64ToNumeric(item.Price),
			SubTotal:    Float64ToNumeric(item.SubTotal),
		})
		if err != nil {
			return entities.Order{}, err
		}
	}

	// We only return the ID for now. For a full entity, we would fetch items too,
	// but the service already has the full entity and just needs the ID.
	order.ID = row.Uuid.String()
	return order, nil
}

func (r *orderRepository) List(ctx context.Context, userID string) ([]entities.Order, error) {
	var rows []sqlc.Order
	var err error

	if userID != "" {
		userUUID, errParse := uuid.Parse(userID)
		if errParse != nil {
			return nil, errParse
		}
		rows, err = r.db.ListOrdersByUserID(ctx, userUUID)
	} else {
		rows, err = r.db.ListAllOrders(ctx)
	}

	if err != nil {
		return nil, err
	}

	orders := make([]entities.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, toOrderEntityFromSqlc(row, userID))
	}
	return orders, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (entities.Order, error) {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return entities.Order{}, err
	}
	row, err := r.db.GetOrderByUUID(ctx, orderUUID)
	if err != nil {
		return entities.Order{}, err
	}

	items, err := r.db.GetOrderItemsByOrderID(ctx, row.ID)
	if err != nil {
		return entities.Order{}, err
	}

	order := toOrderEntityFromSqlcRow(row)
	order.Items = toOrderItemsEntityFromSqlc(items)
	return order, nil
}

func (r *orderRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (entities.Order, error) {
	row, err := r.db.GetOrderByInvoiceNumber(ctx, invoiceNumber)
	if err != nil {
		return entities.Order{}, err
	}

	items, err := r.db.GetOrderItemsByOrderID(ctx, row.ID)
	if err != nil {
		return entities.Order{}, err
	}

	order := toOrderEntityFromSqlcInvoiceRow(row)
	order.Items = toOrderItemsEntityFromSqlc(items)
	return order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.db.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		Uuid:   orderUUID,
		Status: status,
	})
}

func (r *orderRepository) UpdateStatusAtomic(ctx context.Context, id string, expectedCurrent string, newStatus string) error {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	rowsAffected, err := r.db.UpdateOrderStatusAtomic(ctx, sqlc.UpdateOrderStatusAtomicParams{
		Uuid:     orderUUID,
		Status:   expectedCurrent,
		Status_2: newStatus,
	})
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows // Using pgx error for "no documents" equivalent
	}
	return nil
}

func (r *orderRepository) ConfirmPaymentAtomic(ctx context.Context, invoiceNumber, expectedCurrentPaymentStatus, newPaymentStatus, newOrderStatus, transactionID string) (int64, error) {
	return r.db.ConfirmPaymentAtomic(ctx, sqlc.ConfirmPaymentAtomicParams{
		InvoiceNumber:        invoiceNumber,
		PaymentStatus:        newPaymentStatus,
		Status:               newOrderStatus,
		PaymentTransactionID: &transactionID,
	})
}

func (r *orderRepository) MarkInventoryReturned(ctx context.Context, id string) error {
	orderUUID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	rowsAffected, err := r.db.MarkInventoryReturned(ctx, orderUUID)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return pgx.ErrNoRows // Either already marked or not found
	}
	return nil
}

func (r *orderRepository) GetCustomerInfoByUserID(ctx context.Context, userID string) (string, string, error) {
	uUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", "", err
	}
	// We can use the existing GetUserByUUID and GetUserProfileByUUID from sqlc
	user, err := r.db.GetUserByUUID(ctx, uUUID)
	if err != nil {
		return "", "", err
	}
	profile, err := r.db.GetUserProfileByUUID(ctx, uUUID)
	var name string
	if err == nil {
		if profile.FullName != nil && *profile.FullName != "" {
			name = *profile.FullName
		} else if profile.UserName != nil {
			name = *profile.UserName
		}
	}
	return user.Email, name, nil
}

func (r *orderRepository) GetExpiredPendingOrders(ctx context.Context) ([]entities.Order, error) {
	rows, err := r.db.GetExpiredPendingOrders(ctx)
	if err != nil {
		return nil, err
	}

	orders := make([]entities.Order, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, toOrderEntityFromSqlc(row, "")) // User ID not mapped in this simple sweep query, but we only need order ID anyway
	}
	return orders, nil
}

// Helpers
func toOrderEntityFromSqlc(row sqlc.Order, userID string) entities.Order {
	var shopID, paymentGroupID, couponCode, paymentMethod, shipping, phone, txID string
	if row.ShopID != nil {
		shopID = *row.ShopID
	}
	if row.PaymentGroupID != nil {
		paymentGroupID = *row.PaymentGroupID
	}
	if row.CouponCode != nil {
		couponCode = *row.CouponCode
	}
	if row.PaymentMethod != nil {
		paymentMethod = *row.PaymentMethod
	}
	if row.ShippingAddress != nil {
		shipping = *row.ShippingAddress
	}
	if row.ContactPhone != nil {
		phone = *row.ContactPhone
	}
	if row.PaymentTransactionID != nil {
		txID = *row.PaymentTransactionID
	}

	var txIDPtr *string
	if txID != "" {
		txIDPtr = &txID
	}

	subTotal, _ := row.SubTotal.Float64Value()
	discount, _ := row.DiscountAmount.Float64Value()
	tax, _ := row.TaxAmount.Float64Value()
	total, _ := row.TotalAmount.Float64Value()

	var expiresAt *time.Time
	if !row.ExpiresAt.IsZero() {
		t := row.ExpiresAt
		expiresAt = &t
	}

	return entities.Order{
		ID:                   row.Uuid.String(),
		PaymentGroupID:       paymentGroupID,
		ShopID:               shopID,
		UserID:               userID,
		InvoiceNumber:        row.InvoiceNumber,
		SubTotal:             subTotal.Float64,
		CouponCode:           couponCode,
		DiscountAmount:       discount.Float64,
		TaxAmount:            tax.Float64,
		TotalAmount:          total.Float64,
		Status:               row.Status,
		PaymentStatus:        row.PaymentStatus,
		PaymentMethod:        paymentMethod,
		PaymentTransactionID: txIDPtr,
		ShippingAddress:      shipping,
		ContactPhone:         phone,
		IsDeleted:            row.IsDeleted,
		InventoryReturned:    row.InventoryReturned,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		ExpiresAt:            expiresAt,
	}
}

func toOrderEntityFromSqlcRow(row sqlc.GetOrderByUUIDRow) entities.Order {
	var shopID, paymentGroupID, couponCode, paymentMethod, shipping, phone, txID string
	if row.ShopID != nil {
		shopID = *row.ShopID
	}
	if row.PaymentGroupID != nil {
		paymentGroupID = *row.PaymentGroupID
	}
	if row.CouponCode != nil {
		couponCode = *row.CouponCode
	}
	if row.PaymentMethod != nil {
		paymentMethod = *row.PaymentMethod
	}
	if row.ShippingAddress != nil {
		shipping = *row.ShippingAddress
	}
	if row.ContactPhone != nil {
		phone = *row.ContactPhone
	}
	if row.PaymentTransactionID != nil {
		txID = *row.PaymentTransactionID
	}

	var txIDPtr *string
	if txID != "" {
		txIDPtr = &txID
	}

	subTotal, _ := row.SubTotal.Float64Value()
	discount, _ := row.DiscountAmount.Float64Value()
	tax, _ := row.TaxAmount.Float64Value()
	total, _ := row.TotalAmount.Float64Value()

	var expiresAt *time.Time
	if !row.ExpiresAt.IsZero() {
		t := row.ExpiresAt
		expiresAt = &t
	}

	return entities.Order{
		ID:                   row.Uuid.String(),
		PaymentGroupID:       paymentGroupID,
		ShopID:               shopID,
		UserID:               row.UserUuid.String(),
		InvoiceNumber:        row.InvoiceNumber,
		SubTotal:             subTotal.Float64,
		CouponCode:           couponCode,
		DiscountAmount:       discount.Float64,
		TaxAmount:            tax.Float64,
		TotalAmount:          total.Float64,
		Status:               row.Status,
		PaymentStatus:        row.PaymentStatus,
		PaymentMethod:        paymentMethod,
		PaymentTransactionID: txIDPtr,
		ShippingAddress:      shipping,
		ContactPhone:         phone,
		IsDeleted:            row.IsDeleted,
		InventoryReturned:    row.InventoryReturned,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		ExpiresAt:            expiresAt,
	}
}

func toOrderEntityFromSqlcInvoiceRow(row sqlc.GetOrderByInvoiceNumberRow) entities.Order {
	var shopID, paymentGroupID, couponCode, paymentMethod, shipping, phone, txID string
	if row.ShopID != nil {
		shopID = *row.ShopID
	}
	if row.PaymentGroupID != nil {
		paymentGroupID = *row.PaymentGroupID
	}
	if row.CouponCode != nil {
		couponCode = *row.CouponCode
	}
	if row.PaymentMethod != nil {
		paymentMethod = *row.PaymentMethod
	}
	if row.ShippingAddress != nil {
		shipping = *row.ShippingAddress
	}
	if row.ContactPhone != nil {
		phone = *row.ContactPhone
	}
	if row.PaymentTransactionID != nil {
		txID = *row.PaymentTransactionID
	}

	var txIDPtr *string
	if txID != "" {
		txIDPtr = &txID
	}

	subTotal, _ := row.SubTotal.Float64Value()
	discount, _ := row.DiscountAmount.Float64Value()
	tax, _ := row.TaxAmount.Float64Value()
	total, _ := row.TotalAmount.Float64Value()

	var expiresAt *time.Time
	if !row.ExpiresAt.IsZero() {
		t := row.ExpiresAt
		expiresAt = &t
	}

	return entities.Order{
		ID:                   row.Uuid.String(),
		PaymentGroupID:       paymentGroupID,
		ShopID:               shopID,
		UserID:               row.UserUuid.String(),
		InvoiceNumber:        row.InvoiceNumber,
		SubTotal:             subTotal.Float64,
		CouponCode:           couponCode,
		DiscountAmount:       discount.Float64,
		TaxAmount:            tax.Float64,
		TotalAmount:          total.Float64,
		Status:               row.Status,
		PaymentStatus:        row.PaymentStatus,
		PaymentMethod:        paymentMethod,
		PaymentTransactionID: txIDPtr,
		ShippingAddress:      shipping,
		ContactPhone:         phone,
		IsDeleted:            row.IsDeleted,
		InventoryReturned:    row.InventoryReturned,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		ExpiresAt:            expiresAt,
	}
}

func toOrderItemsEntityFromSqlc(items []sqlc.OrderItem) []entities.OrderItem {
	result := make([]entities.OrderItem, 0, len(items))
	for _, item := range items {
		var sku, thumb string
		if item.Sku != nil {
			sku = *item.Sku
		}
		if item.Thumbnail != nil {
			thumb = *item.Thumbnail
		}

		price, _ := item.Price.Float64Value()
		subTotal, _ := item.SubTotal.Float64Value()

		result = append(result, entities.OrderItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			SKU:         sku,
			Thumbnail:   thumb,
			Quantity:    int64(item.Quantity),
			Price:       price.Float64,
			SubTotal:    subTotal.Float64,
		})
	}
	return result
}

func Float64ToNumeric(f float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(f) // Can also use num.Int/Exp logic or pgtype.Numeric{} init but Scan works for float64/string in pgtype v5 if DB returns it. Wait, actually pgtype.Numeric has `ScanFloat64` or we can convert string to numeric.
	// Actually for pgtype.Numeric, setting from float is:
	_ = num.Scan(fmt.Sprintf("%f", f))
	return num
}
