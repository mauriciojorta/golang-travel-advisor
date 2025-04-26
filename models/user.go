package models

import (
	"errors"
	"time"

	"example.com/travel-advisor/db"
	"example.com/travel-advisor/utils"
)

type User struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email" binding:"required"`
	Password      string    `json:"password"`
	CreationDate  time.Time `json:"creationDate"`
	UpdateDate    time.Time `json:"updateDate"`
	LastLoginDate time.Time `json:"lastLoginDate"`

	FindByEmail         func() error
	Create              func() error
	ValidateCredentials func() error
}

var NewUser = func(email, password string) *User {
	user := &User{
		Email:    email,
		Password: password,
	}

	// Set default implementations for FindUser and Create
	user.FindByEmail = user.defaultFindUser
	user.Create = user.defaultCreate
	user.ValidateCredentials = user.defaultValidateCredentials

	return user
}

func (u *User) defaultCreate() error {

	query := `INSERT INTO users(email, password, creation_date, update_date) 
	VALUES (?, ?, ?, ?)`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)

	if err != nil {
		return err
	}

	result, err := stmt.Exec(u.Email, hashedPassword, time.Now(), time.Now())
	if err != nil {
		return err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	u.ID = int64(userId)
	return err
}

func (u *User) defaultFindUser() error {
	query := "SELECT id FROM users WHERE email=?"
	row := db.DB.QueryRow(query, u.Email)

	err := row.Scan(&u.ID)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) defaultValidateCredentials() error {
	query := "SELECT id,password FROM users WHERE email=?"
	row := db.DB.QueryRow(query, u.Email)

	var retrievedPassword string
	err := row.Scan(&u.ID, &retrievedPassword)
	if err != nil {
		return err
	}

	isValid := utils.CheckPasswordHash(retrievedPassword, u.Password)
	if !isValid {
		return errors.New("Credentials invalid.")
	}

	return nil
}
