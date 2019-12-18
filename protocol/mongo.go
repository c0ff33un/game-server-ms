package protocol

import (
	"context"
	"log"
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
		url := os.Getenv("MONGO_URL")
		log.Printf("Connecting to db %v\n", url)

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			log.Println(err)
			return
		}

		err = client.Ping(context.TODO(), nil)

		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Connected to db")
		rooms = client.Database(dbName).Collection(collection)
		rooms.DeleteMany(context.TODO(), bson.D{})
	})
}

func AddRoom(room *Room) {
	GetRooms()
	log.Println("Adding room to database")
	res, err := rooms.InsertOne(context.TODO(), room)
	id := res.InsertedID
	log.Printf("Inserted Id %v\n", id)
	if err != nil {
		log.Println(err)
	}
	log.Println("Finished adding room to database")
}
