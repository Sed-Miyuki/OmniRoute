package domain

import (
	"context"

	tripTypes "github.com/Sed-Miyuki/OmniRoute/services/trip-service/pkg/types"
	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"
	"github.com/Sed-Miyuki/OmniRoute/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TripModel struct {
	ID       primitive.ObjectID
	UserID   string
	Status   string
	RideFare *RideFareModel
	Driver *pb.TripDriver
}

func (t *TripModel) ToProto() *pb.Trip{
	return &pb.Trip{
		Id: t.ID.Hex(),
		UserID: t.UserID,
		SelectedRideFare: t.RideFare.ToProto(),
		Status: t.Status,
		Driver: t.Driver,
		Route: t.RideFare.Route.ToProto(),
	}
}

type TripRepository interface {
	CreateTrip(ctx context.Context, trip *TripModel) (*TripModel, error)
	SaveRideFare(ctx context.Context,f *RideFareModel) error
	GetFareByID(ctx context.Context,id string) (*RideFareModel,error)
}

type TripService interface {
	CreateTrip(ctx context.Context, fare *RideFareModel) (*TripModel, error)
	GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*RideFareModel
	GenerateTripFares(ctx context.Context,fares []*RideFareModel,userid string,route *tripTypes.OsrmApiResponse) ([]*RideFareModel,error)

	GetAndValidateFare(ctx context.Context,fareID,userID string) (*RideFareModel,error)
}