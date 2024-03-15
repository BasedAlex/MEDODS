package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/basedalex/medods-test/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webPort  = "8080"
	mongoURL = "mongodb://localhost:27018"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo
	mongoClient, err := connectToMongo()

	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err  = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := &Config{
		Models: data.New(client),
	}

	log.Println("Starting service on port", webPort)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	coll := client.Database("logs").Collection("refresh_token")
	
	indexName, err := coll.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "userid", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Print(err, indexName)
	}

	err = srv.ListenAndServe()

	if err != nil {
		log.Panic()
	}
}

func connectToMongo() (*mongo.Client, error) {

	// create connection options
	clientOptions := options.Client().ApplyURI(mongoURL)

	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect 
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting:", err)
		return nil, err
	}

	log.Println("Connected to mongo!")

	return c, nil
}