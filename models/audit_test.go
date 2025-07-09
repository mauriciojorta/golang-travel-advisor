package models

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDefaultCreateAuditEvent_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				t.Fatalf("Failed to commit transaction: %v", err)
			}
		}
	}()

	auditEvent := NewAuditEvent(1, "Created something")

	mock.ExpectPrepare("INSERT INTO audit_events").
		ExpectExec().
		WithArgs(auditEvent.UserID, auditEvent.EventDescription, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(42, 1))

	err = auditEvent.defaultCreateAuditEvent(tx)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), auditEvent.ID)
	assert.NotNil(t, auditEvent.EventDate)
	assert.NoError(t, mock.ExpectationsWereMet())
	mock.ExpectCommit()
}

func TestDefaultCreateAuditEvent_PrepareError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				t.Fatalf("Failed to commit transaction: %v", err)
			}
		}
	}()

	auditEvent := NewAuditEvent(1, "Prepare error")

	mock.ExpectPrepare("INSERT INTO audit_events").
		WillReturnError(errors.New("prepare failed"))

	err = auditEvent.defaultCreateAuditEvent(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prepare failed")
	mock.ExpectRollback()
}

func TestDefaultCreateAuditEvent_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				t.Fatalf("Failed to commit transaction: %v", err)
			}
		}
	}()

	auditEvent := NewAuditEvent(1, "Exec error")

	mock.ExpectPrepare("INSERT INTO audit_events").
		ExpectExec().
		WithArgs(auditEvent.UserID, auditEvent.EventDescription, sqlmock.AnyArg()).
		WillReturnError(errors.New("exec failed"))

	err = auditEvent.defaultCreateAuditEvent(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exec failed")
	mock.ExpectRollback()

}

func TestDefaultCreateAuditEvent_LastInsertIdError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			err = tx.Rollback()
			if err != nil {
				t.Fatalf("Failed to rollback transaction: %v", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				t.Fatalf("Failed to commit transaction: %v", err)
			}
		}
	}()

	auditEvent := NewAuditEvent(1, "LastInsertId error")

	mock.ExpectPrepare("INSERT INTO audit_events").
		ExpectExec().
		WithArgs(auditEvent.UserID, auditEvent.EventDescription, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("last insert id failed")))

	err = auditEvent.defaultCreateAuditEvent(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "last insert id failed")
	mock.ExpectRollback()

}
