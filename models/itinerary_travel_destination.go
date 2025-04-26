package models

import (
	"database/sql"
	"time"

	"example.com/travel-advisor/db"
)

type ItineraryTravelDestination struct {
	ID            int64     `json:"id"`
	Country       string    `json:"country" binding:"required"`
	City          string    `json:"city" binding:"required"`
	ItineraryID   int64     `json:"itineraryId"`
	ArrivalDate   time.Time `json:"arrivalDate" binding:"required"`
	DepartureDate time.Time `json:"departureDate" binding:"required"`
	CreationDate  time.Time `json:"creationDate"`
	UpdateDate    time.Time `json:"updateDate"`

	Find   func() error
	Create func(*sql.Tx) error
	Update func() error
	Delete func() error
}

var NewItineraryTravelDestination = func(country string, city string, itineraryId int64, arrivalDate time.Time, departureDate time.Time) *ItineraryTravelDestination {
	destination := &ItineraryTravelDestination{
		Country:       country,
		City:          city,
		ItineraryID:   itineraryId,
		ArrivalDate:   arrivalDate,
		DepartureDate: departureDate,
	}

	// Set default implementations for Find, Create, Update, and Delete
	destination.Find = destination.defaultFind
	destination.Create = destination.defaultCreate
	destination.Update = destination.defaultUpdate
	destination.Delete = destination.defaultDelete

	return destination
}

func (d *ItineraryTravelDestination) defaultFind() error {

	query := `SELECT id, country, city, itinerary_id, arrival_date, departure_date
	FROM itinerary_travel_destinations WHERE itinerary_id = ?`
	row := db.DB.QueryRow(query, d.ItineraryID)

	err := row.Scan(&d.ID, &d.Country, &d.City, &d.ItineraryID, &d.ArrivalDate, &d.DepartureDate)
	if err != nil {
		return err
	}

	return nil
}

func (d *ItineraryTravelDestination) defaultCreate(tx *sql.Tx) error {
	query := `INSERT INTO itinerary_travel_destinations (country, city, itinerary_id, arrival_date, departure_date)
	VALUES (?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(d.Country, d.City, d.ItineraryID, d.ArrivalDate, d.DepartureDate)
	if err != nil {
		return err
	}

	travelDestinationId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	d.ID = travelDestinationId

	return nil
}

func (d *ItineraryTravelDestination) defaultUpdate() error {
	query := `UPDATE itinerary_travel_destinations SET country = ?, city = ?, arrival_date = ?, departure_date = ? WHERE id = ?`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.Country, d.City, d.ArrivalDate, d.DepartureDate, d.ID)
	if err != nil {
		return err
	}

	return nil
}

func (d *ItineraryTravelDestination) defaultDelete() error {
	query := `DELETE FROM itinerary_travel_destinations WHERE id = ?`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.ID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteByItineraryIdTx(tx *sql.Tx, itineraryId int64) error {
	query := `DELETE FROM itinerary_travel_destinations WHERE itinerary_id = ?`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(itineraryId)
	if err != nil {
		return err
	}

	return nil
}
