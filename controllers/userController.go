package controllers

import (
	"context"
	"ginProj/database"
	helper "ginProj/helpers"
	"ginProj/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) (string, error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err!=nil{
		log.Panic(err)
	}
	return string(bytes), nil
}

func VerifyPassword(userPassword string, providedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true

	if err!=nil{
		check = false
		return check, err
	}
	return check, nil
}

func Signup() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		validationErr := validate.Struct(user)
		if validationErr != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":validationErr.Error()})
			return 
		}

		password, err := HashPassword(*user.Password)
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to hash password"})
			return 
		}
		user.Password = &password

		count, err := userCollection.CountDocuments(ctx, bson.M{"email":user.Email});
		if err!=nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email checking error"})
		}
		if count > 0{
			c.JSON(http.StatusConflict, gin.H{"error":"this email already exists"})
		}
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone":user.Phone})
		if err!=nil{
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error":"phone number checking error"})
		}
		if count > 0{
			c.JSON(http.StatusConflict, gin.H{"error":"this phone number already exists"})
		}
		
		user.Created_at = time.Now()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, reToken, _ := helper.GenerateAllTokens(*user.Email, *user.Name, *user.User_type, user.User_id)

		user.Token = &token
		user.Refresh_token = &reToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user item not was created"})
			return 
		}
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
    return func(c *gin.Context) {
        var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
        defer cancel()
        var user models.User
        var foundUser models.User
		log.Printf("Email: %v, Password: %v, Name: %v, User_type: %v", foundUser.Email, foundUser.Password, foundUser.Name, foundUser.User_type)

        if err := c.BindJSON(&user); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "email is incorrect"})
            return
        }

        if foundUser.Email == nil || foundUser.Password == nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }

        passwordIsValid, err := VerifyPassword(*user.Password, *foundUser.Password)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify password"})
            return
        }
        if !passwordIsValid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "password is incorrect"})
            return
        }

        token, reToken, err := helper.GenerateAllTokens(*foundUser.Email, *foundUser.Name, *foundUser.User_type, foundUser.User_id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
            return
        }

        err = helper.UpdateAllTokens(token, reToken, foundUser.User_id)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tokens"})
            return
        }

        err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "user not found after token update"})
            return
        }

        c.JSON(http.StatusOK, foundUser)
    }
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage

		pipeline := mongo.Pipeline{
			bson.D{
				{Key: "$facet", Value: bson.D{
					{Key: "metadata", Value: bson.A{
						bson.D{{Key: "$count", Value: "total"}},
					}},
					{Key: "data", Value: bson.A{
						bson.D{{Key: "$skip", Value: startIndex}},
						bson.D{{Key: "$limit", Value: recordPerPage}},
					}},
				}},
			},
		}

		result, err := userCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while fetching user items"})
			return
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode user items"})
			return
		}

		if len(allUsers) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []interface{}{}, "total": 0})
			return
		}

		metadataArray, ok := allUsers[0]["metadata"].(bson.A)
		if !ok || len(metadataArray) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": allUsers[0]["data"], "total": 0})
			return
		}

		metadata, ok := metadataArray[0].(bson.M)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"data": allUsers[0]["data"], "total": 0})
			return
		}

		total, ok := metadata["total"].(int32)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"data": allUsers[0]["data"], "total": 0})
			return
		}

		response := gin.H{
			"total": total,
			"data":  allUsers[0]["data"],
		}

		c.JSON(http.StatusOK, response)
	}
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context){
		userId := c.Param("user_id")

		if err:= helper.MatchUserTypeToUid(c, userId); err!= nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return 
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id":userId}).Decode(&user)
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
			return 
		}
		c.JSON(http.StatusOK, user)
	}
}