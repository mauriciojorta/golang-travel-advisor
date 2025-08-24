package requests

import "time"

type CreateItineraryRequest struct {
	Title        string             `json:"title" binding:"required,max=128" example:"Trip to Spain"`
	Description  string             `json:"description" binding:"omitempty,max=512" example:"Summer vacation in Spain"`
	Notes        *string            `json:"notes" binding:"omitnil,omitempty,max=512" example:"I want to enjoy the nightlife"`
	Destinations []*DestinationItem `json:"destinations" binding:"required,min=0,max=20,dive"`
}

type UpdateItineraryRequest struct {
	ID           int64              `json:"id" binding:"required" example:"1"`
	Title        string             `json:"title" binding:"required,max=128" example:"Trip to Spain"`
	Description  string             `json:"description" binding:"omitempty,max=256" example:"Summer vacation in Spain"`
	Notes        *string            `json:"notes" binding:"omitnil,omitempty,max=512" example:"I want to enjoy the nightlife"`
	Destinations []*DestinationItem `json:"destinations" binding:"required,min=0,max=20,dive"`
}

type DestinationItem struct {
	Country       string    `json:"country" binding:"required,max=128" example:"Spain"`
	City          string    `json:"city" binding:"required,max=128" example:"Madrid"`
	ArrivalDate   time.Time `json:"arrivalDate" binding:"required" example:"2024-07-01T00:00:00Z"`
	DepartureDate time.Time `json:"departureDate" binding:"required" example:"2024-07-05T00:00:00Z"`
}
