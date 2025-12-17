package repository

import (
	"travel-platform/travel-platform/internal/models"

	"gorm.io/gorm"
)

// Test yazmak kolaylaşır (mock repository)

// Veritabanı değişirse (Postgres → Mongo gibi) kodun geri kalanı etkilenmez

// Clean Architecture / SOLID prensiplerine uygundur
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

// Constructor function
// Dışarıdan bir *gorm.DB alır userRepository oluşturur
// Kodun geri kalanı concrete struct’ı değil interface’i kullanır
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := r.db.First(&user, id).Error
	if result != nil {
		return nil, result
	}
	return &user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := r.db.Where("email = ?", email).First(&user).Error
	if result != nil {
		return nil, result
	}
	return &user, nil
}
func (r *userRepository) GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := r.db.Find(&users).Error
	if result != nil {
		return nil, result
	}
	return users, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}
