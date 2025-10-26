package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client{
	err := godotenv.Load(".env")
	if err!=nil{
		log.Fatal("Error .env file loading")
	}

	MongoDB := os.Getenv("MONGODB_URL")

	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err!=nil{
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err!=nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")

	return client
}

var Client *mongo.Client = DBinstance()
func GetDBNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		log.Fatal("Invalid MongoDB URL format")
	}
	return parts[3]
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection{
	MongoDB := os.Getenv("MONGODB_URL")
	dbName := GetDBNameFromURL(MongoDB)
	collection := client.Database(dbName).Collection(collectionName)
	return collection
}