package models

import (
	"time"

	"example.com/travel-advisor/db"
	log "github.com/sirupsen/logrus"
)

type Itinerary struct {
	ID                 int64                         `json:"id" example:"1"`
	Title              string                        `json:"title" binding:"required" example:"Trip to Spain"`
	Description        string                        `json:"description" example:"Summer vacation in Spain"`
	CreationDate       *time.Time                    `json:"creationDate,omitempty" example:"2024-06-01T00:00:00Z"`
	UpdateDate         *time.Time                    `json:"updateDate,omitempty" example:"2024-06-01T00:00:00Z"`
	TravelDestinations []*ItineraryTravelDestination `json:"travelDestinations,omitempty"`
	OwnerID            int64                         `json:"ownerId" example:"1"`
	Notes              *string                       `json:"notes,omitempty" example:"I want to enjoy the nightlife"`

	FindById            func(id int64, includeDestinations bool) (*Itinerary, error) `json:"-"`
	FindLightweightById func(id int64) (*Itinerary, error)                           `json:"-"`
	FindByOwnerId       func(ownerId int64) ([]*Itinerary, error)                    `json:"-"`
	Create              func() error                                                 `json:"-"`
	Update              func() error                                                 `json:"-"`
	Delete              func() error                                                 `json:"-"`
}

var InitItinerary = func() *Itinerary {
	return InitItineraryFunctions(&Itinerary{})
}

var InitItineraryFunctions = func(itinerary *Itinerary) *Itinerary {
	// Set default SQL implementations for FindById, FindByOwnerId, Create, Update and Delete. In the future there could be implementations for
	// other NoSQL DB systems like MongoDB
	itinerary.FindById = itinerary.defaultFindById
	itinerary.FindLightweightById = itinerary.defaultFindLightweightById
	itinerary.FindByOwnerId = itinerary.defaultFindByOwnerId
	itinerary.Create = itinerary.defaultCreate
	itinerary.Update = itinerary.defaultUpdate
	itinerary.Delete = itinerary.defaultDelete

	return itinerary
}

var NewItinerary = func(title string, description string, notes *string, travelDestinations []*ItineraryTravelDestination) *Itinerary {
	itinerary := &Itinerary{
		Title:              title,
		Description:        description,
		Notes:              notes,
		TravelDestinations: travelDestinations,
	}

	return InitItineraryFunctions(itinerary)

}

func (i *Itinerary) defaultFindById(id int64, includeDestinations bool) (*Itinerary, error) {
	query := `SELECT id, title, description, notes, owner_id, creation_date, update_date
	FROM itineraries WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	itinerary := &Itinerary{}
	err := row.Scan(&itinerary.ID, &itinerary.Title, &itinerary.Description, &itinerary.Notes, &itinerary.OwnerID, &itinerary.CreationDate, &itinerary.UpdateDate)
	if err != nil {
		log.Errorf("Error fetching itinerary by ID %d: %v", id, err)
		return nil, err
	}

	if includeDestinations {
		// Fetch travel destinations for the itinerary
		destinationEntity := InitItineraryTravelDestination()

		travelDestinations, err := destinationEntity.FindByItineraryId(id)
		if err != nil {
			log.Errorf("Error fetching travel destinations for itinerary ID %d: %v", id, err)
			return nil, err
		}

		itinerary.TravelDestinations = travelDestinations
	}

	return itinerary, nil

}

func (i *Itinerary) defaultFindLightweightById(id int64) (*Itinerary, error) {
	query := `SELECT id, owner_id
	FROM itineraries WHERE id = ?`
	row := db.DB.QueryRow(query, id)

	itinerary := &Itinerary{}

	err := row.Scan(&itinerary.ID, &itinerary.OwnerID)
	if err != nil {
		log.Errorf("Error fetching lightweight itinerary by ID %d: %v", id, err)
		return nil, err
	}

	return itinerary, nil

}

func (i *Itinerary) defaultFindByOwnerId(ownerId int64) ([]*Itinerary, error) {
	query := `SELECT id, title, description, notes, owner_id, creation_date, update_date
	FROM itineraries WHERE owner_id = ?`

	rows, err := db.DB.Query(query, ownerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var itineraries []*Itinerary

	for rows.Next() {
		var itinerary Itinerary
		err := rows.Scan(&itinerary.ID, &itinerary.Title, &itinerary.Description, &itinerary.Notes, &itinerary.OwnerID, &itinerary.CreationDate, &itinerary.UpdateDate)
		if err != nil {
			return nil, err
		}

		// Fetch travel destinations for the itinerary
		destinationEntity := InitItineraryTravelDestination()

		travelDestinations, err := destinationEntity.FindByItineraryId(itinerary.ID)
		if err != nil {
			log.Errorf("Error fetching travel destinations for itinerary ID %d: %v", itinerary.ID, err)
			return nil, err
		}

		itinerary.TravelDestinations = travelDestinations

		itineraries = append(itineraries, &itinerary)
	}

	return itineraries, nil
}

func (i *Itinerary) defaultCreate() error {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction for itinerary creation: %v", err)
		return err
	}

	defer db.HandleTransaction(tx, &err)
	if err != nil {
		log.Errorf("Error starting transaction for itinerary creation: %v", err)
		return err
	}

	queryItinerary := `INSERT INTO itineraries(title, description, notes, owner_id, creation_date, update_date) 
	VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(queryItinerary)
	if err != nil {
		log.Errorf("Error preparing insert for itinerary: %v", err)
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(i.Title, i.Description, i.Notes, i.OwnerID, time.Now(), time.Now())
	if err != nil {
		log.Errorf("Error executing insert for itinerary: %v", err)
		return err
	}

	itineraryId, err := result.LastInsertId()
	if err != nil {
		log.Errorf("Error getting last insert ID for itinerary: %v", err)
		return err
	}

	i.ID = itineraryId

	for idx := range i.TravelDestinations {
		i.TravelDestinations[idx].ItineraryID = itineraryId
		err = i.TravelDestinations[idx].Create(tx)
		if err != nil {
			log.Errorf("Error creating travel destination for itinerary ID %d: %v", itineraryId, err)
			return err
		}
	}

	return nil
}

func (i *Itinerary) defaultUpdate() error {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction for itinerary update: %v", err)
		return err
	}

	defer db.HandleTransaction(tx, &err)
	if err != nil {
		log.Errorf("Error starting transaction for itinerary update: %v", err)
		return err
	}

	query := `UPDATE itineraries SET title = ?, description = ?, notes = ?, update_date = ? WHERE id = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing update for itinerary: %v", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Title, i.Description, i.Notes, time.Now(), i.ID)
	if err != nil {
		log.Errorf("Error executing update for itinerary ID %d: %v", i.ID, err)
		return err
	}

	destination := InitItineraryTravelDestination()

	// Clear existing travel destinations for this itinerary
	err = destination.DeleteByItineraryIdTx(i.ID, tx)
	if err != nil {
		log.Errorf("Error deleting existing travel destinations for itinerary ID %d: %v", i.ID, err)
		return err
	}

	// Insert new travel destinations
	for idx := range i.TravelDestinations {
		(i.TravelDestinations)[idx].ItineraryID = i.ID
		err = (i.TravelDestinations)[idx].Create(tx)
		if err != nil {
			log.Errorf("Error creating travel destination for itinerary ID %d: %v", i.ID, err)
			return err
		}
	}

	return nil
}

func (i *Itinerary) defaultDelete() error {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Errorf("Error starting transaction for itinerary deletion: %v", err)
		return err
	}

	defer db.HandleTransaction(tx, &err)
	if err != nil {
		log.Errorf("Error starting transaction for itinerary deletion: %v", err)
		return err
	}

	// Mark all jobs of itinerary for full future deletion
	job := InitItineraryFileJob()
	err = job.SoftDeleteJobsByItineraryIdTx(i.ID, tx)
	if err != nil {
		log.Errorf("Error marking jobs for deletion for itinerary ID %d: %v", i.ID, err)
		return err
	}

	// Clear existing travel destinations for this itinerary
	destination := InitItineraryTravelDestination()
	err = destination.DeleteByItineraryIdTx(i.ID, tx)
	if err != nil {
		log.Errorf("Error deleting existing travel destinations for itinerary ID %d: %v", i.ID, err)
		return err
	}

	// Delete itinerary
	query := `DELETE FROM itineraries WHERE id = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Errorf("Error preparing delete for itinerary ID %d: %v", i.ID, err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.ID)
	if err != nil {
		log.Errorf("Error executing delete for itinerary ID %d: %v", i.ID, err)
		return err
	}

	return nil
}
