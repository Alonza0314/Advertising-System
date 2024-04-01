package storage

import (
	"context"
	"log"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// define the mongo-client struct
type MgoClient struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

// insert an ad into the collection
func (c *MgoClient) InsertOneRecord(user *File) error {
	insertResult, err := c.collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	id := insertResult.InsertedID.(primitive.ObjectID)
	log.Println("Insert AD ID:", id.Hex())
	return nil
}

// establish a new mongo-client
func NewMgoClient(uri, database, table string) (*MgoClient, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	// ping to try if it is connected
	if err = client.Ping(context.Background(), nil); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Connected to MongoDB")
	db := client.Database(database)
	collection := db.Collection(table)
	return &MgoClient{client: client, db: db, collection: collection}, nil
}

// close the mongo-client
func CloseMongoDB(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Println(err)
	}
	log.Println("Disconnected from MongoDB")
}

// read the info of mongodb from config file
func SetUri(config string) (string, string, string, error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile(config)
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Error reading configuration file:", err)
		return "", "", "", err
	}
	uri := viper.GetString("mongodb.uri")
	database := viper.GetString("mongodb.database")
	collection := viper.GetString("mongodb.collection")
	return uri, database, collection, nil
}
