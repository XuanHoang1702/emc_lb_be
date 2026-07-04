package repository

import (
	"context"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductRepository interface {
	Create(context.Context, entities.Product) (entities.Product, error)
	List(context.Context) ([]entities.Product, error)
}

type productRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(collection *mongo.Collection) ProductRepository {
	return &productRepository{collection: collection}
}

func (r *productRepository) Create(ctx context.Context, product entities.Product) (entities.Product, error) {
	doc := toProductDoc(product)
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return entities.Product{}, err
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if ok {
		doc.ID = id
	}

	return toProductEntity(doc), nil
}

func (r *productRepository) List(ctx context.Context) ([]entities.Product, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	products := make([]entities.Product, 0)
	for cursor.Next(ctx) {
		var doc productDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		products = append(products, toProductEntity(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// ============================================================================
// Database Models & Mappers
// ============================================================================

type productDoc struct {
	ID bson.ObjectID `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Slug        string `bson:"slug"`
	Description string `bson:"description,omitempty"`
	ShortDesc   string `bson:"short_desc,omitempty"`
	Price         float64 `bson:"price"`
	OriginalPrice float64 `bson:"original_price,omitempty"`
	CostPrice     float64 `bson:"cost_price,omitempty"`
	SKU            string `bson:"sku"`
	Barcode        string `bson:"barcode,omitempty"`
	Stock          int64  `bson:"stock"`
	SoldCount      int64  `bson:"sold_count"`
	AllowBackorder bool   `bson:"allow_backorder"`
	Thumbnail string   `bson:"thumbnail,omitempty"`
	Images    []string `bson:"images,omitempty"`
	CategoryID bson.ObjectID `bson:"category_id,omitempty"`
	BrandID    bson.ObjectID `bson:"brand_id,omitempty"`
	Attributes map[string]string `bson:"attributes,omitempty"`
	Tags       []string          `bson:"tags,omitempty"`
	Weight float64 `bson:"weight,omitempty"`
	Length float64 `bson:"length,omitempty"`
	Width  float64 `bson:"width,omitempty"`
	Height float64 `bson:"height,omitempty"`
	MetaTitle       string `bson:"meta_title,omitempty"`
	MetaDescription string `bson:"meta_description,omitempty"`
	Status     string `bson:"status"`
	IsFeatured bool   `bson:"is_featured"`
	IsDeleted  bool   `bson:"is_deleted"`
	AverageRating float64 `bson:"average_rating,omitempty"`
	ReviewCount   int64   `bson:"review_count,omitempty"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func toProductDoc(p entities.Product) productDoc {
	doc := productDoc{
		Name: p.Name,
		Slug: p.Slug,
		Description: p.Description,
		ShortDesc: p.ShortDesc,
		Price: p.Price,
		OriginalPrice: p.OriginalPrice,
		CostPrice: p.CostPrice,
		SKU: p.SKU,
		Barcode: p.Barcode,
		Stock: p.Stock,
		SoldCount: p.SoldCount,
		AllowBackorder: p.AllowBackorder,
		Thumbnail: p.Thumbnail,
		Images: p.Images,
		Attributes: p.Attributes,
		Tags: p.Tags,
		Weight: p.Weight,
		Length: p.Length,
		Width: p.Width,
		Height: p.Height,
		MetaTitle: p.MetaTitle,
		MetaDescription: p.MetaDescription,
		Status: p.Status,
		IsFeatured: p.IsFeatured,
		IsDeleted: p.IsDeleted,
		AverageRating: p.AverageRating,
		ReviewCount: p.ReviewCount,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if p.ID != "" {
		if objID, err := bson.ObjectIDFromHex(p.ID); err == nil {
			doc.ID = objID
		}
	}
	if p.CategoryID != "" {
		if catID, err := bson.ObjectIDFromHex(p.CategoryID); err == nil {
			doc.CategoryID = catID
		}
	}
	if p.BrandID != "" {
		if brandID, err := bson.ObjectIDFromHex(p.BrandID); err == nil {
			doc.BrandID = brandID
		}
	}
	return doc
}

func toProductEntity(doc productDoc) entities.Product {
	catID := ""
	if !doc.CategoryID.IsZero() {
		catID = doc.CategoryID.Hex()
	}
	brandID := ""
	if !doc.BrandID.IsZero() {
		brandID = doc.BrandID.Hex()
	}
	return entities.Product{
		ID: doc.ID.Hex(),
		Name: doc.Name,
		Slug: doc.Slug,
		Description: doc.Description,
		ShortDesc: doc.ShortDesc,
		Price: doc.Price,
		OriginalPrice: doc.OriginalPrice,
		CostPrice: doc.CostPrice,
		SKU: doc.SKU,
		Barcode: doc.Barcode,
		Stock: doc.Stock,
		SoldCount: doc.SoldCount,
		AllowBackorder: doc.AllowBackorder,
		Thumbnail: doc.Thumbnail,
		Images: doc.Images,
		CategoryID: catID,
		BrandID: brandID,
		Attributes: doc.Attributes,
		Tags: doc.Tags,
		Weight: doc.Weight,
		Length: doc.Length,
		Width: doc.Width,
		Height: doc.Height,
		MetaTitle: doc.MetaTitle,
		MetaDescription: doc.MetaDescription,
		Status: doc.Status,
		IsFeatured: doc.IsFeatured,
		IsDeleted: doc.IsDeleted,
		AverageRating: doc.AverageRating,
		ReviewCount: doc.ReviewCount,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}
}



