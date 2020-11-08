package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func initMongo(connectionURI string) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))

	if err != nil {
		log.Fatal("Could not connect to Mongo")
	}

	return client
}

func getDocument(id string) (*donatorData, error) {
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)

	// create an OID bson primitive based on the ID that comes in on the request
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	// find and unmarshal the document to a struct we can return
	var data donatorData
	filter := bson.M{"_id": oid}
	err = collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func getDocumentByDonorID(id int64) (*donatorData, error) {
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)

	// find and unmarshal the document to a struct we can return
	var data donatorData
	filter := bson.M{"donor.id": id}
	err := collection.FindOne(context.Background(), filter).Decode(&data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func addDonation(documentId string, d donation) (*donatorData, error) {
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)

	// find and unmarshal the document to a struct we can return
	var data donatorData
	filter := bson.M{"_id": documentId}
	update := bson.M{
		"$push": bson.M{
			"donations": d,
		},
	}

	log.Info(fmt.Sprintf("Updating donor %v with donation amount %v", documentId, d.DonationValue))
	err := collection.FindOneAndUpdate(context.Background(), filter, update).Decode(&data)
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
	match := bson.D{
		primitive.E{
			Key: "$match",
			Value: bson.D{
				primitive.E{
					Key: "donationValue",
					Value: bson.D{
						primitive.E{
							Key:   "$gt",
							Value: 0,
						},
					},
				},
			},
		},
	}
	group := bson.D{
		primitive.E{
			Key: "$group",
			Value: bson.D{
				primitive.E{
					Key:   "_id",
					Value: "$originalTweetID",
				},
				primitive.E{
					Key: "total",
					Value: bson.D{
						primitive.E{
							Key:   "$sum",
							Value: "$donationValue",
						},
					},
				},
			},
		},
	}

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
	match := bson.D{
		primitive.E{
			Key: "$match",
			Value: bson.D{
				primitive.E{
					Key: "donations.donationValue",
					Value: bson.D{
						primitive.E{
							Key:   "$gt",
							Value: 0,
						},
					},
				},
			},
		},
	}
	group := bson.D{
		primitive.E{
			Key: "$group",
			Value: bson.D{
				primitive.E{
					Key:   "_id",
					Value: "$donor.screenname",
				},
				primitive.E{
					Key: "total",
					Value: bson.D{
						primitive.E{
							Key:   "$sum",
							Value: "$donationValue",
						},
					},
				},
			},
		},
	}

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
func aggregateAllDonatedTweets() (*[]donatorData, error) {
	filter := bson.D{
		primitive.E{
			Key: "donations.donationValue",
			Value: bson.D{
				primitive.E{
					Key:   "$gt",
					Value: 0,
				},
			},
		},
	}
	collection := mongoClient.Database(cfg.Database).Collection(cfg.Collection)
	resultCursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var results []donatorData
	if err = resultCursor.All(context.Background(), &results); err != nil {
		log.Error(err)
		return nil, err
	}

	return &results, nil
}
