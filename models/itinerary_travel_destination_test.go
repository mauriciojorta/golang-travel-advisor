package models

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"example.com/travel-advisor/db"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDestinationTravelDestination_Find_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ItineraryID: 1,
	}

	rows := sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date"}).
		AddRow(1, "Test Country", "Test City", 1, time.Now(), time.Now().Add(48*time.Hour))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date FROM itinerary_travel_destinations WHERE itinerary_id = ?").
		WithArgs(destination.ItineraryID).
		WillReturnRows(rows)

	// Act
	err = destination.defaultFind()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), destination.ID)
	assert.Equal(t, "Test Country", destination.Country)
	assert.Equal(t, "Test City", destination.City)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationTravelDestination_Find_NoRows(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ItineraryID: 1,
	}

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date FROM itinerary_travel_destinations WHERE itinerary_id = ?").
		WithArgs(destination.ItineraryID).
		WillReturnError(sql.ErrNoRows)

	// Act
	err = destination.defaultFind()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationTravelDestination_Find_QueryError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ItineraryID: 1,
	}

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date FROM itinerary_travel_destinations WHERE itinerary_id = ?").
		WithArgs(destination.ItineraryID).
		WillReturnError(errors.New("query error"))

	// Act
	err = destination.defaultFind()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "query error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationCreate_Success(t *testing.T) {
	// Arrange
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
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	destination := &ItineraryTravelDestination{
		Country:       "USA",
		City:          "New York",
		ItineraryID:   1,
		ArrivalDate:   time.Now(),
		DepartureDate: time.Now().Add(24 * time.Hour),
	}

	query := `INSERT INTO itinerary_travel_destinations \(country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date\) VALUES \(\?, \?, \?, \?, \?, \?, \?\)`
	mock.ExpectPrepare(query).ExpectExec().
		WithArgs(destination.Country, destination.City, destination.ItineraryID, destination.ArrivalDate, destination.DepartureDate, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Act
	err = destination.defaultCreate(tx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), destination.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationCreate_PrepareError(t *testing.T) {
	// Arrange
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
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	destination := &ItineraryTravelDestination{
		Country:       "USA",
		City:          "New York",
		ItineraryID:   1,
		ArrivalDate:   time.Now(),
		DepartureDate: time.Now().Add(24 * time.Hour),
	}

	query := `INSERT INTO itinerary_travel_destinations \(country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date\) VALUES \(\?, \?, \?, \?, \?, \?, \?\)`
	mock.ExpectPrepare(query).WillReturnError(sql.ErrConnDone)

	// Act
	err = destination.defaultCreate(tx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationCreate_ExecError(t *testing.T) {
	// Arrange
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
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	destination := &ItineraryTravelDestination{
		Country:       "USA",
		City:          "New York",
		ItineraryID:   1,
		ArrivalDate:   time.Now(),
		DepartureDate: time.Now().Add(24 * time.Hour),
	}

	query := `INSERT INTO itinerary_travel_destinations \(country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date\) VALUES \(\?, \?, \?, \?, \?, \?, \?\)`
	mock.ExpectPrepare(query)
	mock.ExpectExec(query).
		WithArgs(destination.Country, destination.City, destination.ItineraryID, destination.ArrivalDate, destination.DepartureDate, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	// Act
	err = destination.defaultCreate(tx)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationUpdate_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ID:            1,
		Country:       "USA",
		City:          "New York",
		ArrivalDate:   time.Now(),
		DepartureDate: time.Now().Add(24 * time.Hour),
	}

	query := `UPDATE itinerary_travel_destinations SET country = \?, city = \?, arrival_date = \?, departure_date = \?, update_date = \? WHERE id = \?`
	mock.ExpectPrepare(query).ExpectExec().
		WithArgs(destination.Country, destination.City, destination.ArrivalDate, destination.DepartureDate, sqlmock.AnyArg(), destination.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = destination.defaultUpdate()

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationUpdate_PrepareError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ID:            1,
		Country:       "USA",
		City:          "New York",
		ArrivalDate:   time.Now(),
		DepartureDate: time.Now().Add(24 * time.Hour),
	}

	query := `UPDATE itinerary_travel_destinations SET country = \?, city = \?, arrival_date = \?, departure_date = \?, update_date = \? WHERE id = \?`
	mock.ExpectPrepare(query).WillReturnError(sql.ErrConnDone)

	// Act
	err = destination.defaultUpdate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationUpdate_ExecError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ID:            1,
		Country:       "USA",
		City:          "New York",
		ArrivalDate:   time.Now(),
		DepartureDate: time.Now().Add(24 * time.Hour),
	}

	query := `UPDATE itinerary_travel_destinations SET country = \?, city = \?, arrival_date = \?, departure_date = \?, update_date = \? WHERE id = \?`
	mock.ExpectPrepare(query)
	mock.ExpectExec(query).
		WithArgs(destination.Country, destination.City, destination.ArrivalDate, destination.DepartureDate, sqlmock.AnyArg(), destination.ID).
		WillReturnError(sql.ErrNoRows)

	// Act
	err = destination.defaultUpdate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationDelete_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ID: 1,
	}

	query := `DELETE FROM itinerary_travel_destinations WHERE id = \?`
	mock.ExpectPrepare(query).ExpectExec().
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = destination.defaultDelete()

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationDelete_PrepareError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ID: 1,
	}

	query := `DELETE FROM itinerary_travel_destinations WHERE id = \?`
	mock.ExpectPrepare(query).WillReturnError(sql.ErrConnDone)

	// Act
	err = destination.defaultDelete()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationDelete_ExecError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	destination := &ItineraryTravelDestination{
		ID: 1,
	}

	query := `DELETE FROM itinerary_travel_destinations WHERE id = \?`
	mock.ExpectPrepare(query).ExpectExec().
		WithArgs(destination.ID).
		WillReturnError(sql.ErrNoRows)

	// Act
	err = destination.defaultDelete()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationDeleteByItineraryIdTx_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	mock.ExpectBegin()

	tx, err := dbMock.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	itineraryId := int64(1)

	query := `DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`
	mock.ExpectPrepare(query).ExpectExec().
		WithArgs(itineraryId).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = DeleteByItineraryIdTx(tx, itineraryId)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationDeleteByItineraryIdTx_PrepareError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	mock.ExpectBegin()

	tx, err := dbMock.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	itineraryId := int64(1)

	query := `DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`
	mock.ExpectPrepare(query).WillReturnError(sql.ErrConnDone)

	// Act
	err = DeleteByItineraryIdTx(tx, itineraryId)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDestinationDeleteByItineraryIdTx_ExecError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	mock.ExpectBegin()

	tx, err := dbMock.Begin()
	assert.NoError(t, err)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	itineraryId := int64(1)

	query := `DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`
	mock.ExpectPrepare(query)
	mock.ExpectExec(query).
		WithArgs(itineraryId).
		WillReturnError(sql.ErrNoRows)

	// Act
	err = DeleteByItineraryIdTx(tx, itineraryId)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestValidateItineraryDestinationsDates_EmptySlice(t *testing.T) {
	destinations := []ItineraryTravelDestination{}
	err := ValidateItineraryDestinationsDates(&destinations)
	assert.Error(t, err)
	assert.Equal(t, "At least one destination is required", err.Error())
}

func TestValidateItineraryDestinationsDates_TooManyDestinations(t *testing.T) {
	destinations := make([]ItineraryTravelDestination, 21)
	for i := range destinations {
		destinations[i] = ItineraryTravelDestination{
			ArrivalDate:   time.Now(),
			DepartureDate: time.Now().Add(24 * time.Hour),
		}
	}
	err := ValidateItineraryDestinationsDates(&destinations)
	assert.Error(t, err)
	assert.Equal(t, "The itinerary cannot have more than 20 destinations", err.Error())
}

func TestValidateItineraryDestinationsDates_Over30Days(t *testing.T) {
	now := time.Now()
	destinations := []ItineraryTravelDestination{
		{
			ArrivalDate:   now,
			DepartureDate: now.Add(10 * 24 * time.Hour),
		},
		{
			ArrivalDate:   now.Add(40 * 24 * time.Hour),
			DepartureDate: now.Add(41 * 24 * time.Hour),
		},
	}
	err := ValidateItineraryDestinationsDates(&destinations)
	assert.Error(t, err)
	assert.Equal(t, "The itinerary cannot span more than 30 days", err.Error())
}

func TestValidateItineraryDestinationsDates_ValidSingleDestination(t *testing.T) {
	now := time.Now()
	destinations := []ItineraryTravelDestination{
		{
			ArrivalDate:   now,
			DepartureDate: now.Add(5 * 24 * time.Hour),
		},
	}
	err := ValidateItineraryDestinationsDates(&destinations)
	assert.NoError(t, err)
}

func TestValidateItineraryDestinationsDates_ValidMultipleDestinations(t *testing.T) {
	now := time.Now()
	destinations := []ItineraryTravelDestination{
		{
			ArrivalDate:   now,
			DepartureDate: now.Add(5 * 24 * time.Hour),
		},
		{
			ArrivalDate:   now.Add(2 * 24 * time.Hour),
			DepartureDate: now.Add(10 * 24 * time.Hour),
		},
		{
			ArrivalDate:   now.Add(8 * 24 * time.Hour),
			DepartureDate: now.Add(15 * 24 * time.Hour),
		},
	}
	err := ValidateItineraryDestinationsDates(&destinations)
	assert.NoError(t, err)
}
