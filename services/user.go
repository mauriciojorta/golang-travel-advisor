package services

import (
	"errors"

	"example.com/travel-advisor/db"
	"example.com/travel-advisor/models" // Replace with the actual path to the User struct
	"example.com/travel-advisor/utils"
	log "github.com/sirupsen/logrus"
)

type UserServiceInterface interface {
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	ValidateCredentials(user *models.User, password string) error
	GenerateLoginToken(user *models.User) (string, error)
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
		log.Error("Email cannot be empty")
		return nil, errors.New("email cannot be empty")
	}
	user := models.InitUser()
	return user.FindByEmail(email)
}

// Create creates a new user
func (us *UserService) Create(user *models.User) error {
	if user == nil {
		log.Error("User instance is nil")
		return errors.New("user instance is nil")
	}
	return user.Create()
}

// ValidateCredentials validates the user's credentials
func (us *UserService) ValidateCredentials(user *models.User, password string) error {
	if user == nil {
		log.Error("User instance is nil")
		return errors.New("user instance is nil")
	}
	if password == "" {
		log.Error("Password cannot be empty")
		return errors.New("password cannot be empty")
	}
	return user.ValidateCredentials(password)
}

func (us *UserService) GenerateLoginToken(user *models.User) (string, error) {
	token, err := utils.GenerateToken(user.Email, user.ID)
	if err != nil {
		log.Errorf("Error generating token: %v", err)
		return "", errors.New("error generating token")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction for login auditing: %v", err)
		return "", err
	}

	defer db.HandleTransaction(tx, &err)

	err = user.UpdateLastLoginDate(tx)
	if err != nil {
		log.Errorf("Error updating last login date: %v", err)
		return "", errors.New("error updating last login date")
	}

	auditEvent := models.NewAuditEvent(user.ID, "User login")
	err = auditEvent.CreateAuditEvent(tx)
	if err != nil {
		log.Errorf("Error saving login event: %v", err)
		return "", errors.New("error saving login event")
	}

	return token, nil
}
