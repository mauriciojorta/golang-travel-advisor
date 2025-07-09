package utils

import (
	"errors"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

var (
	secretKey     string
	secretKeyOnce sync.Once
)

// Initialize the secret key once
func InitJwtSecretKey() error {
	var err error

	secretKeyOnce.Do(func() {
		secretKey := os.Getenv("JWT_SECRET_KEY")
		if secretKey == "" {
			log.Error("JWT_SECRET_KEY environment variable is not set")
			err = errors.New("JWT_SECRET_KEY environment variable is not set")
		}
	})

	return err
}

var GenerateToken = func(email string, userId int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"userId": userId,
		"exp":    time.Now().Add(time.Hour * 2).Unix(),
	})

	return token.SignedString([]byte(secretKey))
}

var VerifyToken = func(token string) (int64, error) {
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			log.Error("Unexpected signing method: ", token.Header["alg"])
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		log.Error("Error parsing token: ", err)
		return 0, errors.New("could not parse token")
	}

	tokenIsValid := parsedToken.Valid

	if !tokenIsValid {
		log.Error("Invalid token!")
		return 0, errors.New("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)

	if !ok {
		log.Error("Invalid token claims!")
		return 0, errors.New("invalid token claims")
	}

	// email := claims["email"].(string)
	userId := int64(claims["userId"].(float64))

	return userId, nil
}
