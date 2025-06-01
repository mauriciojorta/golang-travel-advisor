package services

import (
	"errors"

	"example.com/travel-advisor/models" // Replace with the actual path to the User struct
)

type UserServiceInterface interface {
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	ValidateCredentials(user *models.User) error
}

type UserService struct{}

// singleton instance
var userServiceInstance = &UserService{}

// GetUserService returns the singleton instance of UserService
var GetUserService = func() UserServiceInterface {
	return userServiceInstance
}

// FindByEmail retrieves the user by their email
func (us *UserService) FindByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	user := models.NewUser(email, "") // Create a new User instance with the email
	err := user.FindByEmail()
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Create creates a new user
func (us *UserService) Create(user *models.User) error {
	if user == nil {
		return errors.New("user instance is nil")
	}
	return user.Create()
}

// ValidateCredentials validates the user's credentials
func (us *UserService) ValidateCredentials(user *models.User) error {
	if user == nil {
		return errors.New("user instance is nil")
	}
	return user.ValidateCredentials()
}
