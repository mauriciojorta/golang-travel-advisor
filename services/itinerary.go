package services

import (
	"errors"
	"fmt"
	"time"

	"example.com/travel-advisor/models"
	log "github.com/sirupsen/logrus"
)

type ItineraryServiceInterface interface {
	FindById(id int64, includeDestinations bool) (*models.Itinerary, error)
	FindLightweightById(id int64) (*models.Itinerary, error)
	FindByOwnerId(ownerId int64) ([]*models.Itinerary, error)
	Create(itinerary *models.Itinerary) error
	Update(itinerary *models.Itinerary) error
	Delete(id int64) error
	ValidateItineraryDestinationsDates(destinations []*models.ItineraryTravelDestination) error
}

type ItineraryService struct {
}

// singleton instance
var itineraryServiceInstance = &ItineraryService{}

// GetItineraryService returns the singleton instance of ItineraryService
var GetItineraryService = func() ItineraryServiceInterface {
	return itineraryServiceInstance
}

// FindById retrieves the itinerary by its ID
func (is *ItineraryService) FindById(id int64, includeDestinations bool) (*models.Itinerary, error) {
	if id <= 0 {
		log.Error("Invalid itinerary ID provided")
		return nil, errors.New("invalid itinerary ID")
	}
	itinerary := models.InitItinerary() // Create a new Itinerary instance
	return itinerary.FindById(id, includeDestinations)
}

// FindById retrieves a "lightweight" itinerary instance (just the ID and owner ID) by its ID
func (is *ItineraryService) FindLightweightById(id int64) (*models.Itinerary, error) {
	if id <= 0 {
		log.Error("Invalid itinerary ID provided")
		return nil, errors.New("invalid itinerary ID")
	}
	itinerary := models.InitItinerary() // Create a new Itinerary instance
	return itinerary.FindLightweightById(id)
}

// FindByOwnerId retrieves itineraries by owner ID
func (is *ItineraryService) FindByOwnerId(ownerId int64) ([]*models.Itinerary, error) {
	if ownerId <= 0 {
		log.Error("Invalid owner ID provided")
		return nil, errors.New("invalid owner ID")
	}
	itinerary := models.InitItinerary() // Create a new Itinerary instance
	return itinerary.FindByOwnerId(ownerId)
}

// Create creates a new itinerary
func (is *ItineraryService) Create(itinerary *models.Itinerary) error {
	if itinerary == nil {
		log.Error("Itinerary instance is nil")
		return errors.New("itinerary instance is nil")
	}
	return itinerary.Create()
}

// Update updates the itinerary
func (is *ItineraryService) Update(itinerary *models.Itinerary) error {
	if itinerary == nil {
		log.Error("Itinerary instance is nil")
		return errors.New("itinerary instance is nil")
	}
	return itinerary.Update()
}

// Delete deletes the itinerary
func (is *ItineraryService) Delete(id int64) error {
	if id <= 0 {
		log.Error("Invalid itinerary ID provided")
		return errors.New("invalid itinerary ID")
	}
	itinerary := models.InitItinerary() // Create a new Itinerary instance
	itinerary.ID = id                   // Set the ID for the itinerary instance
	return itinerary.Delete()
}
func (is *ItineraryService) ValidateItineraryDestinationsDates(destinations []*models.ItineraryTravelDestination) error {
	if len(destinations) == 0 {
		log.Error("At least one destination is required")
		return fmt.Errorf("at least one destination is required")
	}

	if len(destinations) > 20 {
		log.Error("The itinerary cannot have more than 20 destinations")
		return fmt.Errorf("the itinerary cannot have more than 20 destinations")
	}

	// Find oldest arrival and latest departure
	var oldestArrival, latestDeparture time.Time
	for i, dest := range destinations {
		if i == 0 {
			oldestArrival = dest.ArrivalDate
			latestDeparture = dest.DepartureDate
		} else {
			if dest.ArrivalDate.Before(oldestArrival) {
				oldestArrival = dest.ArrivalDate
			}
			if dest.DepartureDate.After(latestDeparture) {
				latestDeparture = dest.DepartureDate
			}
		}
	}

	if latestDeparture.Sub(oldestArrival).Hours()/24 > 30 {
		log.Error("The itinerary cannot span more than 30 days")
		return fmt.Errorf("the itinerary cannot span more than 30 days")
	}

	return nil
}
