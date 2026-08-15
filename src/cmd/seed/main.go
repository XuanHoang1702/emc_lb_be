package main

import (
	"context"
	"fmt"
	"log"

	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found or error loading, relying on system env variables")
	}

	dsn := utils.GetEnv("POSTGRES_DSN", "")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is not set")
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping postgres: %v", err)
	}

	queries := sqlc.New(db)

	// Helper for *string
	strPtr := func(s string) *string { return &s }

	// Seed Roles
	log.Println("Seeding roles...")
	roles := []string{"admin", "user", "manager"}
	roleMap := make(map[string]sqlc.Role)

	for _, name := range roles {
		role, err := queries.CreateRole(ctx, sqlc.CreateRoleParams{
			Code:        name,
			Name:        name,
			Description: strPtr(fmt.Sprintf("%s role", name)),
		})
		if err != nil {
			log.Printf("Role '%s' might already exist or err: %v", name, err)
		} else {
			roleMap[name] = role
			log.Printf("Created role: %s", name)
		}
	}

	// Seed Permissions
	log.Println("Seeding permissions...")
	permissions := []struct {
		Resource string
		Action   string
	}{
		// Product
		{"product", "create"},
		{"product", "update"},
		{"product", "delete"},

		// Category
		{"category", "create"},
		{"category", "update"},
		{"category", "delete"},

		// Brand
		{"brand", "create"},
		{"brand", "update"},
		{"brand", "delete"},

		// Order
		{"manage_orders", "manage"},
	}

	permMap := make(map[string]sqlc.Permission)
	for _, p := range permissions {
		permName := fmt.Sprintf("%s:%s", p.Resource, p.Action)
		if p.Resource == "manage_orders" {
			permName = "manage_orders"
		}

		perm, err := queries.CreatePermission(ctx, sqlc.CreatePermissionParams{
			Code:        permName,
			Name:        permName,
			Description: strPtr(fmt.Sprintf("Can %s %s", p.Action, p.Resource)),
		})
		if err != nil {
			log.Printf("Permission '%s' might already exist or err: %v", permName, err)
		} else {
			permMap[permName] = perm
			log.Printf("Created permission: %s", permName)

			// Auto assign to admin
			_ = queries.AssignPermissionToRole(ctx, sqlc.AssignPermissionToRoleParams{
				RoleCode:       "admin",
				PermissionCode: permName,
			})
		}
	}

	log.Println("Seeding finished. If you saw 'might already exist' errors, this is normal for a re-run on an existing database.")
}
