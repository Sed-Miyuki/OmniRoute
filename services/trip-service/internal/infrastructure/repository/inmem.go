package repository

import (
	"context"
	"fmt"

	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/domain"
)

type inmemRepository struct {
	trips     map[string]*domain.TripModel
	rideFares map[string]*domain.RideFareModel
}

func NewInmemRepository() *inmemRepository {
	return &inmemRepository{
		trips:     make(map[string]*domain.TripModel),
		rideFares: make(map[string]*domain.RideFareModel),
	}
}

func (r *inmemRepository) CreateTrip(ctx context.Context, trip *domain.TripModel) (*domain.TripModel, error) {
	r.trips[trip.ID.Hex()] = trip
	return trip, nil
}

func (r *inmemRepository) SaveRideFare(ctx context.Context,f *domain.RideFareModel) error{
	r.rideFares[f.ID.Hex()]=f
	return nil
}

func (r *inmemRepository) GetFareByID(ctx context.Context,id string) (*domain.RideFareModel,error){
	fare,exist:=r.rideFares[id]
	if !exist{
		return nil,fmt.Errorf("fare does not exist with ID: %v",id)
	}
	return fare,nil 
}