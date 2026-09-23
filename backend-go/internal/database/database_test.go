package database_test

import (
	"testing"

	"backend-go/internal/config"
	"backend-go/internal/database"
)

func TestAutoMigrate_AllEntities(t *testing.T) {
	cfg := &config.Config{
		DBDriver: "sqlite",
		DBDSN:    ":memory:",
		Env:      "development",
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test sqlite database: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate failed for all entities: %v", err)
	}
}
