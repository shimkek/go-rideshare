package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/proto/trip"
	"ride-sharing/shared/types"

	tripTypes "ride-sharing/services/trip-service/pkg/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.TripRepository
}

func NewService(repo domain.TripRepository) *service {
	return &service{repo: repo}
}

func (s *service) CreateTrip(ctx context.Context, fare *domain.RideFareModel) (*domain.TripModel, error) {
	trip := &domain.TripModel{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver:   &trip.Driver{},
	}
	return s.repo.CreateTrip(ctx, trip)
}

func (s *service) GetRoute(ctx context.Context, pickup, destination *types.Coordinate) (*tripTypes.OsrmApiResponse, error) {
	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		pickup.Longitude, pickup.Latitude, destination.Longitude, destination.Latitude)

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route from OSRM API: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response: %v", err)
	}

	var routeResp tripTypes.OsrmApiResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the response: %v", err)
	}

	return &routeResp, nil
}

func (s *service) EstimatePackagePriceWithRoute(route *tripTypes.OsrmApiResponse) []*domain.RideFareModel {
	baseFares := getBaseFares()
	pricingCfc := tripTypes.DefaultPricingConfig()

	estimatedFares := make([]*domain.RideFareModel, len(baseFares))
	for i, f := range baseFares {
		distanceFare := route.Routes[0].Distance * pricingCfc.PricePerUnitOfDistanse
		timeFare := route.Routes[0].Duration * pricingCfc.PricePerMinute
		fare := &domain.RideFareModel{
			PackageSlug: f.PackageSlug,
			// package base price + distance fare + time fare
			TotalPriceInCents: f.TotalPriceInCents + distanceFare + timeFare,
		}

		estimatedFares[i] = fare
	}

	return estimatedFares
}
func (s *service) GenerateTripFares(ctx context.Context, route *tripTypes.OsrmApiResponse, fares []*domain.RideFareModel, userID string) ([]*domain.RideFareModel, error) {
	commitedFares := make([]*domain.RideFareModel, len(fares))
	for i, f := range fares {
		f.UserID = userID
		f.ID = primitive.NewObjectID()
		f.Route = route

		if err := s.repo.SaveRideFare(ctx, f); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %s", err)
		}
		commitedFares[i] = f
	}

	return commitedFares, nil
}

func getBaseFares() []*domain.RideFareModel {
	return []*domain.RideFareModel{
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 300,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}

func (s *service) GetAndValidateFare(ctx context.Context, fareID, userID string) (*domain.RideFareModel, error) {
	fare, err := s.repo.GetRideFareByID(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip fare: %w", err)
	}

	if userID != fare.UserID {
		return nil, fmt.Errorf("the user is not the owner of the fare")
	}
	return fare, nil
}
