package models

import (
	"database/sql"
	"time"

	log "github.com/sirupsen/logrus"

	"example.com/travel-advisor/db"
)

type ItineraryTravelDestination struct {
	ID            int64      `json:"id" example:"1"`
	Country       string     `json:"country" binding:"required" example:"Spain"`
	City          string     `json:"city" binding:"required" example:"Madrid"`
	ItineraryID   int64      `json:"itineraryId" example:"1"`
	ArrivalDate   time.Time  `json:"arrivalDate" binding:"required" example:"2024-07-01T00:00:00Z"`
	DepartureDate time.Time  `json:"departureDate" binding:"required" example:"2024-07-05T00:00:00Z"`
	CreationDate  *time.Time `json:"creationDate,omitempty" example:"2024-06-01T00:00:00Z"`
	UpdateDate    *time.Time `json:"updateDate,omitempty" example:"2024-06-01T00:00:00Z"`

	FindByItineraryId     func(itineraryId int64) ([]*ItineraryTravelDestination, error) `json:"-"`
	Create                func(*sql.Tx) error                                            `json:"-"`
	Update                func() error                                                   `json:"-"`
	Delete                func() error                                                   `json:"-"`
	DeleteByItineraryIdTx func(itineraryId int64, tx *sql.Tx) error                      `json:"-"`
}

var InitItineraryTravelDestination = func() *ItineraryTravelDestination {
	return initItineraryTravelDestinationFunctions(&ItineraryTravelDestination{})
}

var initItineraryTravelDestinationFunctions = func(destination *ItineraryTravelDestination) *ItineraryTravelDestination {
	// Set default implementations for Find, Create, Update, and Delete
	destination.FindByItineraryId = destination.defaultFindByItineraryId
	destination.Create = destination.defaultCreate
	destination.Update = destination.defaultUpdate
	destination.Delete = destination.defaultDelete
	destination.DeleteByItineraryIdTx = destination.defaultDeleteByItineraryIdTx

	return destination
}

var NewItineraryTravelDestination = func(country string, city string, arrivalDate time.Time, departureDate time.Time) *ItineraryTravelDestination {
	destination := &ItineraryTravelDestination{
		Country:       country,
		City:          city,
		ArrivalDate:   arrivalDate,
		DepartureDate: departureDate,
	}

	return initItineraryTravelDestinationFunctions(destination)
}

func (d *ItineraryTravelDestination) defaultFindByItineraryId(itineraryId int64) ([]*ItineraryTravelDestination, error) {

	query := `SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date
	FROM itinerary_travel_destinations WHERE itinerary_id = ? ORDER BY arrival_date ASC`
	destRows, err := db.DB.Query(query, itineraryId)
	if err != nil {
		log.Errorf("Error querying itinerary travel destinations: %v", err)
		return nil, err
	}

	var travelDestinations []*ItineraryTravelDestination

	for destRows.Next() {
		var destination ItineraryTravelDestination
		err := destRows.Scan(&destination.ID, &destination.Country, &destination.City, &destination.ItineraryID, &destination.ArrivalDate, &destination.DepartureDate, &destination.CreationDate, &destination.UpdateDate)
		if err != nil {
			log.Errorf("Error scanning itinerary travel destination: %v", err)
			destRows.Close()
			return nil, err
		}
		travelDestinations = append(travelDestinations, &destination)
	}
	destRows.Close()

	return travelDestinations, nil
}

func (d *ItineraryTravelDestination) defaultCreate(tx *sql.Tx) error {
	query := `INSERT INTO itinerary_travel_destinations (country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date)
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing insert for itinerary travel destination: %v", err)
		return err
	}

	defer stmt.Close()

	result, err := stmt.Exec(d.Country, d.City, d.ItineraryID, d.ArrivalDate, d.DepartureDate, time.Now(), time.Now())
	if err != nil {
		log.Errorf("Error executing insert for itinerary travel destination: %v", err)
		return err
	}

	travelDestinationId, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID for itinerary travel destination: %v", err)
		return err
	}

	d.ID = travelDestinationId

	return nil
}

func (d *ItineraryTravelDestination) defaultUpdate() error {
	query := `UPDATE itinerary_travel_destinations SET country = ?, city = ?, arrival_date = ?, departure_date = ?, update_date = ? WHERE id = ?`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing update for itinerary travel destination: %v", err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.Country, d.City, d.ArrivalDate, d.DepartureDate, time.Now(), d.ID)
	if err != nil {
		log.Errorf("Error executing update for itinerary travel destination: %v", err)
		return err
	}

	return nil
}

func (d *ItineraryTravelDestination) defaultDelete() error {
	query := `DELETE FROM itinerary_travel_destinations WHERE id = ?`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing delete for itinerary travel destination: %v", err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(d.ID)
	if err != nil {
		log.Errorf("Error executing delete for itinerary travel destination: %v", err)
		return err
	}

	return nil
}

func (d *ItineraryTravelDestination) defaultDeleteByItineraryIdTx(itineraryId int64, tx *sql.Tx) error {
	query := `DELETE FROM itinerary_travel_destinations WHERE itinerary_id = ?`

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing delete for itinerary travel destinations by itinerary ID: %v", err)
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(itineraryId)
	if err != nil {
		log.Errorf("Error executing delete for itinerary travel destinations by itinerary ID: %v", err)
		return err
	}

	return nil
}
