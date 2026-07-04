package repository

import (
	"context"
	"time"

	"emc_lb/src/pkg/entities"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CategoryRepository interface {
	Create(context.Context, entities.Category) (entities.Category, error)
	List(context.Context) ([]entities.Category, error)
	GetByID(context.Context, string) (entities.Category, error)
	Update(context.Context, string, map[string]any) (entities.Category, error)
	Delete(context.Context, string, map[string]any) error
	ExistsByID(context.Context, string) (bool, error)
	ExistsByName(context.Context, string, *string) (bool, error)
	ExistsBySlug(context.Context, string, *string) (bool, error)
	ExistsByPosition(context.Context, int64, *string) (bool, error)
	EnsureIndexes(context.Context) error
}

type categoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository(collection *mongo.Collection) CategoryRepository {
	return &categoryRepository{collection: collection}
}


func (r *categoryRepository) Create(ctx context.Context, category entities.Category) (entities.Category, error) {
	doc := toCategoryDoc(category)
	result, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return entities.Category{}, err
	}

	id, ok := result.InsertedID.(bson.ObjectID)
	if ok {
		doc.ID = id
	}

	return toCategoryEntity(doc), nil
}

func (r *categoryRepository) List(ctx context.Context) ([]entities.Category, error) {
	opts := options.Find().SetSort(bson.D{
		{Key: "position", Value: 1},
		{Key: "created_at", Value: -1},
	})

	cursor, err := r.collection.Find(ctx, bson.M{"is_deleted": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	categories := make([]entities.Category, 0)
	for cursor.Next(ctx) {
		var doc categoryDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		categories = append(categories, toCategoryEntity(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (entities.Category, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Category{}, err
	}
	var doc categoryDoc
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objID,
		"is_deleted": false,
	}).Decode(&doc)
	if err != nil {
		return entities.Category{}, err
	}

	return toCategoryEntity(doc), nil
}

func (r *categoryRepository) Update(ctx context.Context, id string, update map[string]any) (entities.Category, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return entities.Category{}, err
	}
	var doc categoryDoc
	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID, "is_deleted": false},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if err != nil {
		return entities.Category{}, err
	}

	return toCategoryEntity(doc), nil
}

func (r *categoryRepository) Delete(ctx context.Context, id string, update map[string]any) error {
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


func (r *categoryRepository) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "name", Value: 1}},
			Options: options.Index().
				SetName("idx_categories_name_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
		{
			Keys: bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().
				SetName("idx_categories_slug_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
		{
			Keys: bson.D{{Key: "position", Value: 1}},
			Options: options.Index().
				SetName("idx_categories_position_unique").
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
		{
			Keys: bson.D{
				{Key: "is_deleted", Value: 1},
				{Key: "position", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_categories_listing"),
		},
		{
			Keys: bson.D{{Key: "parent_id", Value: 1}},
			Options: options.Index().
				SetName("idx_categories_parent_id").
				SetPartialFilterExpression(bson.M{"is_deleted": false}),
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, models)
	return err
}

// ============================================================================
// Database Models & Mappers
// ============================================================================

type categoryDoc struct {
	ID bson.ObjectID `bson:"_id,omitempty"`
	Name        string `bson:"name"`
	Slug        string `bson:"slug"`
	Description string `bson:"description,omitempty"`
	ParentID *bson.ObjectID `bson:"parent_id,omitempty"`
	Thumbnail string `bson:"thumbnail,omitempty"`
	Banner    string `bson:"banner,omitempty"`
	MetaTitle       string `bson:"meta_title,omitempty"`
	MetaDescription string `bson:"meta_description,omitempty"`
	Position   int64 `bson:"position"`
	IsFeatured bool  `bson:"is_featured"`
	Status    string `bson:"status"`
	IsDeleted bool   `bson:"is_deleted"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at"`
}

func toCategoryDoc(c entities.Category) categoryDoc {
	doc := categoryDoc{
		Name: c.Name,
		Slug: c.Slug,
		Description: c.Description,
		Thumbnail: c.Thumbnail,
		Banner: c.Banner,
		MetaTitle: c.MetaTitle,
		MetaDescription: c.MetaDescription,
		Position: c.Position,
		IsFeatured: c.IsFeatured,
		Status: c.Status,
		IsDeleted: c.IsDeleted,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		DeletedAt: c.DeletedAt,
	}
	if c.ID != "" {
		if id, err := bson.ObjectIDFromHex(c.ID); err == nil {
			doc.ID = id
		}
	}
	if c.ParentID != nil && *c.ParentID != "" {
		if id, err := bson.ObjectIDFromHex(*c.ParentID); err == nil {
			doc.ParentID = &id
		}
	}
	return doc
}

func toCategoryEntity(doc categoryDoc) entities.Category {
	c := entities.Category{
		ID: doc.ID.Hex(),
		Name: doc.Name,
		Slug: doc.Slug,
		Description: doc.Description,
		Thumbnail: doc.Thumbnail,
		Banner: doc.Banner,
		MetaTitle: doc.MetaTitle,
		MetaDescription: doc.MetaDescription,
		Position: doc.Position,
		IsFeatured: doc.IsFeatured,
		Status: doc.Status,
		IsDeleted: doc.IsDeleted,
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
		DeletedAt: doc.DeletedAt,
	}
	if doc.ParentID != nil && !doc.ParentID.IsZero() {
		pid := doc.ParentID.Hex()
		c.ParentID = &pid
	}
	return c
}

// ============================================================================
// Existence Checks
// ============================================================================

func (r *categoryRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	objID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	err = r.collection.FindOne(ctx, bson.M{
		"_id":        objID,
		"is_deleted": false,
	}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *categoryRepository) ExistsByName(ctx context.Context, name string, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "name", name, excludeID)
}

func (r *categoryRepository) ExistsBySlug(ctx context.Context, slug string, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "slug", slug, excludeID)
}

func (r *categoryRepository) ExistsByPosition(ctx context.Context, position int64, excludeID *string) (bool, error) {
	return r.existsByField(ctx, "position", position, excludeID)
}

func (r *categoryRepository) existsByField(ctx context.Context, field string, value any, excludeID *string) (bool, error) {
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

