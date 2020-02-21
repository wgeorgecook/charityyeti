package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func initMongo(connectionURI string) *mongo.Client {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))

	if err != nil {
		log.Fatal("Could not connect to Mongo")
	}

	return client
}

func getDocument(id string) (*charityYetiData, error) {
	collection := mongoClient.Database("cfg.Database").Collection("twitterData")

	// create an OID bson primitive based on the ID that comes in on the request
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// find and unmarshal the document to a struct we can return
	var data charityYetiData
	filter := bson.M{"_id": oid}
	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func updateDocument(u charityYetiData) (*charityYetiData, error) {
	collection := mongoClient.Database("cfg.Database").Collection("twitterData")

	// create an OID bson primitive based on the ID that comes in on the request
	oid, err := primitive.ObjectIDFromHex(u.ID)
	if err != nil {
		return nil, err
	}

	// find and unmarshal the document to a struct we can return
	var data charityYetiData
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": bson.M{"donationValue": u.DonationValue}}

	log.Info(fmt.Sprintf("Updating record %v with donationValue %v", u.ID, u.DonationValue))
	err = collection.FindOneAndUpdate(context.Background(), filter, update).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
