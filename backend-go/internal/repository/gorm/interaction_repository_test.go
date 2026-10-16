package gorm

import (
	"context"
	"testing"
	"time"

	"backend-go/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	gormPkg "gorm.io/gorm"
)

func TestFavoriteRepository_ListByUser(t *testing.T) {
	db, err := gormPkg.Open(sqlite.Open("file::memory:?cache=shared"), &gormPkg.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&entity.User{}, &entity.Profile{}, &entity.Favorite{})
	assert.NoError(t, err)

	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	u1 := entity.User{ID: uuid.New(), Email: "u1@test.com"}
	u2 := entity.User{ID: uuid.New(), Email: "u2@test.com"}
	p2 := entity.Profile{ID: uuid.New(), UserID: u2.ID, FirstName: "Fav", LastName: "User"}
	fav := entity.Favorite{ID: uuid.New(), UserID: u1.ID, FavoriteUserID: u2.ID, CreatedAt: time.Now()}

	assert.NoError(t, db.Create(&u1).Error)
	assert.NoError(t, db.Create(&u2).Error)
	assert.NoError(t, db.Create(&p2).Error)
	assert.NoError(t, db.Create(&fav).Error)

	list, err := repo.ListByUser(ctx, u1.ID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.NotNil(t, list[0].FavoriteUser)
	assert.NotNil(t, list[0].FavoriteUser.Profile)
	assert.Equal(t, "Fav", list[0].FavoriteUser.Profile.FirstName)
}

func TestPostgres_RealQueries(t *testing.T) {
	dsn := "postgres://postgres:gentechco@localhost:5432/ns_db?sslmode=disable"
	db, err := gormPkg.Open(gormPostgres.Open(dsn), &gormPkg.Config{})
	if err != nil {
		t.Skip("Postgres not available")
	}

	catRepo := NewCatalogServiceRepository(db)
	services, err := catRepo.List(context.Background(), "", "")
	assert.NoError(t, err)
	t.Logf("Found %d catalog services", len(services))

	favRepo := NewFavoriteRepository(db)
	favs, err := favRepo.ListByUser(context.Background(), uuid.New())
	assert.NoError(t, err)
	t.Logf("Found %d favorites", len(favs))
}
