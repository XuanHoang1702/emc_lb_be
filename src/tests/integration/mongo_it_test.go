//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"emc_lb/src/internal/repository"
	"emc_lb/src/pkg/entities"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func couponEntity(code string, now time.Time) entities.Coupon {
	return entities.Coupon{
		Code:           code,
		Type:           "percentage",
		Value:          10,
		MinOrderAmount: 100,
		UsageLimit:     100,
		UsageCount:     0,
		StartDate:      now.Add(-time.Hour),
		EndDate:        now.Add(24 * time.Hour),
		IsActive:       true,
		IsDeleted:      false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// runMongo starts a real MongoDB container and returns a connected client.
func runMongo(t *testing.T, ctx context.Context) *mongo.Client {
	t.Helper()

	container, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		t.Fatalf("start mongo: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	uri, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("mongo connection string: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping mongo: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return client
}

func TestMongo_CouponRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	client := runMongo(t, ctx)

	coll := client.Database("emc_lb").Collection("coupons")
	repo := repository.NewCouponRepository(coll)

	now := time.Now().UTC()
	created, err := repo.Create(ctx, couponEntity("SAVE10", now))
	if err != nil {
		t.Fatalf("Create coupon: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected created coupon to have an ID")
	}

	got, err := repo.GetByCode(ctx, "SAVE10")
	if err != nil {
		t.Fatalf("GetByCode: %v", err)
	}
	if got.Code != "SAVE10" {
		t.Fatalf("expected code SAVE10, got %s", got.Code)
	}
	if got.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, got.ID)
	}

	if err := repo.IncrementUsage(ctx, "SAVE10", 1); err != nil {
		t.Fatalf("IncrementUsage: %v", err)
	}
	afterInc, _ := repo.GetByCode(ctx, "SAVE10")
	if afterInc.UsageCount != 1 {
		t.Fatalf("expected usage_count 1, got %d", afterInc.UsageCount)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 coupon, got %d", len(list))
	}

	updated, err := repo.Update(ctx, created.ID, map[string]any{"is_active": false})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.IsActive {
		t.Fatal("expected is_active false after update")
	}
}

func TestMongo_CouponRepository_SoftDeleteExcluded(t *testing.T) {
	ctx := context.Background()
	client := runMongo(t, ctx)

	coll := client.Database("emc_lb").Collection("coupons")
	repo := repository.NewCouponRepository(coll)

	created, err := repo.Create(ctx, couponEntity("GONE", time.Now().UTC()))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Simulate a soft delete then confirm GetByCode no longer returns it.
	objID, err := bson.ObjectIDFromHex(created.ID)
	if err != nil {
		t.Fatalf("parse created id: %v", err)
	}
	del := bson.M{"$set": bson.M{"is_deleted": true}}
	if _, err := coll.UpdateOne(ctx, bson.M{"_id": objID}, del); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	_, err = repo.GetByCode(ctx, "GONE")
	if err == nil {
		t.Fatal("expected GetByCode to fail for soft-deleted coupon")
	}
	if err != mongo.ErrNoDocuments {
		t.Fatalf("expected mongo.ErrNoDocuments, got %v", err)
	}
}
