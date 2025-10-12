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

	userID := req.GetUserID()
	route, err := h.service.GetRoute(ctx, startCoords, endCoords)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get route: %v", err)
	}

	estimatedFares := h.service.EstimatePackagePriceWithRoute(route)
	fares, err := h.service.GenerateTripFares(ctx, route, estimatedFares, userID)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to get generate trip fares: %v", err)
	}

	return &pb.PreviewTripResponse{
		Route:     route.ToProto(),
		RideFares: domain.ToRideFareProto(fares),
	}, nil
}
func (h *gRPCHandler) CreateTrip(ctx context.Context, req *pb.CreateTripRequest) (*pb.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	fare, err := h.service.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed validating fare: %v", err)
	}

	trip, err := h.service.CreateTrip(ctx, fare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed creating trip: %v", err)
	}

	tripProto := &pb.CreateTripResponse{
		TripID: trip.ID.Hex(),
		Trip: &pb.Trip{
			Id:           trip.ID.Hex(),
			UserID:       trip.UserID,
			Status:       trip.Status,
			SelectedFare: trip.RideFare.ToProto(),
		},
	}
	return tripProto, nil
}
