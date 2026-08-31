package grpc

import (
	"context"
	"log"

	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/domain"
	"github.com/Sed-Miyuki/OmniRoute/services/trip-service/internal/infrastructure/events"
	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"
	"github.com/Sed-Miyuki/OmniRoute/shared/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct{
	pb.UnimplementedTripServiceServer
	service domain.TripService
	publisher *events.TripEventPublisher
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService, publisher *events.TripEventPublisher) *gRPCHandler{
	handler:=&gRPCHandler{
		service: service,
		publisher: publisher,
	}

	pb.RegisterTripServiceServer(server, handler)
	return handler
}

func (h *gRPCHandler) PreviewTrip(ctx context.Context,req *pb.PreviewTripRequest) (*pb.PreviewTripResponse,error){
	pickup:=req.GetStartingLocation()
	destination:=req.GetEndLocation()

	pickupCoord:=&types.Coordinate{
		Latitude: pickup.Latitude,
		Longitude: pickup.Longitude,
	}
	destinationCoord:=&types.Coordinate{
		Latitude: destination.Latitude,
		Longitude: destination.Longitude,
	}

	userId:=req.GetUserID()
	route,err:=h.service.GetRoute(ctx,pickupCoord,destinationCoord)
	if err!=nil{
		log.Println(err)
		return nil,status.Errorf(codes.Internal, "failed to get route: %v",err)
	}

	estimatedFares:=h.service.EstimatePackagesPriceWithRoute(route)
	fares,err:=h.service.GenerateTripFares(ctx,estimatedFares,userId,route)
	if err != nil {
		return nil, status.Errorf(codes.Internal,"failed to generate the ride fares: %v",err)
	}
	return &pb.PreviewTripResponse{
		Route: route.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	},nil
}

func (h *gRPCHandler) CreateTrip(ctx context.Context,req *pb.CreateTripRequest) (*pb.CreateTripResponse,error){
	fareID:=req.GetRideFareID()
	userID:=req.GetUserID()
	rideFare,err:=h.service.GetAndValidateFare(ctx,fareID,userID)
	if err!=nil{
		return nil,status.Errorf(codes.Internal,"failed to validate the fare: %v",err)
	}
	
	trip,err:=h.service.CreateTrip(ctx,rideFare)
	if err!=nil{
		return nil,status.Errorf(codes.Internal,"failed to create the trip: %v",err)
	}
	if err := h.publisher.PublishTripCreated(ctx,trip); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish the trip created event: %v", err)
	}
	return &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
	},nil
}