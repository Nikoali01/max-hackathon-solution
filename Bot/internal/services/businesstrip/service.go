package businesstrip

import (
	"context"
	"fmt"
	"time"
)

type Trip struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Destination string    `json:"destination"`
	Purpose     string    `json:"purpose"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Status      string    `json:"status"` // "pending", "approved", "rejected", "completed"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Service interface {
	CreateTrip(ctx context.Context, userID, destination, purpose string, startDate, endDate time.Time) (*Trip, error)
	GetUserTrips(ctx context.Context, userID string) ([]Trip, error)
	GetTrip(ctx context.Context, tripID string) (*Trip, error)
}

type mockService struct {
	trips map[string]*Trip
}

func NewMock() Service {
	return &mockService{
		trips: make(map[string]*Trip),
	}
}

func (s *mockService) CreateTrip(ctx context.Context, userID, destination, purpose string, startDate, endDate time.Time) (*Trip, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	now := time.Now()
	trip := &Trip{
		ID:          fmt.Sprintf("TRIP-%d", now.Unix()),
		UserID:      userID,
		Destination: destination,
		Purpose:     purpose,
		StartDate:   startDate,
		EndDate:     endDate,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.trips[trip.ID] = trip
	return trip, nil
}

func (s *mockService) GetUserTrips(ctx context.Context, userID string) ([]Trip, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var result []Trip
	for _, trip := range s.trips {
		if trip.UserID == userID {
			result = append(result, *trip)
		}
	}
	return result, nil
}

func (s *mockService) GetTrip(ctx context.Context, tripID string) (*Trip, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	trip, exists := s.trips[tripID]
	if !exists {
		return nil, nil
	}
	return trip, nil
}

