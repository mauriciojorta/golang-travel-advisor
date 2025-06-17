package services

import (
	"errors"
	"fmt"
	"time"

	"example.com/travel-advisor/models"
)

type ItineraryServiceInterface interface {
	ValidateOwnership(itineraryId int64, currentUserId int64) (bool, error)
	ExistById(id int64) (bool, error)
	FindById(id int64, includeDestinations bool) (*models.Itinerary, error)
	FindByOwnerId(ownerId int64) (*[]models.Itinerary, error)
	Create(itinerary *models.Itinerary) error
	Update(itinerary *models.Itinerary) error
	Delete(id int64) error
	ValidateItineraryDestinationsDates(destinations *[]models.ItineraryTravelDestination) error
}

type ItineraryService struct {
}

// singleton instance
var itineraryServiceInstance = &ItineraryService{}

// GetItineraryService returns the singleton instance of ItineraryService
var GetItineraryService = func() ItineraryServiceInterface {
	return itineraryServiceInstance
}

func (is *ItineraryService) ValidateOwnership(itineraryId int64, currentUserId int64) (bool, error) {
	if itineraryId <= 0 {
		return false, errors.New("invalid itinerary ID")
	}
	itinerary := models.NewItinerary("", "", nil, nil) // Create a new Itinerary instance
	itinerary.ID = itineraryId                         // Set the ID for the itinerary instance
	exists, err := itinerary.ValidateOwnership(currentUserId)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (is *ItineraryService) ExistById(itineraryId int64) (bool, error) {
	if itineraryId <= 0 {
		return false, errors.New("invalid itinerary ID")
	}
	itinerary := models.NewItinerary("", "", nil, nil) // Create a new Itinerary instance
	itinerary.ID = itineraryId                         // Set the ID for the itinerary instance
	exists, err := itinerary.ExistById()
	if err != nil {
		return false, err
	}
	return exists, nil
}

// FindById retrieves the itinerary by its ID
func (is *ItineraryService) FindById(id int64, includeDestinations bool) (*models.Itinerary, error) {
	if id <= 0 {
		return nil, errors.New("invalid itinerary ID")
	}
	itinerary := models.NewItinerary("", "", nil, nil) // Create a new Itinerary instance
	itinerary.ID = id                                  // Set the ID for the itinerary instance
	err := itinerary.FindById(includeDestinations)
	if err != nil {
		return nil, err
	}
	return itinerary, nil
}

// FindByOwnerId retrieves itineraries by owner ID
func (is *ItineraryService) FindByOwnerId(ownerId int64) (*[]models.Itinerary, error) {
	if ownerId <= 0 {
		return nil, errors.New("invalid owner ID")
	}
	itinerary := models.NewItinerary("", "", nil, nil) // Create a new Itinerary instance
	itinerary.OwnerID = ownerId                        // Set the owner ID for the itinerary instance
	return itinerary.FindByOwnerId()
}

// Create creates a new itinerary
func (is *ItineraryService) Create(itinerary *models.Itinerary) error {
	if itinerary == nil {
		return errors.New("itinerary instance is nil")
	}
	return itinerary.Create()
}

// Update updates the itinerary
func (is *ItineraryService) Update(itinerary *models.Itinerary) error {
	if itinerary == nil {
		return errors.New("itinerary instance is nil")
	}
	return itinerary.Update()
}

// Delete deletes the itinerary
func (is *ItineraryService) Delete(id int64) error {
	if id <= 0 {
		return errors.New("invalid itinerary ID")
	}
	itinerary := models.NewItinerary("", "", nil, nil) // Create a new Itinerary instance
	itinerary.ID = id                                  // Set the ID for the itinerary instance
	return itinerary.Delete()
}
func (is *ItineraryService) ValidateItineraryDestinationsDates(destinations *[]models.ItineraryTravelDestination) error {
	if len(*destinations) == 0 {
		return fmt.Errorf("At least one destination is required")
	}

	if len(*destinations) > 20 {
		return fmt.Errorf("The itinerary cannot have more than 20 destinations")
	}

	// Find oldest arrival and latest departure
	var oldestArrival, latestDeparture time.Time
	for i, dest := range *destinations {
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
		return fmt.Errorf("The itinerary cannot span more than 30 days")
	}

	return nil
}
