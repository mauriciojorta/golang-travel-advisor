package models

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"
)

type AuditEvent struct {
	ID               int64
	UserID           int64
	EventDescription string
	EventDate        *time.Time

	CreateAuditEvent func(*sql.Tx) error
}

var InitAuditEventFunctions = func(auditEvent *AuditEvent) *AuditEvent {
	// Set default SQL implementation for CreateAuditEvent. In the future there could be implementations for
	// other NoSQL DB systems like MongoDB
	auditEvent.CreateAuditEvent = auditEvent.defaultCreateAuditEvent

	return auditEvent
}

var NewAuditEvent = func(userId int64, eventDescription string) *AuditEvent {
	auditEvent := &AuditEvent{
		UserID:           userId,
		EventDescription: eventDescription,
	}

	return InitAuditEventFunctions(auditEvent)

}

func (ae *AuditEvent) defaultCreateAuditEvent(tx *sql.Tx) error {
	query := `INSERT INTO audit_events(user_id, event_description, event_date) 
	VALUES (?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing statement for user creation: %v", err)
		return err
	}
	defer stmt.Close()

	now := time.Now()
	ae.EventDate = &now

	result, err := stmt.Exec(ae.UserID, ae.EventDescription, ae.EventDate)
	if err != nil {
		log.Errorf("Error executing statement for user creation: %v", err)
		return err
	}

	auditId, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID for user creation: %v", err)
		return err
	}

	ae.ID = int64(auditId)
	return err
}
