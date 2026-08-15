package repository

import (
	"context"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CouponRepository interface {
	Create(ctx context.Context, coupon entities.Coupon) (entities.Coupon, error)
	GetByID(ctx context.Context, id string) (entities.Coupon, error)
	GetByCode(ctx context.Context, code string) (entities.Coupon, error)
	List(ctx context.Context) ([]entities.Coupon, error)
	Update(ctx context.Context, id string, update map[string]any) (entities.Coupon, error)
	IncrementUsage(ctx context.Context, code string, count int64) error
}

type couponRepository struct {
	collection *mongo.Collection
}

func NewCouponRepository(collection *mongo.Collection) CouponRepository {
	return &couponRepository{collection: collection}
}

type couponDoc struct {
	ID                bson.ObjectID `bson:"_id,omitempty"`
	Code              string        `bson:"code"`
	Type              string        `bson:"type"`
	Value             float64       `bson:"value"`
	MinOrderAmount    float64       `bson:"min_order_amount"`
	MaxDiscountAmount float64       `bson:"max_discount_amount"`
	UsageLimit        int64         `bson:"usage_limit"`
	UsageCount        int64         `bson:"usage_count"`
	StartDate         time.Time     `bson:"start_date"`
	EndDate           time.Time     `bson:"end_date"`
	IsActive          bool          `bson:"is_active"`
	IsDeleted         bool          `bson:"is_deleted"`
	CreatedAt         time.Time     `bson:"created_at"`
	UpdatedAt         time.Time     `bson:"updated_at"`
}

func toCouponDoc(c entities.Coupon) couponDoc {
	doc := couponDoc{
		Code:              c.Code,
		Type:              c.Type,
		Value:             c.Value,
		MinOrderAmount:    c.MinOrderAmount,
		MaxDiscountAmount: c.MaxDiscountAmount,
		UsageLimit:        c.UsageLimit,
		UsageCount:        c.UsageCount,
		StartDate:         c.StartDate,
		EndDate:           c.EndDate,
		IsActive:          c.IsActive,
		IsDeleted:         c.IsDeleted,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
	if c.ID != "" {
		if objID, err := bson.ObjectIDFromHex(c.ID); err == nil {
			doc.ID = objID
		}
	}
	return doc
}

func toCouponEntity(doc couponDoc) entities.Coupon {
	return entities.Coupon{
		ID:                doc.ID.Hex(),
		Code:              doc.Code,
		Type:              doc.Type,
		Value:             doc.Value,
		MinOrderAmount:    doc.MinOrderAmount,
		MaxDiscountAmount: doc.MaxDiscountAmount,
		UsageLimit:        doc.UsageLimit,
		UsageCount:        doc.UsageCount,
		StartDate:         doc.StartDate,
		EndDate:           doc.EndDate,
		IsActive:          doc.IsActive,
		IsDeleted:         doc.IsDeleted,
		CreatedAt:         doc.CreatedAt,
		UpdatedAt:         doc.UpdatedAt,
	}
}

func (r *couponRepository) Create(ctx context.Context, coupon entities.Coupon) (entities.Coupon, error) {
	doc := toCouponDoc(coupon)
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return entities.Coupon{}, err
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if ok {
		doc.ID = id
	}

	return toCouponEntity(doc), nil
}

func (r *couponRepository) GetByID(ctx context.Context, id string) (entities.Coupon, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Coupon{}, err
	}

	var doc couponDoc
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objID,
		"is_deleted": false,
	}).Decode(&doc)

	if err != nil {
		return entities.Coupon{}, err
	}
	return toCouponEntity(doc), nil
}

func (r *couponRepository) GetByCode(ctx context.Context, code string) (entities.Coupon, error) {
	var doc couponDoc
	err := r.collection.FindOne(ctx, bson.M{
		"code":       code,
		"is_deleted": false,
	}).Decode(&doc)

	if err != nil {
		return entities.Coupon{}, err
	}
	return toCouponEntity(doc), nil
}

func (r *couponRepository) List(ctx context.Context) ([]entities.Coupon, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"is_deleted": false}, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var coupons []entities.Coupon
	for cursor.Next(ctx) {
		var doc couponDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		coupons = append(coupons, toCouponEntity(doc))
	}

	return coupons, nil
}

func (r *couponRepository) Update(ctx context.Context, id string, update map[string]any) (entities.Coupon, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Coupon{}, err
	}

	var doc couponDoc
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID, "is_deleted": false},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)

	if err != nil {
		return entities.Coupon{}, err
	}

	return toCouponEntity(doc), nil
}

func (r *couponRepository) IncrementUsage(ctx context.Context, code string, count int64) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{
			"code":       code,
			"is_deleted": false,
			"$or": []bson.M{
				{"usage_limit": 0},
				{"$expr": bson.M{"$lt": bson.A{"$usage_count", "$usage_limit"}}},
			},
		},
		bson.M{"$inc": bson.M{"usage_count": count}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
