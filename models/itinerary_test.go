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
	itinerary := &Itinerary{}

	// Mock main itinerary row
	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", "A test trip", 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	// Mock travel destinations rows
	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}).
				AddRow(10, "Country1", "City1", 1, now, now.Add(12*time.Hour), time.Now(), time.Now().Add(2*time.Hour)).
				AddRow(11, "Country2", "City2", 1, now.Add(12*time.Hour), now.Add(24*time.Hour), time.Now(), time.Now().Add(2*time.Hour)),
		)

	it, err := itinerary.defaultFindById(1, true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), it.ID)
	assert.Equal(t, "Test Title", it.Title)
	assert.Equal(t, "Test Description", it.Description)
	assert.Equal(t, "A test trip", *it.Notes)
	assert.Equal(t, int64(2), it.OwnerID)
	assert.Len(t, (it.TravelDestinations), 2)
	assert.Equal(t, int64(10), (it.TravelDestinations)[0].ID)
	assert.Equal(t, "Country1", (it.TravelDestinations)[0].Country)
	assert.Equal(t, "City1", (it.TravelDestinations)[0].City)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_WithoutDestinationsSuccess(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{}

	// Mock main itinerary row
	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", "A test trip", 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	it, err := itinerary.defaultFindById(1, false)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), it.ID)
	assert.Equal(t, "Test Title", it.Title)
	assert.Equal(t, "Test Description", it.Description)
	assert.Equal(t, "A test trip", *it.Notes)
	assert.Equal(t, int64(2), it.OwnerID)
	assert.Nil(t, it.TravelDestinations)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_WithDestinationsSuccessNullNotes(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	now := time.Now()
	itinerary := &Itinerary{}

	// Mock main itinerary row
	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", nil, 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	// Mock travel destinations rows
	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}).
				AddRow(10, "Country1", "City1", 1, now, now.Add(12*time.Hour), time.Now(), time.Now().Add(2*time.Hour)).
				AddRow(11, "Country2", "City2", 1, now.Add(12*time.Hour), now.Add(24*time.Hour), time.Now(), time.Now().Add(2*time.Hour)),
		)

	it, err := itinerary.defaultFindById(1, true)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), it.ID)
	assert.Equal(t, "Test Title", it.Title)
	assert.Equal(t, "Test Description", it.Description)
	assert.Nil(t, it.Notes)
	assert.Equal(t, int64(2), it.OwnerID)
	assert.Len(t, (it.TravelDestinations), 2)
	assert.Equal(t, int64(10), (it.TravelDestinations)[0].ID)
	assert.Equal(t, "Country1", (it.TravelDestinations)[0].Country)
	assert.Equal(t, "City1", (it.TravelDestinations)[0].City)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_RowScanError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	_, err = itinerary.defaultFindById(1, true)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindById_DestinationsQueryError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", nil, 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	_, err = itinerary.defaultFindById(1, true)
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
	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE id = \\?").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
				AddRow(1, "Test Title", "Test Description", nil, 2, time.Now(), time.Now().Add(2*time.Hour)),
		)

	// Return a row with a wrong type to force scan error
	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}).
				AddRow(10, "Country1", "City1", 1, now, now.Add(12*time.Hour), time.Now(), time.Now().Add(2*time.Hour)).
				AddRow("not-an-int", "Country2", "City2", 1, now.Add(12*time.Hour), now.Add(24*time.Hour), time.Now(), time.Now().Add(2*time.Hour)),
		)

	_, err = itinerary.defaultFindById(1, true)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindLightweightById_Success(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, owner_id FROM itineraries WHERE id = \\?").
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).
			AddRow(42, 99))

	it, err := itinerary.defaultFindLightweightById(42)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), it.ID)
	assert.Equal(t, int64(99), it.OwnerID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindLightweightById_NoRows(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, owner_id FROM itineraries WHERE id = \\?").
		WithArgs(100).
		WillReturnError(sql.ErrNoRows)

	_, err = itinerary.defaultFindLightweightById(100)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryDefaultFindLightweightById_ScanError(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()
	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, owner_id FROM itineraries WHERE id = \\?").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"id", "owner_id"}).
			AddRow("not-an-int", 99))

	_, err = itinerary.defaultFindLightweightById(123)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_Success(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", "A test trip", 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}))

	// Act
	itineraries, err := itinerary.defaultFindByOwnerId(1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, itineraries, 1)
	assert.Equal(t, int64(1), (itineraries)[0].ID)
	assert.Equal(t, "Test Title", (itineraries)[0].Title)
	assert.Equal(t, "Test Description", (itineraries)[0].Description)
	assert.Equal(t, "A test trip", *(itineraries)[0].Notes)
	assert.NotNil(t, *(itineraries)[0].CreationDate)
	assert.NotNil(t, *(itineraries)[0].UpdateDate)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_SuccessNullNotes(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", nil, 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date", "creation_date", "update_date"}))

	// Act
	itineraries, err := itinerary.defaultFindByOwnerId(1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, itineraries, 1)
	assert.Equal(t, int64(1), (itineraries)[0].ID)
	assert.Equal(t, "Test Title", (itineraries)[0].Title)
	assert.Equal(t, "Test Description", (itineraries)[0].Description)
	assert.Nil(t, (itineraries)[0].Notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestItineraryFindByOwnerId_NoRows(t *testing.T) {
	// Arrange
	dbMock, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer dbMock.Close()

	db.DB = dbMock

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	// Act
	_, err = itinerary.defaultFindByOwnerId(1)

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

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(1).
		WillReturnError(assert.AnError)

	// Act
	_, err = itinerary.defaultFindByOwnerId(1)

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

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", nil, 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnError(sql.ErrConnDone)

	// Act
	_, err = itinerary.defaultFindByOwnerId(1)

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

	itinerary := &Itinerary{}

	mock.ExpectQuery("SELECT id, title, description, notes, owner_id, creation_date, update_date FROM itineraries WHERE owner_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "notes", "owner_id", "creation_date", "update_date"}).
			AddRow(1, "Test Title", "Test Description", nil, 1, time.Now(), time.Now().Add(2*time.Hour)))

	mock.ExpectQuery("SELECT id, country, city, itinerary_id, arrival_date, departure_date, creation_date, update_date FROM itinerary_travel_destinations WHERE itinerary_id = \\? ORDER BY arrival_date ASC").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "country", "city", "itinerary_id", "arrival_date", "departure_date"}).
			AddRow("not-an-int", "Country1", "City1", 1, time.Now(), time.Now().Add(12*time.Hour)))

	// Act
	_, err = itinerary.defaultFindByOwnerId(1)

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
		TravelDestinations: []*ItineraryTravelDestination{
			NewItineraryTravelDestination("Country 1", "City 1", time.Now(), time.Now().Add(24*time.Hour)),
			NewItineraryTravelDestination("Country 2", "City 2", time.Now(), time.Now().Add(24*time.Hour)),
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", "A test trip", int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for _, destination := range itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	err = itinerary.defaultCreate()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, itinerary.Title, "Test Title")
	assert.Equal(t, itinerary.Description, "Test Description")
	assert.Equal(t, &testNotes, itinerary.Notes)
	assert.Equal(t, int64(1), itinerary.OwnerID)
	assert.Len(t, itinerary.TravelDestinations, 2)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[0].ItineraryID)
	assert.Equal(t, "Country 1", itinerary.TravelDestinations[0].Country)
	assert.Equal(t, "City 1", itinerary.TravelDestinations[0].City)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[1].ItineraryID)
	assert.Equal(t, "Country 2", itinerary.TravelDestinations[1].Country)
	assert.Equal(t, "City 2", itinerary.TravelDestinations[1].City)
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
		TravelDestinations: []*ItineraryTravelDestination{
			NewItineraryTravelDestination("Country 1", "City 1", time.Now(), time.Now().Add(24*time.Hour)),
			NewItineraryTravelDestination("Country 2", "City 2", time.Now(), time.Now().Add(24*time.Hour)),
		},
	}

	mock.ExpectBegin()

	mock.ExpectPrepare("INSERT INTO itineraries").
		ExpectExec().
		WithArgs("Test Title", "Test Description", sqlmock.AnyArg(), int64(1), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	for _, destination := range itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	err = itinerary.defaultCreate()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, itinerary.Title, "Test Title")
	assert.Equal(t, itinerary.Description, "Test Description")
	assert.Nil(t, itinerary.Notes)
	assert.Equal(t, int64(1), itinerary.OwnerID)
	assert.Len(t, itinerary.TravelDestinations, 2)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[0].ItineraryID)
	assert.Equal(t, "Country 1", itinerary.TravelDestinations[0].Country)
	assert.Equal(t, "City 1", itinerary.TravelDestinations[0].City)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[1].ItineraryID)
	assert.Equal(t, "Country 2", itinerary.TravelDestinations[1].Country)
	assert.Equal(t, "City 2", itinerary.TravelDestinations[1].City)
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
		TravelDestinations: []*ItineraryTravelDestination{
			NewItineraryTravelDestination("Country 1", "City 1", time.Now(), time.Now().Add(24*time.Hour)),
			NewItineraryTravelDestination("Country 2", "City 2", time.Now(), time.Now().Add(24*time.Hour)),
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
		OwnerID:     1,
		Notes:       &testNotes,
		TravelDestinations: []*ItineraryTravelDestination{
			NewItineraryTravelDestination("Country 1", "City 1", time.Now(), time.Now().Add(24*time.Hour)),
			NewItineraryTravelDestination("Country 2", "City 2", time.Now(), time.Now().Add(24*time.Hour)),
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

	for _, destination := range itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, itinerary.Title, "Updated Title")
	assert.Equal(t, itinerary.Description, "Updated Description")
	assert.Equal(t, &testNotes, itinerary.Notes)
	assert.Equal(t, int64(1), itinerary.OwnerID)
	assert.Len(t, itinerary.TravelDestinations, 2)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[0].ItineraryID)
	assert.Equal(t, "Country 1", itinerary.TravelDestinations[0].Country)
	assert.Equal(t, "City 1", itinerary.TravelDestinations[0].City)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[1].ItineraryID)
	assert.Equal(t, "Country 2", itinerary.TravelDestinations[1].Country)
	assert.Equal(t, "City 2", itinerary.TravelDestinations[1].City)
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
		OwnerID:     1,
		Notes:       nil,
		TravelDestinations: []*ItineraryTravelDestination{
			NewItineraryTravelDestination("Country 1", "City 1", time.Now(), time.Now().Add(24*time.Hour)),
			NewItineraryTravelDestination("Country 2", "City 2", time.Now(), time.Now().Add(24*time.Hour)),
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

	for _, destination := range itinerary.TravelDestinations {
		mock.ExpectPrepare(`INSERT INTO itinerary_travel_destinations`).ExpectExec().
			WithArgs(destination.Country, destination.City, int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	mock.ExpectCommit()

	// Act
	err = itinerary.defaultUpdate()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(1), itinerary.ID)
	assert.Equal(t, itinerary.Title, "Updated Title")
	assert.Equal(t, itinerary.Description, "Updated Description")
	assert.Nil(t, itinerary.Notes)
	assert.Equal(t, int64(1), itinerary.OwnerID)
	assert.Len(t, itinerary.TravelDestinations, 2)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[0].ItineraryID)
	assert.Equal(t, "Country 1", itinerary.TravelDestinations[0].Country)
	assert.Equal(t, "City 1", itinerary.TravelDestinations[0].City)
	assert.Equal(t, itinerary.ID, itinerary.TravelDestinations[1].ItineraryID)
	assert.Equal(t, "Country 2", itinerary.TravelDestinations[1].Country)
	assert.Equal(t, "City 2", itinerary.TravelDestinations[1].City)
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
		TravelDestinations: []*ItineraryTravelDestination{
			NewItineraryTravelDestination("Country 1", "City 1", time.Now(), time.Now().Add(24*time.Hour)),
			NewItineraryTravelDestination("Country 2", "City 2", time.Now(), time.Now().Add(24*time.Hour)),
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
	originalInitItineraryFileJob := InitItineraryFileJob
	defer func() { InitItineraryFileJob = originalInitItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	InitItineraryFileJob = func() *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
		return nil
	}

	// Mock DeleteByItineraryIdTx
	originalInitItineraryTravelDestination := InitItineraryTravelDestination
	defer func() { InitItineraryTravelDestination = originalInitItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	InitItineraryTravelDestination = func() *ItineraryTravelDestination {
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
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

	originalInitItineraryFileJob := InitItineraryFileJob
	defer func() { InitItineraryFileJob = originalInitItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	InitItineraryFileJob = func() *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
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

	// Mock SoftDeleteJobsByItineraryIdTx
	originalInitItineraryFileJob := InitItineraryFileJob
	defer func() { InitItineraryFileJob = originalInitItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	InitItineraryFileJob = func() *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
		return nil
	}

	// Mock DeleteByItineraryIdTx
	originalInitItineraryTravelDestination := InitItineraryTravelDestination
	defer func() { InitItineraryTravelDestination = originalInitItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	InitItineraryTravelDestination = func() *ItineraryTravelDestination {
		return mockDest
	}

	mockDest.DeleteByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
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

	// Mock SoftDeleteJobsByItineraryIdTx
	originalInitItineraryFileJob := InitItineraryFileJob
	defer func() { InitItineraryFileJob = originalInitItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	InitItineraryFileJob = func() *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
		return nil
	}

	// Mock DeleteByItineraryIdTx
	originalInitItineraryTravelDestination := InitItineraryTravelDestination
	defer func() { InitItineraryTravelDestination = originalInitItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	InitItineraryTravelDestination = func() *ItineraryTravelDestination {
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
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

	// Mock SoftDeleteJobsByItineraryIdTx
	originalInitItineraryFileJob := InitItineraryFileJob
	defer func() { InitItineraryFileJob = originalInitItineraryFileJob }()
	mockJob := &ItineraryFileJob{}
	InitItineraryFileJob = func() *ItineraryFileJob {
		return mockJob
	}
	mockJob.SoftDeleteJobsByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
		return nil
	}

	// Mock DeleteByItineraryIdTx
	originalInitItineraryTravelDestination := InitItineraryTravelDestination
	defer func() { InitItineraryTravelDestination = originalInitItineraryTravelDestination }()
	mockDest := &ItineraryTravelDestination{}
	InitItineraryTravelDestination = func() *ItineraryTravelDestination {
		return mockDest
	}
	mockDest.DeleteByItineraryIdTx = func(itineraryId int64, tx *sql.Tx) error {
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
