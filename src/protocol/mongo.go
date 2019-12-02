package protocol

import (
	"context"
	"fmt"
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName     = "testing"
	collection = "rooms"
)

var connect sync.Once
var rooms *mongo.Collection

func GetRooms() (*mongo.Collection, error) {
	// Lazy Loading
	var err error
	connect.Do(func() {
		url := os.Getenv("MONGO_URL")
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://"+url))
		if err != nil {
			return
		}

		err = client.Ping(context.TODO(), nil)

		if err != nil {
			return
		}

		rooms = client.Database(dbName).Collection(collection)

	})

	if err != nil {
		return nil, err
	}

	return rooms, nil
}

func AddRoom(room *Room) {
	res, err := rooms.InsertOne(context.TODO(), bson.M{"name": "pi", "value": 3.14159})
	id := res.InsertedID
	fmt.Printf("Inserted Id %v\n", id)
	if err != nil {
		fmt.Println(err)
	}
}
