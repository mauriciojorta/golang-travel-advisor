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

func TestItineraryFindByOwnerId_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	rows := sqlmock.NewRows([]string{"id", "title", "description", "travel_start_date", "travel_end_date", "owner_id"}).
		AddRow(1, "Test Title", "Test Description", time.Now(), time.Now(), 1)

	mock.ExpectQuery("SELECT id, title, description, travel_start_date, travel_end_date, owner_id FROM itineraries WHERE owner_id = ?").
		WithArgs(itinerary.OwnerID).
		WillReturnRows(rows)

	// Act
	err = itinerary.defaultFindByOwnerId()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, "Test Title", itinerary.Title)
	assert.Equal(t, "Test Description", itinerary.Description)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_NoRows(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	mock.ExpectQuery("SELECT id, title, description, travel_start_date, travel_end_date, owner_id FROM itineraries WHERE owner_id = ?").
		WithArgs(itinerary.OwnerID).
		WillReturnError(sql.ErrNoRows)

	// Act
	err = itinerary.defaultFindByOwnerId()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_QueryError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	mock.ExpectQuery("SELECT id, title, description, travel_start_date, travel_end_date, owner_id FROM itineraries WHERE owner_id = ?").
		WithArgs(itinerary.OwnerID).
		WillReturnError(assert.AnError)

	// Act
	err = itinerary.defaultFindByOwnerId()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryItinerary_Create_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		Title:           "Test Title",
		Description:     "Test Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(48 * time.Hour),
		OwnerID:         1,
		TravelDestinations: []ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for _, destination := range itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, destination.ItineraryID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	err = itinerary.defaultCreate()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryItinerary_Create_BeginTransactionError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectBegin().WillReturnError(errors.New("transaction error"))

	itinerary := &Itinerary{
		Title:           "Test Title",
		Description:     "Test Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(48 * time.Hour),
		OwnerID:         1,
	}

	err = itinerary.defaultCreate()
	assert.Error(t, err)
	assert.Equal(t, "transaction error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryItinerary_Create_PrepareError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		WillReturnError(errors.New("prepare error"))

	mock.ExpectRollback()

	itinerary := &Itinerary{
		Title:           "Test Title",
		Description:     "Test Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(48 * time.Hour),
		OwnerID:         1,
	}

	err = itinerary.defaultCreate()
	assert.Error(t, err)
	assert.Equal(t, "prepare error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryItinerary_Create_ExecError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("exec error"))

	mock.ExpectRollback()

	itinerary := &Itinerary{
		Title:           "Test Title",
		Description:     "Test Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(48 * time.Hour),
		OwnerID:         1,
	}

	err = itinerary.defaultCreate()
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryItinerary_Create_TravelDestinationError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		Title:           "Test Title",
		Description:     "Test Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(48 * time.Hour),
		OwnerID:         1,
		TravelDestinations: []ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", sqlmock.AnyArg(), sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert destinations error"))

	mock.ExpectRollback()

	err = itinerary.defaultCreate()
	assert.Error(t, err)
	assert.Equal(t, "insert destinations error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryUpdate_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID:              1,
		Title:           "Updated Title",
		Description:     "Updated Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(24 * time.Hour),
		TravelDestinations: []ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, travel_start_date = \?, travel_end_date = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, itinerary.TravelStartDate, itinerary.TravelEndDate, sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = ?`).
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	for _, destination := range itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, destination.ItineraryID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryUpdate_PrepareStatementError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID: 1,
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, travel_start_date = \?, travel_end_date = \?, update_date = \? WHERE id = \?`).
		WillReturnError(errors.New("prepare statement error"))

	mock.ExpectRollback()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "prepare statement error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryUpdate_DeleteDestinationsError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID: 1,
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, travel_start_date = \?, travel_end_date = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, itinerary.TravelStartDate, itinerary.TravelEndDate, sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`).
		WithArgs(itinerary.ID).
		WillReturnError(errors.New("delete destinations error"))

	mock.ExpectRollback()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "delete destinations error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryUpdate_InsertDestinationsError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID:              1,
		Title:           "Updated Title",
		Description:     "Updated Description",
		TravelStartDate: time.Now(),
		TravelEndDate:   time.Now().Add(24 * time.Hour),
		TravelDestinations: []ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, travel_start_date = \?, travel_end_date = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, itinerary.TravelStartDate, itinerary.TravelEndDate, sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`).
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert destinations error"))

	mock.ExpectRollback()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "insert destinations error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDelete_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID: 1,
	}

	mock.ExpectPrepare("DELETE FROM itineraries WHERE id = ?").
		ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = itinerary.defaultDelete()

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDelete_PrepareError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID: 1,
	}

	mock.ExpectPrepare("DELETE FROM itineraries WHERE id = ?").
		WillReturnError(errors.New("prepare statement error"))

	// Act
	err = itinerary.defaultDelete()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "prepare statement error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDelete_ExecError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID: 1,
	}

	mock.ExpectPrepare("DELETE FROM itineraries WHERE id = ?").
		ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnError(errors.New("exec error"))

	// Act
	err = itinerary.defaultDelete()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}
