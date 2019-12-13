package protocol

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

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

func GetRooms() {
	// Lazy Loading
	connect.Do(func() {
		url := "mongodb://" + os.Getenv("MONGO_URL")
		fmt.Printf("Connecting to db %v\n", url)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			fmt.Println(err)
			return
		}

		err = client.Ping(context.TODO(), nil)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Connected to db")
		rooms = client.Database(dbName).Collection(collection)
		rooms.DeleteMany(context.TODO(), bson.D{})
	})
}

func AddRoom(room *Room) {
	GetRooms()
	fmt.Println("Adding room to database")
	res, err := rooms.InsertOne(context.TODO(), room)
	id := res.InsertedID
	fmt.Printf("Inserted Id %v\n", id)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Finished adding room to database")

	cursor, err := rooms.Find(context.TODO(), bson.D{})
	if err != nil {
		fmt.Println(err)
	}
	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}
	for _, result := range results {
		fmt.Println(result)
	}
}
