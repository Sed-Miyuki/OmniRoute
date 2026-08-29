package main

import (
	"context"

	pb "github.com/Sed-Miyuki/OmniRoute/shared/proto/driver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedDriverServiceServer
	Service *Service
}

func NewGrpcHandler(s *grpc.Server, service *Service) {
	handler:=&grpcHandler{
		Service: service,
	}
	pb.RegisterDriverServiceServer(s,handler)
}

func (h *grpcHandler) RegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method RegisterDriver not implemented")
}
func (h *grpcHandler) UnRegisterDriver(ctx context.Context, req *pb.RegisterDriverRequest) (*pb.RegisterDriverResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UnRegisterDriver not implemented")
}
