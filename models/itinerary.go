package models

import (
	"time"

	"example.com/travel-advisor/db"
)

type Itinerary struct {
	ID                 int64                        `json:"id"`
	Title              string                       `json:"title" binding:"required"`
	Description        string                       `json:"description"`
	CreationDate       time.Time                    `json:"creationDate"`
	UpdateDate         time.Time                    `json:"updateDate"`
	TravelDestinations []ItineraryTravelDestination `json:"travelDestinations"`
	TravelStartDate    time.Time                    `json:"travelStartDate"`
	TravelEndDate      time.Time                    `json:"travelEndDate"`
	OwnerID            int64                        `json:"ownerId"`
	ItineraryFilePath  string                       `json:"itineraryFilePath"`

	FindByOwnerId func() error
	Create        func() error
	Update        func() error
	Delete        func() error
}

var NewItinerary = func(title string, description string, travelStartDate time.Time, travelEndDate time.Time, travelDestinations []ItineraryTravelDestination) *Itinerary {
	itinerary := &Itinerary{
		Title:              title,
		Description:        description,
		TravelStartDate:    travelStartDate,
		TravelEndDate:      travelEndDate,
		TravelDestinations: travelDestinations,
	}

	// Set default implementations for FindByOwnerId, Create, Update, Delete, and GenerateItineraryFile
	itinerary.FindByOwnerId = itinerary.defaultFindByOwnerId
	itinerary.Create = itinerary.defaultCreate
	itinerary.Update = itinerary.defaultUpdate
	itinerary.Delete = itinerary.defaultDelete

	return itinerary
}

func (i *Itinerary) defaultFindByOwnerId() error {

	query := `SELECT id, title, description, travel_start_date, travel_end_date, owner_id
	FROM itineraries WHERE owner_id = ?`
	row := db.DB.QueryRow(query, i.OwnerID)

	err := row.Scan(&i.ID, &i.Title, &i.Description, &i.TravelStartDate, &i.TravelEndDate, &i.OwnerID)

	if err != nil {
		return err
	}

	return nil
}

func (i *Itinerary) defaultCreate() error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	defer db.HandleTransaction(tx, &err)
	if err != nil {
		return err
	}

	queryItinerary := `INSERT INTO itineraries(title, description, travel_start_date, travel_end_date, owner_id, creation_date, update_date) 
	VALUES (?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(queryItinerary)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(i.Title, i.Description, i.TravelStartDate, i.TravelEndDate, i.OwnerID, time.Now(), time.Now())
	if err != nil {
		return err
	}

	itineraryId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	i.ID = itineraryId

	for _, travelDestination := range i.TravelDestinations {
		travelDestination = *NewItineraryTravelDestination(travelDestination.Country, travelDestination.City, travelDestination.ItineraryID, travelDestination.ArrivalDate, travelDestination.DepartureDate)
		err = travelDestination.Create(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Itinerary) defaultUpdate() error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	defer db.HandleTransaction(tx, &err)
	if err != nil {
		return err
	}

	query := `UPDATE itineraries SET title = ?, description = ?, travel_start_date = ?, travel_end_date = ?, update_date = ? WHERE id = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Title, i.Description, i.TravelStartDate, i.TravelEndDate, time.Now(), i.ID)
	if err != nil {
		return err
	}

	// Clear existing travel destinations for this itinerary
	err = DeleteByItineraryIdTx(tx, i.ID)
	if err != nil {
		return err
	}

	// Insert new travel destinations
	for _, travelDestination := range i.TravelDestinations {
		travelDestination = *NewItineraryTravelDestination(travelDestination.Country, travelDestination.City, travelDestination.ItineraryID, travelDestination.ArrivalDate, travelDestination.DepartureDate)
		err = travelDestination.Create(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Itinerary) defaultDelete() error {
	query := `DELETE FROM itineraries WHERE id = ?`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.ID)
	if err != nil {
		return err
	}

	return nil
}
