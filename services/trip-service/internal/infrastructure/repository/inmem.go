package repository

import (
	"context"
	"fmt"

	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/domain"
	pbd "github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"
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

func (r *inmemRepository) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error){
	trip,ok:=r.trips[id]
	if !ok{
		return nil,nil
	}
	return trip,nil
}

func (r *inmemRepository) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error{
	trip,ok:=r.trips[tripID]
	if !ok{
		return fmt.Errorf("trip not found with id: %s",tripID)
	}

	trip.Status=status
	if driver!=nil{
		trip.Driver=&pb.TripDriver{
			Id: driver.Id,
			Name: driver.Name,
			ProfilePicture: driver.ProfilePicture,
			CarPlate: driver.CarPlate,
		}
	}
	return nil
}