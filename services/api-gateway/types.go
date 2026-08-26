package main

import (
	"github.com/Sed-Miyuki/OmniRoute/shared/types"
	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/trip"
)

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *previewTripRequest) ToProto() *pb.PreviewTripRequest{
	return &pb.PreviewTripRequest{
		UserID: p.UserID,
		StartingLocation: &pb.Coordinate{
			Latitude: p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		EndLocation: &pb.Coordinate{
			Latitude: p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type startTripRequest struct{
	RideFareID string `json:"rideFareID"`
	UserID string `json:"userID"`
}

func (c *startTripRequest) toProto() *pb.CreateTripRequest {
	return &pb.CreateTripRequest{
		RideFareID: c.RideFareID,
		UserID:     c.UserID,
	}
}