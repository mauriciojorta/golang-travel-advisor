package models

import (
	"errors"
	"testing"

	"example.com/travel-advisor/db"
	"example.com/travel-advisor/utils"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUser_Create(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectPrepare("INSERT INTO users").
		ExpectExec().
		WithArgs("test@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	user := NewUser("test@example.com", "password123")

	err = user.Create()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_Create_Error(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectPrepare("INSERT INTO users").
		ExpectExec().
		WithArgs("test@example.com", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert error"))

	user := NewUser("test@example.com", "password123")

	err = user.Create()
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_FindUser_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectQuery("SELECT id FROM users WHERE email=?").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	user := InitUser()

	u, err := user.FindByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_FindUser_NotFound(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectQuery("SELECT id FROM users WHERE email=?").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // No rows returned

	user := InitUser()

	_, err = user.FindByEmail("test@example.com")
	assert.Error(t, err)
	assert.Equal(t, "sql: no rows in result set", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_FindUser_DBError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectQuery("SELECT id FROM users WHERE email=?").
		WithArgs("test@example.com").
		WillReturnError(errors.New("database error"))

	user := InitUser()

	_, err = user.FindByEmail("test@example.com")
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_ValidateCredentials(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	password := "password123"
	hashedPassword, err := utils.HashPassword(password)

	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	mock.ExpectQuery("SELECT id,password FROM users WHERE email=?").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(1, hashedPassword))

	user := InitUser()
	user.Email = "test@example.com"

	err = user.ValidateCredentials("password123")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUser_ValidateCredentials_Invalid(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectQuery("SELECT id,password FROM users WHERE email=?").
		WithArgs("test@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "password"}).AddRow(1, "hashedPassword"))

	user := InitUser()
	user.Email = "test@example.com"

	err = user.ValidateCredentials("password123")
	assert.Error(t, err)
	assert.Equal(t, "credentials invalid", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
