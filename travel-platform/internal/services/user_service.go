package services

import (
	"fmt"
	"travel-platform/travel-platform/internal/models"
	"travel-platform/travel-platform/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(email, password, firstName, lastName string) (*models.User, error)
	Login(email, password string) (*models.User, error)
	GetProfile(userID uint) (*models.User, error)
	UpdateProfile(userID uint, firstName, lastName string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(email, password, firstName, lastName string) (*models.User, error) {

	existingUser, _ := s.repo.GetUserByEmail(email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// ise userService nesnesinin metodudur ve bu nesneye bağlı işlemleri gerçekleştirebilir.
func (s *userService) Login(email, password string) (*models.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	return user, nil
}

func (s *userService) GetProfile(userID uint) (*models.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *userService) UpdateProfile(userID uint, firstName, lastName string) (*models.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	err = s.repo.UpdateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}
