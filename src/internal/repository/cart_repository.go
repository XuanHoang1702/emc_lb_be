package repository

import (
	"context"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CartRepository interface {
	GetByUserID(ctx context.Context, userID string) (entities.Cart, error)
	Save(ctx context.Context, cart entities.Cart) (entities.Cart, error)
	ClearCart(ctx context.Context, userID string) error
	EnsureIndexes(ctx context.Context) error
}

type cartRepository struct {
	collection *mongo.Collection
}

func NewCartRepository(collection *mongo.Collection) CartRepository {
	return &cartRepository{collection: collection}
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID string) (entities.Cart, error) {
	var doc cartDoc
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&doc)
	if err == mongo.ErrNoDocuments {
		// Return empty cart if not found
		return entities.Cart{
			UserID: userID,
			Items:  []entities.CartItem{},
		}, nil
	}
	if err != nil {
		return entities.Cart{}, err
	}
	return toCartEntity(doc), nil
}

func (r *cartRepository) Save(ctx context.Context, cart entities.Cart) (entities.Cart, error) {
	doc := toCartDoc(cart)
	doc.UpdatedAt = time.Now().UTC()

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	
	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = doc.UpdatedAt
	}

	update := bson.M{
		"$set": bson.M{
			"items":      doc.Items,
			"updated_at": doc.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"user_id":    doc.UserID,
			"created_at": doc.CreatedAt,
		},
	}

	var updatedDoc cartDoc
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"user_id": doc.UserID},
		update,
		opts,
	).Decode(&updatedDoc)

	if err != nil {
		return entities.Cart{}, err
	}

	return toCartEntity(updatedDoc), nil
}

func (r *cartRepository) ClearCart(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}

func (r *cartRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().
				SetName("idx_cart_user_id_unique").
				SetUnique(true),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, models)
	return err
}

// ============================================================================
// Database Models & Mappers
// ============================================================================

type cartItemDoc struct {
	ProductID string `bson:"product_id"`
	Quantity  int64  `bson:"quantity"`
}

type cartDoc struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"user_id"`
	Items     []cartItemDoc `bson:"items"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

func toCartDoc(c entities.Cart) cartDoc {
	items := make([]cartItemDoc, 0, len(c.Items))
	for _, item := range c.Items {
		items = append(items, cartItemDoc{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	doc := cartDoc{
		UserID:    c.UserID,
		Items:     items,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}

	if c.ID != "" {
		if id, err := bson.ObjectIDFromHex(c.ID); err == nil {
			doc.ID = id
		}
	}
	return doc
}

func toCartEntity(doc cartDoc) entities.Cart {
	items := make([]entities.CartItem, 0, len(doc.Items))
	for _, item := range doc.Items {
		items = append(items, entities.CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return entities.Cart{
		ID:        doc.ID.Hex(),
		UserID:    doc.UserID,
		Items:     items,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}
