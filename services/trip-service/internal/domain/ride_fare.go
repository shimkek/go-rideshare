package domain

import (
	"time"

	tripTypes "ride-sharing/services/trip-service/pkg/types"
	pb "ride-sharing/shared/proto/trip"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RideFareModel struct {
	ID                primitive.ObjectID
	UserID            string
	PackageSlug       string
	TotalPriceInCents float64
	ExpiresAt         time.Time
	Route             *tripTypes.OsrmApiResponse
}

func (f RideFareModel) ToProto() *pb.RideFare {
	return &pb.RideFare{
		Id:                f.ID.Hex(),
		UserID:            f.UserID,
		PackageSlug:       f.PackageSlug,
		TotalPriceInCents: f.TotalPriceInCents,
	}
}

func ToRideFareProto(fares []*RideFareModel) []*pb.RideFare {
	protoFares := make([]*pb.RideFare, len(fares))
	for i, f := range fares {
		protoFares[i] = f.ToProto()
	}
	return protoFares
}
