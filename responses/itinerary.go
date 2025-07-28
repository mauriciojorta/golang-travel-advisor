package responses

import "example.com/travel-advisor/models"

type CreateItineraryResponse struct {
	Message     string `json:"message" example:"Itinerary created."`
	ItineraryID int64  `json:"itineraryId" example:"123"`
}

type GetItineraryResponse struct {
	Itinerary *models.Itinerary `json:"itinerary"` // Example JSON representation
}

type GetItinerariesResponse struct {
	Itineraries []*models.Itinerary `json:"itineraries"` // Example JSON representation
}

type StartItineraryJobResponse struct {
	Message string `json:"message" example:"Job started successfully."`
	JobId   int64  `json:"jobId" example:"123"`
}

type StopItineraryJobResponse struct {
	Message string `json:"message" example:"Itinerary job stopped."`
}

type GetItineraryJobResponse struct {
	Job *models.ItineraryFileJob `json:"job"` // Example JSON representation
}

type GetItineraryJobsResponse struct {
	Jobs []*models.ItineraryFileJob `json:"job"`
}

type UpdateItineraryResponse struct {
	Message string `json:"message" example:"Itinerary updated."`
}

type DeleteItineraryResponse struct {
	Message string `json:"message" example:"Itinerary deleted."`
}

type DeleteItineraryJobResponse struct {
	Message string `json:"message" example:"Itinerary job deleted."`
}
