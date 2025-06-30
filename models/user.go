package models

import (
	"errors"
	"time"

	"example.com/travel-advisor/db"
	"example.com/travel-advisor/utils"
	log "github.com/sirupsen/logrus"
)

type User struct {
	ID            int64      `json:"id"`
	Email         string     `json:"email" binding:"required"`
	Password      string     `json:"password"`
	CreationDate  *time.Time `json:"creationDate"`
	UpdateDate    *time.Time `json:"updateDate"`
	LastLoginDate *time.Time `json:"lastLoginDate"`

	FindByEmail         func(email string) (*User, error) `json:"-"`
	Create              func() error                      `json:"-"`
	ValidateCredentials func(password string) error       `json:"-"`
}

var InitUser = func() *User {
	return InitUserFunctions(&User{})
}

var InitUserFunctions = func(user *User) *User {
	// Set default implementations for FindByEmail, Create, and ValidateCredentials
	user.FindByEmail = user.defaultFindUser
	user.Create = user.defaultCreate
	user.ValidateCredentials = user.defaultValidateCredentials

	return user
}

var NewUser = func(email, password string) *User {
	user := &User{
		Email:    email,
		Password: password,
	}

	return InitUserFunctions(user)
}

func (u *User) defaultCreate() error {

	query := `INSERT INTO users(email, password, creation_date, update_date) 
	VALUES (?, ?, ?, ?)`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing statement for user creation: %v", err)
		return err
	}
	defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)

	if err != nil {
		log.Errorf("Error hashing password for user creation: %v", err)
		return err
	}

	result, err := stmt.Exec(u.Email, hashedPassword, time.Now(), time.Now())
	if err != nil {
		log.Errorf("Error executing statement for user creation: %v", err)
		return err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID for user creation: %v", err)
		return err
	}

	u.ID = int64(userId)
	return err
}

func (u *User) defaultFindUser(email string) (*User, error) {
	user := &User{
		Email: email,
	}

	query := "SELECT id FROM users WHERE email=?"
	row := db.DB.QueryRow(query, email)

	err := row.Scan(&user.ID)
	if err != nil {
		log.Errorf("Error finding user by email: %v", err)
		return nil, err
	}

	return user, nil
}

func (u *User) defaultValidateCredentials(password string) error {

	query := "SELECT id,password FROM users WHERE email=?"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword)
	if err != nil {
		log.Errorf("Error retrieving user credentials: %v", err)
		return err
	}

	isValid := utils.CheckPasswordHash(retrievedPassword, password)
	if !isValid {
		log.Errorf("Invalid credentials for user with email: %s", u.Email)
		return errors.New("credentials invalid.")
	}

	return nil
}
