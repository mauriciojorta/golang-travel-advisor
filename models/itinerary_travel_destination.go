package models

import (
	"database/sql"
	"time"

	"example.com/travel-advisor/db"
)

type ItineraryTravelDestination struct {
	ID            int64      `json:"id"`
	Country       string     `json:"country" binding:"required"`
	City          string     `json:"city" binding:"required"`
	ItineraryID   int64      `json:"itineraryId"`
	ArrivalDate   time.Time  `json:"arrivalDate" binding:"required"`
	DepartureDate time.Time  `json:"departureDate" binding:"required"`
	CreationDate  *time.Time `json:"creationDate"`
	UpdateDate    *time.Time `json:"updateDate"`

	FindByItineraryId     func() (*[]ItineraryTravelDestination, error) `json:"-"`
	Create                func(*sql.Tx) error                           `json:"-"`
	Update                func() error                                  `json:"-"`
	Delete                func() error                                  `json:"-"`
	DeleteByItineraryIdTx func(*sql.Tx) error                           `json:"-"`
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
	destination.FindByItineraryId = destination.defaultFindByItineraryId
	destination.Create = destination.defaultCreate
	destination.Update = destination.defaultUpdate
	destination.Delete = destination.defaultDelete
	destination.DeleteByItineraryIdTx = destination.defaultDeleteByItineraryIdTx

	return destination
}

func (d *ItineraryTravelDestination) defaultFindByItineraryId() (*[]ItineraryTravelDestination, error) {

	query := `SELECT id, country, city, itinerary_id, arrival_date, departure_date
	FROM itinerary_travel_destinations WHERE itinerary_id = ? ORDER BY arrival_date ASC`
	destRows, err := db.DB.Query(query, d.ItineraryID)
	if err != nil {
		return nil, err
	}

	var travelDestinations []ItineraryTravelDestination

	for destRows.Next() {
		var destination ItineraryTravelDestination
		err := destRows.Scan(&destination.ID, &destination.Country, &destination.City, &destination.ItineraryID, &destination.ArrivalDate, &destination.DepartureDate)
		if err != nil {
			destRows.Close()
			return nil, err
		}
		travelDestinations = append(travelDestinations, destination)
	}
	destRows.Close()

	return &travelDestinations, nil
}

func (d *ItineraryTravelDestination) defaultCreate(tx *sql.Tx) error {
	query := `INSERT INTO itinerary_travel_destinations (country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(d.Country, d.City, d.ItineraryID, d.ArrivalDate, d.DepartureDate, time.Now(), time.Now())
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
	query := `UPDATE itinerary_travel_destinations SET country = ?, city = ?, arrival_date = ?, departure_date = ?, update_date = ? WHERE id = ?`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.Country, d.City, d.ArrivalDate, d.DepartureDate, time.Now(), d.ID)
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

func (d *ItineraryTravelDestination) defaultDeleteByItineraryIdTx(tx *sql.Tx) error {
	query := `DELETE FROM itinerary_travel_destinations WHERE itinerary_id = ?`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.ItineraryID)
	if err != nil {
		return err
	}

	return nil
}
