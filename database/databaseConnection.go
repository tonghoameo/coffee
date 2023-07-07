package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {

	MongoDB := "mongodb+srv://binbomb:LfitLbcGI8CT25NQ@cluster0.vsi7e.gcp.mongodb.net/?retryWrites=true&w=majority"
	fmt.Println(MongoDB)
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected db successful")
	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection = client.Database("storeCoffee").Collection(collectionName)
	return collection
}
