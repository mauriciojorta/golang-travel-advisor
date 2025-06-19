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

func TestItineraryDefaultFindById_WithDestinationsSuccess(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	now := time.Now()
	itinerary := &Itinerary{ID: 1}

	// Mock main itinerary row
	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", "A test trip", 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	// Mock travel destinations rows
	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}).
				AddRow(10, "Country1", "City1", 1, now, now.Add(12*time.Hour), time.Now(), time.Now().Add(2*time.Hour)).
				AddRow(11, "Country2", "City2", 1, now.Add(12*time.Hour), now.Add(24*time.Hour), time.Now(), time.Now().Add(2*time.Hour)),
		)

	err = itinerary.defaultFindById(true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, "Test Title", itinerary.Title)
	assert.Equal(t, "Test Description", itinerary.Description)
	assert.Equal(t, "A test trip", *itinerary.Notes)
	assert.Equal(t, int64(2), itinerary.OwnerID)
	assert.Len(t, (*itinerary.TravelDestinations), 2)
	assert.Equal(t, int64(10), (*itinerary.TravelDestinations)[0].ID)
	assert.Equal(t, "Country1", (*itinerary.TravelDestinations)[0].Country)
	assert.Equal(t, "City1", (*itinerary.TravelDestinations)[0].City)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_WithoutDestinationsSuccess(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	// Mock main itinerary row
	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", "A test trip", 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	err = itinerary.defaultFindById(false)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, "Test Title", itinerary.Title)
	assert.Equal(t, "Test Description", itinerary.Description)
	assert.Equal(t, "A test trip", *itinerary.Notes)
	assert.Equal(t, int64(2), itinerary.OwnerID)
	assert.Nil(t, itinerary.TravelDestinations)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_WithDestinationsSuccessNullNotes(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	now := time.Now()
	itinerary := &Itinerary{ID: 1}

	// Mock main itinerary row
	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", nil, 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	// Mock travel destinations rows
	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}).
				AddRow(10, "Country1", "City1", 1, now, now.Add(12*time.Hour), time.Now(), time.Now().Add(2*time.Hour)).
				AddRow(11, "Country2", "City2", 1, now.Add(12*time.Hour), now.Add(24*time.Hour), time.Now(), time.Now().Add(2*time.Hour)),
		)

	err = itinerary.defaultFindById(true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, "Test Title", itinerary.Title)
	assert.Equal(t, "Test Description", itinerary.Description)
	assert.Nil(t, itinerary.Notes)
	assert.Equal(t, int64(2), itinerary.OwnerID)
	assert.Len(t, (*itinerary.TravelDestinations), 2)
	assert.Equal(t, int64(10), (*itinerary.TravelDestinations)[0].ID)
	assert.Equal(t, "Country1", (*itinerary.TravelDestinations)[0].Country)
	assert.Equal(t, "City1", (*itinerary.TravelDestinations)[0].City)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_RowScanError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(itinerary.ID).
		WillReturnError(sql.ErrNoRows)

	err = itinerary.defaultFindById(true)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_DestinationsQueryError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", nil, 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(itinerary.ID).
		WillReturnError(sql.ErrConnDone)

	err = itinerary.defaultFindById(true)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_DestinationScanError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	now := time.Now()
	itinerary := &Itinerary{ID: 1}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", nil, 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	// Return a row with a wrong type to force scan error
	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(itinerary.ID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}).
				AddRow(10, "Country1", "City1", 1, now, now.Add(12*time.Hour), time.Now(), time.Now().Add(2*time.Hour)).
				AddRow("not-an-int", "Country2", "City2", 1, now.Add(12*time.Hour), now.Add(24*time.Hour), time.Now(), time.Now().Add(2*time.Hour)),
		)

	err = itinerary.defaultFindById(true)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(itinerary.OwnerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", "A test trip", 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}))

	// Act
	itineraries, err := itinerary.defaultFindByOwnerId()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, *itineraries, 1)
	assert.Equal(t, int64(1), (*itineraries)[0].ID)
	assert.Equal(t, "Test Title", (*itineraries)[0].Title)
	assert.Equal(t, "Test Description", (*itineraries)[0].Description)
	assert.Equal(t, "A test trip", *(*itineraries)[0].Notes)
	assert.NotNil(t, *(*itineraries)[0].CreationDate)
	assert.NotNil(t, *(*itineraries)[0].UpdateDate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_SuccessNullNotes(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(itinerary.OwnerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", nil, 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}))

	// Act
	itineraries, err := itinerary.defaultFindByOwnerId()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, *itineraries, 1)
	assert.Equal(t, int64(1), (*itineraries)[0].ID)
	assert.Equal(t, "Test Title", (*itineraries)[0].Title)
	assert.Equal(t, "Test Description", (*itineraries)[0].Description)
	assert.Nil(t, (*itineraries)[0].Notes)
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

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(itinerary.OwnerID).
		WillReturnError(sql.ErrNoRows)

	// Act
	_, err = itinerary.defaultFindByOwnerId()

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

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(itinerary.OwnerID).
		WillReturnError(assert.AnError)

	// Act
	_, err = itinerary.defaultFindByOwnerId()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindByOwnerId_DestinationQueryError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(itinerary.OwnerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", nil, 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	// Act
	_, err = itinerary.defaultFindByOwnerId()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, sql.ErrConnDone, err)
	assert.NoError(t, mock.ExpectationsWereMet())

}

func TestItineraryDefaultFindByOwnerId_DestinationScanError(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		OwnerID: 1,
	}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(itinerary.OwnerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", nil, 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date"}).
			AddRow("not-an-int", "Country1", "City1", 1, time.Now(), time.Now().Add(12*time.Hour)))

	// Act
	_, err = itinerary.defaultFindByOwnerId()

	// Assert
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

}

func TestItineraryItinerary_Create_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	testNotes := "A test trip"
	itinerary := &Itinerary{
		Title:       "Test Title",
		Description: "Test Description",
		Notes:       &testNotes,
		OwnerID:     1,
		TravelDestinations: &[]ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", "A test trip", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for _, destination := range *itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, destination.ItineraryID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	err = itinerary.defaultCreate()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryItinerary_Create_SuccessNullNotes(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		Title:       "Test Title",
		Description: "Test Description",
		Notes:       nil,
		OwnerID:     1,
		TravelDestinations: &[]ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for _, destination := range *itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, destination.ItineraryID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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
		Title:       "Test Title",
		Description: "Test Description",
		OwnerID:     1,
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
		Title:       "Test Title",
		Description: "Test Description",
		OwnerID:     1,
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
		WithArgs("Test Title", "Test Description", "A test trip", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("exec error"))

	mock.ExpectRollback()

	testNotes := "A test trip"
	itinerary := &Itinerary{
		Title:       "Test Title",
		Description: "Test Description",
		Notes:       &testNotes,
		OwnerID:     1,
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
		Title:       "Test Title",
		Description: "Test Description",
		OwnerID:     1,
		TravelDestinations: &[]ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", nil, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	testNotes := "A test trip"
	itinerary := &Itinerary{
		ID:          1,
		Title:       "Updated Title",
		Description: "Updated Description",
		Notes:       &testNotes,
		TravelDestinations: &[]ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, notes = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, itinerary.Notes, sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectPrepare(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`).ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	for _, destination := range *itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, destination.ItineraryID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryUpdate_SuccessNullNotes(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{
		ID:          1,
		Title:       "Updated Title",
		Description: "Updated Description",
		Notes:       nil,
		TravelDestinations: &[]ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, notes = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, itinerary.Notes, sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectPrepare(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`).ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	for _, destination := range *itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, destination.ItineraryID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, notes = \?, update_date = \? WHERE id = \?`).
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

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, notes = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, sqlmock.AnyArg(), sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectPrepare(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`).ExpectExec().
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
		ID:          1,
		Title:       "Updated Title",
		Description: "Updated Description",
		TravelDestinations: &[]ItineraryTravelDestination{
			{ID: 1, Country: "Country 1", City: "City 1", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
			{ID: 2, Country: "Country 2", City: "City 2", ItineraryID: 1, ArrivalDate: time.Now(), DepartureDate: time.Now().Add(24 * time.Hour)},
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare(`UPDATE itineraries SET title = \?, description = \?, notes = \?, update_date = \? WHERE id = \?`).
		ExpectExec().
		WithArgs(itinerary.Title, itinerary.Description, sqlmock.AnyArg(), sqlmock.AnyArg(), itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectPrepare(`DELETE FROM itinerary_travel_destinations WHERE itinerary_id = \?`).ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("insert destinations error"))

	mock.ExpectRollback()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "insert destinations error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultDelete_Success_WithJobs(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	// Mock transaction begin
	mock.ExpectBegin()

	// Mock SoftDeleteJobsByItineraryIdTx
	originalNewItineraryFileJob := NewItineraryFileJob
	defer func() { NewItineraryFileJob = originalNewItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	NewItineraryFileJob = func(itineraryID int64) *ItineraryFileJob {
		assert.Equal(t, int64(1), itineraryID)
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	// Mock DeleteByItineraryIdTx
	originalNewItineraryTravelDestination := NewItineraryTravelDestination
	defer func() { NewItineraryTravelDestination = originalNewItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	NewItineraryTravelDestination = func(country, city string, itineraryID int64, arrival, departure time.Time) *ItineraryTravelDestination {
		assert.Equal(t, int64(1), itineraryID)
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	// Mock DELETE FROM itineraries
	mock.ExpectPrepare("DELETE FROM itineraries WHERE id = \\?").
		ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err = itinerary.defaultDelete()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultDelete_BeginError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	err = itinerary.defaultDelete()
	assert.Error(t, err)
	assert.Equal(t, "begin error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultDelete_SoftDeleteJobsError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectBegin()

	originalNewItineraryFileJob := NewItineraryFileJob
	defer func() { NewItineraryFileJob = originalNewItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	NewItineraryFileJob = func(itineraryID int64) *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error {
		return errors.New("soft delete jobs error")
	}

	err = itinerary.defaultDelete()
	assert.Error(t, err)
	assert.Equal(t, "soft delete jobs error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultDelete_DeleteDestinationsError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectBegin()

	originalNewItineraryFileJob := NewItineraryFileJob
	defer func() { NewItineraryFileJob = originalNewItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	NewItineraryFileJob = func(itineraryID int64) *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	originalNewItineraryTravelDestination := NewItineraryTravelDestination
	defer func() { NewItineraryTravelDestination = originalNewItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	NewItineraryTravelDestination = func(country, city string, itineraryID int64, arrival, departure time.Time) *ItineraryTravelDestination {
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(tx *sql.Tx) error {
		return errors.New("delete destinations error")
	}

	err = itinerary.defaultDelete()
	assert.Error(t, err)
	assert.Equal(t, "delete destinations error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultDelete_PrepareDeleteItineraryError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectBegin()

	originalNewItineraryFileJob := NewItineraryFileJob
	defer func() { NewItineraryFileJob = originalNewItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	NewItineraryFileJob = func(itineraryID int64) *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	originalNewItineraryTravelDestination := NewItineraryTravelDestination
	defer func() { NewItineraryTravelDestination = originalNewItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	NewItineraryTravelDestination = func(country, city string, itineraryID int64, arrival, departure time.Time) *ItineraryTravelDestination {
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	mock.ExpectPrepare("DELETE FROM itineraries WHERE id = \\?").
		WillReturnError(errors.New("prepare delete itinerary error"))

	mock.ExpectRollback()

	err = itinerary.defaultDelete()
	assert.Error(t, err)
	assert.Equal(t, "prepare delete itinerary error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultDelete_ExecDeleteItineraryError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectBegin()

	originalNewItineraryFileJob := NewItineraryFileJob
	defer func() { NewItineraryFileJob = originalNewItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	NewItineraryFileJob = func(itineraryID int64) *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	originalNewItineraryTravelDestination := NewItineraryTravelDestination
	defer func() { NewItineraryTravelDestination = originalNewItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	NewItineraryTravelDestination = func(country, city string, itineraryID int64, arrival, departure time.Time) *ItineraryTravelDestination {
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(tx *sql.Tx) error {
		return nil
	}

	mock.ExpectPrepare("DELETE FROM itineraries WHERE id = \\?").
		ExpectExec().
		WithArgs(itinerary.ID).
		WillReturnError(errors.New("exec delete itinerary error"))

	mock.ExpectRollback()

	err = itinerary.defaultDelete()
	assert.Error(t, err)
	assert.Equal(t, "exec delete itinerary error", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultExistById_Exists(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectQuery("SELECT 1 FROM itineraries WHERE id = \\? LIMIT 1").
		WithArgs(itinerary.ID).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := itinerary.defaultExistById()
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultExistById_NotExists(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectQuery("SELECT 1 FROM itineraries WHERE id = \\? LIMIT 1").
		WithArgs(itinerary.ID).
		WillReturnError(sql.ErrNoRows)

	exists, err := itinerary.defaultExistById()
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultExistById_ScanError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}

	mock.ExpectQuery("SELECT 1 FROM itineraries WHERE id = \\? LIMIT 1").
		WithArgs(itinerary.ID).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("not-an-int"))

	exists, err := itinerary.defaultExistById()
	assert.Error(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultValidateOwnership_OwnerExists(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}
	currentUserId := int64(2)

	mock.ExpectQuery("SELECT 1 FROM itineraries WHERE id = \\? AND owner_id = \\? LIMIT 1").
		WithArgs(itinerary.ID, currentUserId).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := itinerary.defaultValidateOwnership(currentUserId)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultValidateOwnership_NotOwner(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}
	currentUserId := int64(2)

	mock.ExpectQuery("SELECT 1 FROM itineraries WHERE id = \\? AND owner_id = \\? LIMIT 1").
		WithArgs(itinerary.ID, currentUserId).
		WillReturnError(sql.ErrNoRows)

	exists, err := itinerary.defaultValidateOwnership(currentUserId)
	assert.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultValidateOwnership_ScanError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{ID: 1}
	currentUserId := int64(2)

	mock.ExpectQuery("SELECT 1 FROM itineraries WHERE id = \\? AND owner_id = \\? LIMIT 1").
		WithArgs(itinerary.ID, currentUserId).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow("not-an-int"))

	exists, err := itinerary.defaultValidateOwnership(currentUserId)
	assert.Error(t, err)
	assert.False(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}
