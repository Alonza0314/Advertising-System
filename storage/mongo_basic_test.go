package storage_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
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

// test insertonerecord
func TestInsertOneRecord(t *testing.T) {
	// set uri
	uri, database, collection, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// create a test db
	db := client.Database(database)
	coll := db.Collection(collection)
	testClient := &MgoClient{
		client:     client,
		db:         db,
		collection: coll,
	}

	// create a test AD
	ad := &File{
		Title:      "test AD",
		StartAt:    time.Now(),
		EndAt:      time.Now(),
		Conditions: []Condition{},
	}

	// call insert function
	err = testClient.InsertOneRecord(ad)
	if err != nil {
		t.Errorf("Failed to insert test record: %v", err)
	}

	// query to check test ad is in db
	result := coll.FindOne(context.Background(), bson.M{"title": ad.Title})
	if err := result.Err(); err != nil {
		t.Fatalf("Failed to find inserted test record: %v", err)
	}

	// decode it
	var insertedAd File
	if err := result.Decode(&insertedAd); err != nil {
		t.Fatalf("Failed to decode inserted test record: %v", err)
	}

	// check is it the right test ad
	if insertedAd.Title != ad.Title {
		t.Errorf("Inserted ad title does not match. Got: %s, Want: %s", insertedAd.Title, ad.Title)
	}

	// drop this test db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
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

// test newmgoclient
func TestNewMgoClient(t *testing.T) {
	// get mongo info
	uri, database, collection, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a new client instance
	client, err := NewMgoClient(uri, database, collection)
	defer func() {
		if err := client.client.Disconnect(context.Background()); err != nil {
			t.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	// check client is not nil
	assert.NotNil(t, client)
	// check err is nil
	assert.NoError(t, err)
	// check the database info is correct
	assert.Equal(t, client.db.Name(), database)
	assert.Equal(t, client.collection.Name(), collection)

	// drop this test db
	err = client.client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}
}

// close the mongo-client
func CloseMongoDB(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Println(err)
	}
	log.Println("Disconnected from MongoDB")
}

// test closemongodb
func TestCloseMongoDB(t *testing.T) {
	// set uri
	uri, _, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// ping to check it has connected
	assert.True(t, client.Ping(context.Background(), nil) == nil)

	// close it
	CloseMongoDB(client)

	// ping again to check it has disconnected
	assert.True(t, client.Ping(context.Background(), nil) != nil)
}

// read the info of mongodb from config file
func SetUri(config string) (string, string, string, error) {
	viper.SetConfigType("toml")
	viper.SetConfigFile("test.conf")
	if err := viper.ReadInConfig(); err != nil {
		log.Println("Error reading configuration file:", err)
		return "", "", "", err
	}
	uri := viper.GetString("mongodb.uri")
	database := viper.GetString("mongodb.database")
	collection := viper.GetString("mongodb.collection")
	return uri, database, collection, nil
}

// test seturi
func TestSetUri(t *testing.T) {
	// get correct mongodb info
	viper.SetConfigType("toml")
	viper.SetConfigFile("test.conf")
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read configuration file: %v", err)
		return
	}
	expected_uri := viper.GetString("mongodb.uri")
	expected_database := viper.GetString("mongodb.database")
	expected_collection := viper.GetString("mongodb.collection")

	// get test info
	uri, database, collection, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// check it
	if uri != expected_uri {
		t.Errorf("SetUri returned incorrect URI. Got: %s, Want: %s", uri, expected_uri)
	}
	if database != expected_database {
		t.Errorf("SetUri returned incorrect database name. Got: %s, Want: %s", database, expected_database)
	}
	if collection != expected_collection {
		t.Errorf("SetUri returned incorrect collection name. Got: %s, Want: %s", collection, expected_collection)
	}
}
