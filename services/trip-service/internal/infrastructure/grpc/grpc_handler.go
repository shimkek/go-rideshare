package grpc

import (
	"context"
	"log"
	"ride-sharing/services/trip-service/internal/domain"
	pb "ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pb.UnimplementedTripServiceServer
	service domain.TripService
}

func NewGRPCHandler(server *grpc.Server, service domain.TripService) *gRPCHandler {
	handler := &gRPCHandler{
		service: service,
	}
	pb.RegisterTripServiceServer(server, handler)
	return handler
}
func (h *gRPCHandler) PreviewTrip(ctx context.Context, req *pb.PreviewTripRequest) (*pb.PreviewTripResponse, error) {
	reqStartLocation := req.GetStartLocation()
	startCoords := &types.Coordinate{
		Latitude:  reqStartLocation.Latitude,
		Longitude: reqStartLocation.Longitude,
	}

	reqEndLocation := req.GetEndLocation()
	endCoords := &types.Coordinate{
		Latitude:  reqEndLocation.Latitude,
		Longitude: reqEndLocation.Longitude,
	}
	res, err := h.service.GetRoute(ctx, startCoords, endCoords)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}
	return &pb.PreviewTripResponse{
		TripID:    "123",
		Route:     res.ToProto(),
		RideFares: []*pb.RideFare{},
	}, nil
}
func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	// trip, err := h.service.CreateTrip(ctx, fare)
	return nil, nil
}
