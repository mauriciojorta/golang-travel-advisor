package utils

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Errorf("Error hashing password: %v", err)
		return "", err
	}
	return string(hashedPasswordBytes), nil
}

func CheckPasswordHash(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Errorf("Error checking password hash: %v", err)
		return false
	}
	return true
}
