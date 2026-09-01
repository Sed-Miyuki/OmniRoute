package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/domain"
	tripTypes "github.com/Sed-Miyuki/OmniRoute/services/trip-service/pkg/types"
	pbd "github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
	"github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"
	"github.com/Sed-Miyuki/OmniRoute/shared/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewService(repo domain.TripRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	t := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver: &trip.TripDriver{},
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude, pickup.Latitude,
		destination.Longitude, destination.Latitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route from OSRM API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response: %v", err)
	}

	var routeResp tripTypes.OsrmApiResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &routeResp, nil
}

func (s *service) GetAndValidateFare(ctx context.Context,fareID,userID string) (*domain.RideFareModel,error){
	fare,err:=s.repo.GetFareByID(ctx,fareID)
	if err!=nil{
		return nil,fmt.Errorf("failed to get trip fare: %v",err)
	}

	if fare==nil{
		return nil,fmt.Errorf("fare does not exist")
	}
	if userID!=fare.UserID{
		return nil,fmt.Errorf("fare does not belong to the user %s---------%s",userID,fare.UserID)
	}
	return fare,nil
}

func (s *service) EstimatePackagesPriceWithRoute(route *tripTypes.OsrmApiResponse) []*domain.RideFareModel{
	baseFares:=getBaseFares()
	estimatedFares:=make([]*domain.RideFareModel,len(baseFares))

	for i,fare:=range baseFares{
		estimatedFares[i]=estimateFareRoute(fare,route) 
	}
	return estimatedFares
}
	
func (s *service) GenerateTripFares(ctx context.Context,rideFares []*domain.RideFareModel,userid string,route *tripTypes.OsrmApiResponse) ([]*domain.RideFareModel,error){
	fares:=make([]*domain.RideFareModel,len(rideFares))
	for i,f:=range rideFares{
		id:=primitive.NewObjectID()
		fare:=&domain.RideFareModel{
			UserID: userid,
			ID: id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug: f.PackageSlug,
			Route: route,
		}
		if err:=s.repo.SaveRideFare(ctx,fare);err!=nil{
			return nil,fmt.Errorf("failed to save trip fare: %w",err)
		}
		fares[i]=fare
	} 
	return fares,nil
}

func estimateFareRoute(fare *domain.RideFareModel,route *tripTypes.OsrmApiResponse) *domain.RideFareModel{
	cfg:=tripTypes.DefaultPricingConfig()
	carPackagePrice:=fare.TotalPriceInCents
	distanceKm:=route.Routes[0].Distance
	distanceInMinutes:=route.Routes[0].Duration

	distanceFare:=distanceKm*cfg.PricePerUnitOfDistance
	timeFare:=distanceInMinutes*cfg.PricingPerMinute
	totalPrice:=carPackagePrice+distanceFare+timeFare

	return &domain.RideFareModel{
		TotalPriceInCents: totalPrice,
		PackageSlug: fare.PackageSlug,
	} 
}

func getBaseFares() []*domain.RideFareModel{
	return []*domain.RideFareModel{
		{
			PackageSlug: "suv",
			TotalPriceInCents: 500,
		},
		{
			PackageSlug: "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug: "van",
			TotalPriceInCents: 300,
		},
		{
			PackageSlug: "luxury",
			TotalPriceInCents: 1100,
		},
	}
}

func (s *service) GetTripByID(ctx context.Context, id string) (*domain.TripModel, error){
	return s.repo.GetTripByID(ctx,id)
}
func (s *service) UpdateTrip(ctx context.Context, tripID string, status string, driver *pbd.Driver) error{
	return s.repo.UpdateTrip(ctx,tripID,status,driver)
}