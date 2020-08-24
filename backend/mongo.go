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
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)

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
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)

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

// returns an aggregated collection matched by OriginalTweetID
// and sum up all the donationValues that match that OriginalTweetID
// TODO: pagination
func aggregateDonatedTweets() ([]bson.M, error) {
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)
	match := bson.D{{"$match", bson.D{{"donationValue", bson.D{{"$gt", 0}}}}}}
	group := bson.D{{"$group", bson.D{{"_id", "$originalTweetID"}, {"total", bson.D{{"$sum", "$donationValue"}}}}}}

	resultCursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{match, group})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var results []bson.M
	if err = resultCursor.All(context.Background(), &results); err != nil {
		log.Error(err)
		return nil, err
	}

	return results, nil
}

// returns an aggregated collection matched by invoker.ScreenName
// and sum up all the donationValues that match that invoker.ScreenName
// TODO: pagination
func aggregateDonors() ([]bson.M, error) {
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)
	match := bson.D{{"$match", bson.D{{"donationValue", bson.D{{"$gt", 0}}}}}}
	group := bson.D{{"$group", bson.D{{"_id", "$invoker.screenname"}, {"total", bson.D{{"$sum", "$donationValue"}}}}}}

	resultCursor, err := collection.Aggregate(context.Background(), mongo.Pipeline{match, group})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var results []bson.M
	if err = resultCursor.All(context.Background(), &results); err != nil {
		log.Error(err)
		return nil, err
	}

	return results, nil
}

// returns all data on tweets that have a successful donationValue logged to their document in Mongo
func aggregateAllDonatedTweets() (*[]charityYetiData, error) {
	filter := bson.D{{"donationValue", bson.D{{"$gt", 0}}}}
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)
	resultCursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var results []charityYetiData
	if err = resultCursor.All(context.Background(), &results); err != nil {
		log.Error(err)
		return nil, err
	}

	return &results, nil
}
