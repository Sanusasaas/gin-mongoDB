package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Car struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,"`
	Brand       string             `bson:"brand" json:"brand" validate:"required"`
	Model       string             `bson:"model" json:"model" validate:"required"`
	Year        int                `bson:"year" json:"year" validate:"required"`
	Price       float64            `bson:"price" json:"price" validate:"required"`
	IsAvailable bool               `bson:"is_available" json:"is_available"`
	Created_at  time.Time          `bson:"created_at" json:"created_at"`
	Updated_at  time.Time          `bson:"updated_at" json:"updated_at"`
}
