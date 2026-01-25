package tests

import (
	"testing"
	"travel-platform/internal/models"
	"travel-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	user := &models.User{
		Email:     "repo@test.com",
		Password:  "hashedpassword",
		FirstName: "Repo",
		LastName:  "Test",
	}

	err := repo.CreateUser(user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	var savedUser models.User
	db.First(&savedUser, user.ID)
	assert.Equal(t, "repo@test.com", savedUser.Email)
}

func TestGetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	email := "findme@test.com"
	db.Create(&models.User{Email: email, FirstName: "Find"})

	user, err := repo.GetUserByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, email, user.Email)
}
