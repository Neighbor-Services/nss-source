package main

import (
	"testing"
	"time"

	"backend-go/internal/config"
	"backend-go/internal/database"
	"backend-go/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSeedServicesDataIntegrity(t *testing.T) {
	// Verify that we have all 18 parent categories defined
	assert.Equal(t, 18, len(categoriesSeed), "Expected exactly 18 official parent categories")

	totalServices := 0
	categoryNames := make(map[string]bool)

	for _, cat := range categoriesSeed {
		assert.NotEmpty(t, cat.Name)
		assert.NotEmpty(t, cat.Description)
		assert.NotEmpty(t, cat.Image)
		assert.True(t, len(cat.Services) > 0, "Category %s must have child services", cat.Name)
		assert.False(t, categoryNames[cat.Name], "Duplicate category name: %s", cat.Name)
		categoryNames[cat.Name] = true

		for _, srv := range cat.Services {
			assert.NotEmpty(t, srv.Name)
			assert.NotEmpty(t, srv.Description)
			assert.True(t, srv.BasePrice > 0, "Base price must be positive for %s", srv.Name)
			assert.NotEmpty(t, srv.ServiceLocation)
			assert.True(t, len(srv.Specialties) > 0, "Specialties must be present for %s", srv.Name)
			totalServices++
		}
	}

	assert.Equal(t, 142, totalServices, "Expected 142 child catalog services across 18 categories")
}

func TestSeedServicesExecutionOnDB(t *testing.T) {
	cfg := &config.Config{
		DBDriver: "sqlite",
		DBDSN:    ":memory:",
		Env:      "testing",
	}

	db, err := database.Connect(cfg)
	assert.NoError(t, err)
	assert.NoError(t, database.AutoMigrate(db))

	// Clear previous
	db.Exec("DELETE FROM accounts_profile_catalog_services")
	db.Exec("UPDATE services_servicerequest SET catalog_service_id = NULL")
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.CatalogService{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&entity.Category{})

	// Insert seeds
	for _, catData := range categoriesSeed {
		categoryID := uuid.New()
		cat := entity.Category{
			ID:          categoryID,
			Name:        catData.Name,
			Description: catData.Description,
			Image:       catData.Image,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		assert.NoError(t, db.Create(&cat).Error)

		for _, srvData := range catData.Services {
			price := srvData.BasePrice
			catalogService := entity.CatalogService{
				ID:                     uuid.New(),
				CategoryID:             categoryID,
				Name:                   srvData.Name,
				Description:            srvData.Description,
				BasePrice:              &price,
				Specialties:            entity.JSONSlice(srvData.Specialties),
				DefaultServiceLocation: srvData.ServiceLocation,
				CreatedAt:              time.Now().UTC(),
				UpdatedAt:              time.Now().UTC(),
			}
			assert.NoError(t, db.Create(&catalogService).Error)
		}
	}

	// Verify database contents
	var catCount int64
	db.Model(&entity.Category{}).Count(&catCount)
	assert.Equal(t, int64(18), catCount)

	var srvCount int64
	db.Model(&entity.CatalogService{}).Count(&srvCount)
	assert.Equal(t, int64(142), srvCount)

	// Verify querying specific service with specialties
	var bakeryService entity.CatalogService
	assert.NoError(t, db.Preload("Category").Where("name = ?", "Bakery Services").First(&bakeryService).Error)
	assert.Equal(t, "Food and Catering", bakeryService.Category.Name)
	assert.Contains(t, []string(bakeryService.Specialties), "Donuts")
	assert.Contains(t, []string(bakeryService.Specialties), "Wedding cakes")

	var hairService entity.CatalogService
	assert.NoError(t, db.Preload("Category").Where("name = ?", "Hair Services").First(&hairService).Error)
	assert.Equal(t, "Beauty and Personal Style", hairService.Category.Name)
	assert.Contains(t, []string(hairService.Specialties), "Braiding")
}
