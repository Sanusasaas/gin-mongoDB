package controllers

import (
	"context"
	"ginProj/database"
	helper "ginProj/helpers"
	"ginProj/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var carCollection *mongo.Collection = database.OpenCollection(database.Client, "car")

func CreateCar() gin.HandlerFunc{
	return func(c *gin.Context){
		if err := helper.CheckUserType(c, "ADMIN"); err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return 
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var car models.Car
		if err := c.BindJSON(&car); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return 
		}

		validationErr := validate.Struct(car)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return 
		}

		car.IsAvailable = true
		car.Created_at = time.Now()
		car.Updated_at = car.Created_at
		car.ID = primitive.NewObjectID()

		result, err := carCollection.InsertOne(ctx, car)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to create car"})
			return 
		}
		
		c.JSON(http.StatusOK, result)
	}
}

func GetCars() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := carCollection.Find(ctx, bson.M{})
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch cars"})
			return 
		}
		defer cursor.Close(ctx)

		var cars[]bson.M
		if err = cursor.All(ctx, &cars); err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode cars"})
			return 
		}

		c.JSON(http.StatusOK, cars)
	}
}

func BookCar() gin.HandlerFunc {
    return func(c *gin.Context) {
        userId := c.Param("user_id")
        if err := helper.MatchUserTypeToUid(c, userId); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()

        carId := c.Param("car_id")
        if carId == "" {
            c.JSON(http.StatusBadRequest, gin.H{"error": "car_id is required"})
            return
        }

        objId, err := primitive.ObjectIDFromHex(carId)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid car_id format"})
            return
        }

        var car models.Car
        err = carCollection.FindOne(ctx, bson.M{"_id": objId}).Decode(&car)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "car not found"})
            return
        }

        if !car.IsAvailable {
            c.JSON(http.StatusBadRequest, gin.H{"error": "car is not available"})
            return
        }

        var user models.User
        err = userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
        if err != nil {
            log.Printf("User not found: %v", err)
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }

        if user.Books == nil {
            _, err = userCollection.UpdateOne(
                ctx,
                bson.M{"user_id": userId},
                bson.M{"$set": bson.M{"books": []bson.M{}}},
            )
            if err != nil {
                log.Printf("Failed to initialize books: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize books"})
                return
            }
        }

        book := models.Book{
            CarID:    objId,
            BookedAt: time.Now(),
        }

        updateUser := bson.M{
            "$push": bson.M{
                "books": book,
            },
        }

        _, err = userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, updateUser)
        if err != nil {
            log.Printf("Failed to update user bookings: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user bookings"})
            return
        }

        updateCar := bson.M{
            "$set": bson.M{
                "is_available": false,
                "updated_at":   time.Now(),
            },
        }

        _, err = carCollection.UpdateOne(ctx, bson.M{"_id": objId}, updateCar)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to book car"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "car booked successfully"})
    }
}

func GetUserBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}
		userId := c.Param("user_id")
		if userId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error":"unauthorized"})
			return 
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return 
		}

		c.JSON(http.StatusOK, user.Books)
	}
}

func UpdateCar() gin.HandlerFunc {
	return func(c *gin.Context){
		if err := helper.CheckUserType(c, "ADMIN"); err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		carId := c.Param("car_id")
		if carId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error": "car_id is required"})
			return 
		}

		var car models.Car
		if err := c.BindJSON(&car); err!=nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return 
		}

		validationErr := validate.Struct(car)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return 
		}

		objId, err := primitive.ObjectIDFromHex(carId)
		if err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":"invalid car_id format"})
			return 
		}

		update := bson.M{
			"$set": bson.M{
				"brand": car.Brand,
				"model": car.Model,
				"year": car.Year,
				"price": car.Price,
				"is_available": car.IsAvailable,
				"updated_at": time.Now(),
			},
		}
		
		_, err = carCollection.UpdateOne(ctx, bson.M{"_id":objId}, update)
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to update car"})
			return 
		}
		c.JSON(http.StatusOK, gin.H{"message":"car updated successfully"})
	}
}

func DeleteCar() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN");err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return 
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		carId := c.Param("car_id")
		if carId == ""{
			c.JSON(http.StatusBadRequest, gin.H{"error":"car id is required"})
			return
		}

		objId, err := primitive.ObjectIDFromHex(carId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"invalid car id format"})
			return 
		}

		result, err := carCollection.DeleteOne(ctx, bson.M{"_id": objId})
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to delete this car"})
			return 
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error":"car not found"})
			return 
		}

		c.JSON(http.StatusOK, gin.H{"message":"car was deleted successfully"})
	}
}