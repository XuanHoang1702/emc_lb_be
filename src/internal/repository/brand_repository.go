package repository

import (
	"context"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BrandRepository interface {
	Create(context.Context, entities.Brand) (entities.Brand, error)
	List(context.Context) ([]entities.Brand, error)
	GetByID(context.Context, string) (entities.Brand, error)
	Update(context.Context, string, map[string]any) (entities.Brand, error)
	Delete(context.Context, string, map[string]any) error
	ExistsByName(context.Context, string, *string) (bool, error)
	ExistsBySlug(context.Context, string, *string) (bool, error)
	ExistsByPosition(context.Context, int64, *string) (bool, error)
	ExistsByEmail(context.Context, string, *string) (bool, error)
	EnsureIndexes(context.Context) error
}

type brandRepository struct {
	collection *mongo.Collection
}

func NewBrandRepository(collection *mongo.Collection) BrandRepository {
	return &brandRepository{collection: collection}
}

type brandDoc struct {
	ID bson.ObjectID `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Slug        string `bson:"slug"`
	Description string `bson:"description,omitempty"`
	Logo   string `bson:"logo,omitempty"`
	Banner string `bson:"banner,omitempty"`
	Website     string `bson:"website,omitempty"`
	Email       string `bson:"email,omitempty"`
	Phone       string `bson:"phone,omitempty"`
	Country     string `bson:"country,omitempty"`
	CompanyName string `bson:"company_name,omitempty"`
	MetaTitle       string   `bson:"meta_title,omitempty"`
	MetaDescription string   `bson:"meta_description,omitempty"`
	MetaKeywords    []string `bson:"meta_keywords,omitempty"`
	Position   int64 `bson:"position"`
	IsFeatured bool  `bson:"is_featured"`
	Status    string `bson:"status"`
	IsDeleted bool   `bson:"is_deleted"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at"`
}

func toBrandDoc(b entities.Brand) brandDoc {
	doc := brandDoc{
		Name: b.Name,
		Slug: b.Slug,
		Description: b.Description,
		Logo: b.Logo,
		Banner: b.Banner,
		Website: b.Website,
		Email: b.Email,
		Phone: b.Phone,
		Country: b.Country,
		CompanyName: b.CompanyName,
		MetaTitle: b.MetaTitle,
		MetaDescription: b.MetaDescription,
		MetaKeywords: b.MetaKeywords,
		Position: b.Position,
		IsFeatured: b.IsFeatured,
		Status: b.Status,
		IsDeleted: b.IsDeleted,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
		DeletedAt: b.DeletedAt,
	}
	if b.ID != "" {
		if id, err := bson.ObjectIDFromHex(b.ID); err == nil {
			doc.ID = id
		}
	}
	return doc
}

func toBrandEntity(doc brandDoc) entities.Brand {
	return entities.Brand{
		ID: doc.ID.Hex(),
		Name: doc.Name,
		Slug: doc.Slug,
		Description: doc.Description,
		Logo: doc.Logo,
		Banner: doc.Banner,
		Website: doc.Website,
		Email: doc.Email,
		Phone: doc.Phone,
		Country: doc.Country,
		CompanyName: doc.CompanyName,
		MetaTitle: doc.MetaTitle,
		MetaDescription: doc.MetaDescription,
		MetaKeywords: doc.MetaKeywords,
		Position: doc.Position,
		IsFeatured: doc.IsFeatured,
		Status: doc.Status,
		IsDeleted: doc.IsDeleted,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
		DeletedAt: doc.DeletedAt,
	}
}

func (r *brandRepository) Create(ctx context.Context, brand entities.Brand) (entities.Brand, error) {
	doc := toBrandDoc(brand)
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return entities.Brand{}, err
	}

	if id, ok := result.InsertedID.(bson.ObjectID); ok {
		doc.ID = id
	}

	return toBrandEntity(doc), nil
}

func (r *brandRepository) List(ctx context.Context) ([]entities.Brand, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "position", Value: 1},
		{Key: "created_at", Value: -1},
	})

	cursor, err := r.collection.Find(ctx, bson.M{"is_deleted": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	brands := make([]entities.Brand, 0)
	for cursor.Next(ctx) {
		var doc brandDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		brands = append(brands, toBrandEntity(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}

func (r *brandRepository) GetByID(ctx context.Context, id string) (entities.Brand, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Brand{}, err
	}
	var doc brandDoc
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objID,
		"is_deleted": false,
	}).Decode(&doc)
	if err != nil {
		return entities.Brand{}, err
	}

	return toBrandEntity(doc), nil
}

func (r *brandRepository) Update(ctx context.Context, id string, update map[string]any) (entities.Brand, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Brand{}, err
	}
	var doc brandDoc
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID, "is_deleted": false},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		return entities.Brand{}, err
	}

	return toBrandEntity(doc), nil
}

func (r *brandRepository) Delete(ctx context.Context, id string, update map[string]any) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result, err := r.collection.UpdateOne(ctx, bson.M{
		"_id":        objID,
		"is_deleted": false,
	}, bson.M{"$set": update})
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}

func (r *brandRepository) ExistsByName(ctx context.Context, name string, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "name", name, excludeID)
}

func (r *brandRepository) ExistsBySlug(ctx context.Context, slug string, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "slug", slug, excludeID)
}

func (r *brandRepository) ExistsByPosition(ctx context.Context, position int64, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "position", position, excludeID)
}

func (r *brandRepository) ExistsByEmail(ctx context.Context, email string, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "email", email, excludeID)
}

func (r *brandRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "name", Value: 1}},
			Options: options.Index().
				SetName("idx_brands_name_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
		{
			Keys: bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().
				SetName("idx_brands_slug_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
		{
			Keys: bson.D{{Key: "position", Value: 1}},
			Options: options.Index().
				SetName("idx_brands_position_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetName("idx_brands_email_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"is_deleted": false,
					"email":      bson.M{"$type": "string", "$ne": ""},
				}),
		},
		{
			Keys: bson.D{
				{Key: "is_deleted", Value: 1},
				{Key: "position", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_brands_listing"),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, models)
	return err
}

func (r *brandRepository) existsByField(ctx context.Context, field string, value any, excludeID *string) (bool, error) {
	filter := bson.M{
		field:        value,
		"is_deleted": false,
	}
	if excludeID != nil {
		if id, err := bson.ObjectIDFromHex(*excludeID); err == nil {
			filter["_id"] = bson.M{"$ne": id}
		}
	}

	err := r.collection.FindOne(ctx, filter).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func ToBrandResponse(brand entities.Brand) entities.BrandResponse {
	return entities.BrandResponse{
		ID:              brand.ID,
		Name:            brand.Name,
		Slug:            brand.Slug,
		Description:     brand.Description,
		Logo:            brand.Logo,
		Banner:          brand.Banner,
		Website:         brand.Website,
		Email:           brand.Email,
		Phone:           brand.Phone,
		Country:         brand.Country,
		CompanyName:     brand.CompanyName,
		MetaTitle:       brand.MetaTitle,
		MetaDescription: brand.MetaDescription,
		MetaKeywords:    brand.MetaKeywords,
		Position:        brand.Position,
		IsFeatured:      brand.IsFeatured,
		Status:          brand.Status,
		CreatedAt:       brand.CreatedAt,
		UpdatedAt:       brand.UpdatedAt,
	}
}
