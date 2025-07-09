package services

import (
	"errors"
	"testing"
	"time"

	"example.com/travel-advisor/models"
)

// Helper to create a mock itinerary with stubbed methods
func mockItinerary() *models.Itinerary {
	it := models.NewItinerary("", "", nil, nil)
	it.ID = 1
	it.OwnerID = 2
	it.FindById = func(id int64, _ bool) (*models.Itinerary, error) { return &models.Itinerary{}, nil }
	it.FindLightweightById = func(id int64) (*models.Itinerary, error) {
		return &models.Itinerary{}, nil
	}
	it.FindByOwnerId = func(ownerId int64) ([]*models.Itinerary, error) {
		arr := []*models.Itinerary{it}
		return arr, nil
	}
	it.Create = func() error { return nil }
	it.Update = func() error { return nil }
	it.Delete = func() error { return nil }
	return it
}

func TestFindById_Success(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	called := false
	it.FindById = func(id int64, _ bool) (*models.Itinerary, error) { called = true; return &models.Itinerary{}, nil }
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	got, err := svc.FindById(1, true)
	if err != nil || got == nil || !called {
		t.Errorf("expected success, got err=%v, called=%v", err, called)
	}
}

func TestFindById_InvalidID(t *testing.T) {
	svc := &ItineraryService{}
	got, err := svc.FindById(0, true)
	if err == nil || got != nil {
		t.Errorf("expected error for invalid id")
	}
}

func TestFindById_ErrorFromModel(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.FindById = func(id int64, _ bool) (*models.Itinerary, error) { return nil, errors.New("fail") }
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	got, err := svc.FindById(1, true)
	if err == nil || got != nil {
		t.Errorf("expected error from model")
	}
}

func TestFindLightweightById_Success(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	called := false
	it.FindLightweightById = func(id int64) (*models.Itinerary, error) { called = true; return &models.Itinerary{}, nil }
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	got, err := svc.FindLightweightById(1)
	if err != nil || got == nil || !called {
		t.Errorf("expected success, got err=%v, called=%v", err, called)
	}
}

func TestFindLightweightById_InvalidID(t *testing.T) {
	svc := &ItineraryService{}
	got, err := svc.FindLightweightById(0)
	if err == nil || got != nil {
		t.Errorf("expected error for invalid id")
	}
}

func TestFindLightweightById_ErrorFromModel(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.FindLightweightById = func(id int64) (*models.Itinerary, error) { return nil, errors.New("fail") }
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	got, err := svc.FindLightweightById(1)
	if err == nil || got != nil {
		t.Errorf("expected error from model")
	}
}

func TestFindByOwnerId_Success(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()

	called := false
	it.FindByOwnerId = func(ownerId int64) ([]*models.Itinerary, error) {
		called = true
		arr := []*models.Itinerary{it}
		return arr, nil
	}
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	got, err := svc.FindByOwnerId(2)
	if err != nil || got == nil || !called {
		t.Errorf("expected success, got err=%v, called=%v", err, called)
	}
}

func TestFindByOwnerId_InvalidOwnerID(t *testing.T) {
	svc := &ItineraryService{}
	got, err := svc.FindByOwnerId(0)
	if err == nil || got != nil {
		t.Errorf("expected error for invalid owner id")
	}
}

func TestFindByOwnerId_ErrorFromModel(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()

	it.FindByOwnerId = func(ownerId int64) ([]*models.Itinerary, error) { return nil, errors.New("fail") }
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	got, err := svc.FindByOwnerId(2)
	if err == nil || got != nil {
		t.Errorf("expected error from model")
	}
}

func TestCreate_Success(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.Create = func() error { return nil }
	err := svc.Create(it)
	if err != nil {
		t.Errorf("expected success, got err=%v", err)
	}
}

func TestCreate_NilItinerary(t *testing.T) {
	svc := &ItineraryService{}
	err := svc.Create(nil)
	if err == nil {
		t.Errorf("expected error for nil itinerary")
	}
}

func TestCreate_ErrorFromModel(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.Create = func() error { return errors.New("fail") }
	err := svc.Create(it)
	if err == nil {
		t.Errorf("expected error from model")
	}
}

func TestUpdate_Success(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.Update = func() error { return nil }
	err := svc.Update(it)
	if err != nil {
		t.Errorf("expected success, got err=%v", err)
	}
}

func TestUpdate_NilItinerary(t *testing.T) {
	svc := &ItineraryService{}
	err := svc.Update(nil)
	if err == nil {
		t.Errorf("expected error for nil itinerary")
	}
}

func TestUpdate_ErrorFromModel(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.Update = func() error { return errors.New("fail") }
	err := svc.Update(it)
	if err == nil {
		t.Errorf("expected error from model")
	}
}

func TestDelete_Success(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	it.Delete = func() error { return nil }
	err := svc.Delete(1)
	if err != nil {
		t.Errorf("expected success, got err=%v", err)
	}
}

func TestDelete_InvalidItineraryId(t *testing.T) {
	svc := &ItineraryService{}
	err := svc.Delete(-1)
	if err == nil {
		t.Errorf("expected error for nil itinerary")
	}
}

func TestDelete_ErrorFromModel(t *testing.T) {
	svc := &ItineraryService{}
	it := mockItinerary()
	models.InitItinerary = func() *models.Itinerary {
		return it
	}
	it.Delete = func() error { return errors.New("fail") }
	err := svc.Delete(1)
	if err == nil {
		t.Errorf("expected error from model")
	}
}

func TestValidateItineraryDestinationsDates_Empty(t *testing.T) {
	svc := &ItineraryService{}
	dest := []*models.ItineraryTravelDestination{}
	err := svc.ValidateItineraryDestinationsDates(dest)
	if err == nil {
		t.Errorf("expected error for empty destinations")
	}
}

func TestValidateItineraryDestinationsDates_TooMany(t *testing.T) {
	svc := &ItineraryService{}
	dest := make([]*models.ItineraryTravelDestination, 21)
	err := svc.ValidateItineraryDestinationsDates(dest)
	if err == nil {
		t.Errorf("expected error for too many destinations")
	}
}

func TestValidateItineraryDestinationsDates_TooLongSpan(t *testing.T) {
	svc := &ItineraryService{}
	now := time.Now()
	dest := []*models.ItineraryTravelDestination{
		{ArrivalDate: now, DepartureDate: now.Add(31 * 24 * time.Hour)},
	}
	err := svc.ValidateItineraryDestinationsDates(dest)
	if err == nil {
		t.Errorf("expected error for span > 30 days")
	}
}

func TestValidateItineraryDestinationsDates_Valid(t *testing.T) {
	svc := &ItineraryService{}
	now := time.Now()
	dest := []*models.ItineraryTravelDestination{
		{ArrivalDate: now, DepartureDate: now.Add(10 * 24 * time.Hour)},
		{ArrivalDate: now.Add(2 * 24 * time.Hour), DepartureDate: now.Add(12 * 24 * time.Hour)},
	}
	err := svc.ValidateItineraryDestinationsDates(dest)
	if err != nil {
		t.Errorf("expected valid, got err=%v", err)
	}
}
