package repository

import (
	"context"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrderRepository interface {
	Create(context.Context, entities.Order) (entities.Order, error)
	List(context.Context, string) ([]entities.Order, error)
	GetByID(context.Context, string) (entities.Order, error)
	GetByInvoiceNumber(context.Context, string) (entities.Order, error)
	UpdateStatus(context.Context, string, string) error
	UpdatePaymentStatus(context.Context, string, string) error
	EnsureIndexes(context.Context) error
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(collection *mongo.Collection) OrderRepository {
	return &orderRepository{collection: collection}
}

func (r *orderRepository) Create(ctx context.Context, order entities.Order) (entities.Order, error) {
	doc := toOrderDoc(order)
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return entities.Order{}, err
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if ok {
		doc.ID = id
	}

	return toOrderEntity(doc), nil
}

func (r *orderRepository) List(ctx context.Context, userID string) ([]entities.Order, error) {
	filter := bson.M{"is_deleted": false}
	if userID != "" {
		filter["user_id"] = userID
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	orders := make([]entities.Order, 0)
	for cursor.Next(ctx) {
		var doc orderDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		orders = append(orders, toOrderEntity(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (entities.Order, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Order{}, err
	}
	var doc orderDoc
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objID,
		"is_deleted": false,
	}).Decode(&doc)
	if err != nil {
		return entities.Order{}, err
	}

	return toOrderEntity(doc), nil
}

func (r *orderRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (entities.Order, error) {
	var doc orderDoc
	err := r.collection.FindOne(ctx, bson.M{
		"invoice_number": invoiceNumber,
		"is_deleted":     false,
	}).Decode(&doc)
	if err != nil {
		return entities.Order{}, err
	}
	return toOrderEntity(doc), nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		},
	})
	return err
}

func (r *orderRepository) UpdatePaymentStatus(ctx context.Context, id string, paymentStatus string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"payment_status": paymentStatus,
			"updated_at":     time.Now().UTC(),
		},
	})
	return err
}

func (r *orderRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "invoice_number", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// ============================================================================
// Database Models & Mappers
// ============================================================================

type orderDoc struct {
	ID              bson.ObjectID  `bson:"_id,omitempty"`
	PaymentGroupID  string         `bson:"payment_group_id,omitempty"`
	ShopID          string         `bson:"shop_id,omitempty"`
	UserID          string         `bson:"user_id"`
	InvoiceNumber   string         `bson:"invoice_number"`
	Items           []orderItemDoc `bson:"items"`
	SubTotal        float64        `bson:"sub_total"`
	CouponCode      string         `bson:"coupon_code,omitempty"`
	DiscountAmount  float64        `bson:"discount_amount"`
	TaxAmount       float64        `bson:"tax_amount"`
	TotalAmount     float64        `bson:"total_amount"`
	Status          string         `bson:"status"`
	PaymentStatus   string         `bson:"payment_status"`
	PaymentMethod   string         `bson:"payment_method"`
	ShippingAddress string         `bson:"shipping_address"`
	ContactPhone    string         `bson:"contact_phone"`
	IsDeleted       bool           `bson:"is_deleted"`
	CreatedAt       time.Time      `bson:"created_at"`
	UpdatedAt       time.Time      `bson:"updated_at"`
}

type orderItemDoc struct {
	ProductID string  `bson:"product_id"`
	Quantity  int64   `bson:"quantity"`
	Price     float64 `bson:"price"`
}

func toOrderDoc(o entities.Order) orderDoc {
	doc := orderDoc{
		PaymentGroupID:  o.PaymentGroupID,
		ShopID:          o.ShopID,
		UserID:          o.UserID,
		InvoiceNumber:   o.InvoiceNumber,
		SubTotal:        o.SubTotal,
		CouponCode:      o.CouponCode,
		DiscountAmount:  o.DiscountAmount,
		TaxAmount:       o.TaxAmount,
		TotalAmount:     o.TotalAmount,
		Status:          o.Status,
		PaymentStatus:   o.PaymentStatus,
		PaymentMethod:   o.PaymentMethod,
		ShippingAddress: o.ShippingAddress,
		ContactPhone:    o.ContactPhone,
		IsDeleted:       o.IsDeleted,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
	if o.ID != "" {
		if objID, err := bson.ObjectIDFromHex(o.ID); err == nil {
			doc.ID = objID
		}
	}

	items := make([]orderItemDoc, len(o.Items))
	for i, item := range o.Items {
		items[i] = orderItemDoc{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}
	doc.Items = items

	return doc
}

func toOrderEntity(doc orderDoc) entities.Order {
	items := make([]entities.OrderItem, len(doc.Items))
	for i, item := range doc.Items {
		items[i] = entities.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return entities.Order{
		ID:              doc.ID.Hex(),
		PaymentGroupID:  doc.PaymentGroupID,
		ShopID:          doc.ShopID,
		UserID:          doc.UserID,
		InvoiceNumber:   doc.InvoiceNumber,
		Items:           items,
		SubTotal:        doc.SubTotal,
		CouponCode:      doc.CouponCode,
		DiscountAmount:  doc.DiscountAmount,
		TaxAmount:       doc.TaxAmount,
		TotalAmount:     doc.TotalAmount,
		Status:          doc.Status,
		PaymentStatus:   doc.PaymentStatus,
		PaymentMethod:   doc.PaymentMethod,
		ShippingAddress: doc.ShippingAddress,
		ContactPhone:    doc.ContactPhone,
		IsDeleted:       doc.IsDeleted,
		CreatedAt:       doc.CreatedAt,
		UpdatedAt:       doc.UpdatedAt,
	}
}
