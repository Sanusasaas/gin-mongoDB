package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Book struct {
	CarID    primitive.ObjectID `bson:"car_id" json:"car_id"`
	BookedAt time.Time          `bson:"booked_at" json:"booked_at"`
}
