package repository

import (
	"context"
	"errors"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ShopRepository interface {
	Create(ctx context.Context, shop entities.Shop) (entities.Shop, error)
	GetByID(ctx context.Context, id string) (entities.Shop, error)
	GetByOwnerID(ctx context.Context, ownerID string) (entities.Shop, error)
	Update(ctx context.Context, id string, update map[string]any) (entities.Shop, error)
}

type shopRepository struct {
	collection *mongo.Collection
}

func NewShopRepository(collection *mongo.Collection) ShopRepository {
	return &shopRepository{collection: collection}
}

func (r *shopRepository) Create(ctx context.Context, shop entities.Shop) (entities.Shop, error) {
	shop.CreatedAt = time.Now()
	shop.UpdatedAt = time.Now()
	shop.IsDeleted = false

	_, err := r.collection.InsertOne(ctx, shop)
	if err != nil {
		return entities.Shop{}, err
	}
	return shop, nil
}

func (r *shopRepository) GetByID(ctx context.Context, id string) (entities.Shop, error) {
	var shop entities.Shop
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "is_deleted": false}).Decode(&shop)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.Shop{}, errors.New("shop not found")
		}
		return entities.Shop{}, err
	}
	return shop, nil
}

func (r *shopRepository) GetByOwnerID(ctx context.Context, ownerID string) (entities.Shop, error) {
	var shop entities.Shop
	err := r.collection.FindOne(ctx, bson.M{"owner_id": ownerID, "is_deleted": false}).Decode(&shop)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.Shop{}, errors.New("shop not found")
		}
		return entities.Shop{}, err
	}
	return shop, nil
}

func (r *shopRepository) Update(ctx context.Context, id string, update map[string]any) (entities.Shop, error) {
	update["updated_at"] = time.Now()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedShop entities.Shop
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id, "is_deleted": false},
		bson.M{"$set": update},
		opts,
	).Decode(&updatedShop)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.Shop{}, errors.New("shop not found")
		}
		return entities.Shop{}, err
	}
	return updatedShop, nil
}
