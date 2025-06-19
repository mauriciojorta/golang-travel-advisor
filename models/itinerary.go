package models

import (
	"database/sql"
	"time"

	"example.com/travel-advisor/db"
)

type Itinerary struct {
	ID                 int64                         `json:"id"`
	Title              string                        `json:"title" binding:"required"`
	Description        string                        `json:"description"`
	CreationDate       *time.Time                    `json:"creationDate"`
	UpdateDate         *time.Time                    `json:"updateDate"`
	TravelDestinations *[]ItineraryTravelDestination `json:"travelDestinations"`
	OwnerID            int64                         `json:"ownerId"`
	FileJobs           []ItineraryFileJob            `json:"fileJobs"`
	Notes              *string                       `json:"notes"`

	FindById          func(includeDestinations bool) error    `json:"-"`
	ExistById         func() (bool, error)                    `json:"-"`
	ValidateOwnership func(currentUserId int64) (bool, error) `json:"-"`
	FindByOwnerId     func() (*[]Itinerary, error)            `json:"-"`
	Create            func() error                            `json:"-"`
	Update            func() error                            `json:"-"`
	Delete            func() error                            `json:"-"`
}

var NewItinerary = func(title string, description string, notes *string, travelDestinations *[]ItineraryTravelDestination) *Itinerary {
	itinerary := &Itinerary{
		Title:              title,
		Description:        description,
		Notes:              notes,
		TravelDestinations: travelDestinations,
	}

	// Set default implementations for FindById, FindByOwnerId, Create, Update, Delete, and GenerateItineraryFile
	itinerary.FindById = itinerary.defaultFindById
	itinerary.ExistById = itinerary.defaultExistById
	itinerary.ValidateOwnership = itinerary.defaultValidateOwnership
	itinerary.FindByOwnerId = itinerary.defaultFindByOwnerId
	itinerary.Create = itinerary.defaultCreate
	itinerary.Update = itinerary.defaultUpdate
	itinerary.Delete = itinerary.defaultDelete

	return itinerary
}

func (i *Itinerary) defaultFindById(includeDestinations bool) error {
	query := `SELECT id, title, description, notes, owner_id, creation_date, update_date
	FROM itineraries WHERE id = ?`
	row := db.DB.QueryRow(query, i.ID)

	err := row.Scan(&i.ID, &i.Title, &i.Description, &i.Notes, &i.OwnerID, &i.CreationDate, &i.UpdateDate)
	if err != nil {
		return err
	}

	if includeDestinations {
		// Fetch travel destinations for the itinerary
		destinationEntity := NewItineraryTravelDestination("", "", i.ID, time.Now(), time.Now())

		travelDestinations, err := destinationEntity.FindByItineraryId()
		if err != nil {
			return err
		}

		i.TravelDestinations = travelDestinations
	}

	return nil

}

func (i *Itinerary) defaultExistById() (bool, error) {
	query := `SELECT 1 FROM itineraries WHERE id = ? LIMIT 1`
	row := db.DB.QueryRow(query, i.ID)

	var exists int
	err := row.Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (i *Itinerary) defaultValidateOwnership(currentUserId int64) (bool, error) {
	query := `SELECT 1 FROM itineraries WHERE id = ? AND owner_id = ? LIMIT 1`
	row := db.DB.QueryRow(query, i.ID, currentUserId)

	var exists int
	err := row.Scan(&exists)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (i *Itinerary) defaultFindByOwnerId() (*[]Itinerary, error) {
	query := `SELECT id, title, description, notes, owner_id, creation_date, update_date
	FROM itineraries WHERE owner_id = ?`

	rows, err := db.DB.Query(query, i.OwnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var itineraries []Itinerary

	for rows.Next() {
		var itinerary Itinerary
		err := rows.Scan(&itinerary.ID, &itinerary.Title, &itinerary.Description, &itinerary.Notes, &itinerary.OwnerID, &itinerary.CreationDate, &itinerary.UpdateDate)
		if err != nil {
			return nil, err
		}

		// Fetch travel destinations for the itinerary
		destinationEntity := NewItineraryTravelDestination("", "", itinerary.ID, time.Now(), time.Now())

		travelDestinations, err := destinationEntity.FindByItineraryId()
		if err != nil {
			return nil, err
		}

		itinerary.TravelDestinations = travelDestinations

		itineraries = append(itineraries, itinerary)
	}

	return &itineraries, nil
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

	queryItinerary := `INSERT INTO itineraries(title, description, notes, owner_id, creation_date, update_date) 
	VALUES (?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(queryItinerary)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(i.Title, i.Description, i.Notes, i.OwnerID, time.Now(), time.Now())
	if err != nil {
		return err
	}

	itineraryId, err := result.LastInsertId()
	if err != nil {
		return err
	}

	i.ID = itineraryId

	for _, travelDestination := range *i.TravelDestinations {
		travelDestination.ItineraryID = itineraryId
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

	query := `UPDATE itineraries SET title = ?, description = ?, notes = ?, update_date = ? WHERE id = ?`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Title, i.Description, i.Notes, time.Now(), i.ID)
	if err != nil {
		return err
	}

	destination := NewItineraryTravelDestination("", "", i.ID, time.Now(), time.Now())

	// Clear existing travel destinations for this itinerary
	err = destination.DeleteByItineraryIdTx(tx)
	if err != nil {
		return err
	}

	// Insert new travel destinations
	for _, travelDestination := range *i.TravelDestinations {
		travelDestination.ItineraryID = i.ID
		travelDestination = *NewItineraryTravelDestination(travelDestination.Country, travelDestination.City, travelDestination.ItineraryID, travelDestination.ArrivalDate, travelDestination.DepartureDate)
		err = travelDestination.Create(tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Itinerary) defaultDelete() error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	defer db.HandleTransaction(tx, &err)
	if err != nil {
		return err
	}

	// Mark all jobs of itinerary for full future deletion
	job := NewItineraryFileJob(i.ID)
	err = job.SoftDeleteJobsByItineraryIdTx(tx)
	if err != nil {
		return err
	}

	// Clear existing travel destinations for this itinerary
	destination := NewItineraryTravelDestination("", "", i.ID, time.Now(), time.Now())
	err = destination.DeleteByItineraryIdTx(tx)
	if err != nil {
		return err
	}

	// Delete itinerary
	query := `DELETE FROM itineraries WHERE id = ?`
	stmt, err := tx.Prepare(query)
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
