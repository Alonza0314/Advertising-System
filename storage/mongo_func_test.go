package storage_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// set the ad struct from POST request and the real ad struct
type Condition struct {
	AgeStart int      `json:"agestart"`
	AgeEnd   int      `json:"ageend"`
	Gender   []string `json:"gender"`
	Country  []string `json:"country"`
	Platform []string `json:"platform"`
}
type File struct {
	Title      string      `json:"title"`
	StartAt    time.Time   `json:"startat"`
	EndAt      time.Time   `json:"endat"`
	Conditions []Condition `json:"conditions"`
}
type AdData struct {
	ClientIP string
	Headers  map[string][]string
	Ad       File
}

// set the query struct from GET request
type QueryRequest struct {
	ClientIP string
	Headers  map[string][]string
	Offset   int
	Limit    int
	Age      int
	Gender   string
	Country  string
	Platform string
}

// insert ad into mongodb
func StoreData(ad AdData) error {
	printLogPostRequest(ad)

	// set mongodb connection
	uri, database, collection, err := SetUri("test.conf")
	if err != nil {
		return err
	}
	mgoClient, err := NewMgoClient(uri, database, collection)
	if err != nil {
		return err
	}
	defer CloseMongoDB(mgoClient.client)

	// insert
	if err = mgoClient.InsertOneRecord(&ad.Ad); err != nil {
		return err
	}

	return nil
}

// test storedata
func TestStoreData(t *testing.T) {
	test_file := &File{
		Title:      "test AD",
		StartAt:    time.Now(),
		EndAt:      time.Now(),
		Conditions: []Condition{},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

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

	// query to check if test ad is in db
	var result File
	err = coll.FindOne(context.Background(), bson.M{"title": test_ad.Ad.Title}).Decode(&result)
	if err != nil {
		t.Fatalf("Failed to find test ad in MongoDB: %v", err)
	}
	assert.Equal(t, test_ad.Ad.Title, result.Title)

	// drop this db
	err = db.Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// query ad from db
func QueryData(query QueryRequest) ([]File, error) {
	printLogGetRequest(query)

	// set mongodb connection
	uri, database, collection, err := SetUri("test.conf")
	if err != nil {
		return []File{}, err
	}
	mgoClient, err := NewMgoClient(uri, database, collection)
	if err != nil {
		return []File{}, err
	}
	defer CloseMongoDB(mgoClient.client)

	// set filter
	filter := bson.M{}
	filter["endat"] = bson.M{"$gt": time.Now()}
	if query.Age != 0 {
		filter["$or"] = []bson.M{
			{"$and": []bson.M{
				{"conditions.agestart": bson.M{"$lte": query.Age}},
				{"conditions.ageend": bson.M{"$gte": query.Age}},
			}},
			{"$and": []bson.M{
				{"conditions.agestart": 0},
				{"conditions.ageend": 0},
			}},
		}
	}
	if query.Gender != "" {
		filter["$or"] = []bson.M{
			{"conditions.gender": query.Gender},
			{"conditions.gender": nil},
		}
	}
	if query.Country != "" {
		filter["$or"] = []bson.M{
			{"conditions.country": query.Country},
			{"conditions.country": nil},
		}
	}
	if query.Platform != "" {
		filter["$or"] = []bson.M{
			{"conditions.platform": query.Platform},
			{"conditions.platform": nil},
		}
	}

	// set filter to cursor
	cursor, err := mgoClient.collection.Find(context.Background(), filter)
	if err != nil {
		return []File{}, err
	}
	defer cursor.Close(context.Background())

	// realize finding data
	var results []File
	if err := cursor.All(context.Background(), &results); err != nil {
		return []File{}, nil
	}

	// sort by end time
	sort.Slice(results, func(i, j int) bool { return results[i].EndAt.Before(results[j].EndAt) })

	// check offset and limit
	if query.Offset > len(results) {
		return []File{}, errors.New("offset is out of index in the query results")
	}
	if query.Offset+query.Limit > len(results) {
		results = results[query.Offset:]
	} else {
		results = results[query.Offset : query.Offset+query.Limit]
	}
	return results, nil
}

// test query with offset
func TestQuery_Offset(t *testing.T) {
	test_file0 := &File{
		Title:   "test AD0",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file1 := &File{
		Title:   "test AD1",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file2 := &File{
		Title:   "test AD2",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad0 := AdData{
		Ad: *test_file0,
	}
	test_ad1 := AdData{
		Ad: *test_file1,
	}
	test_ad2 := AdData{
		Ad: *test_file2,
	}

	// insert three test ads
	if err := StoreData(test_ad0); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad1); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad2); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   1,
		Limit:    5,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in offset in query: %v", err)
	}

	assert.Equal(t, 2, len(result))
	assert.Equal(t, test_ad1.Ad.Title, result[0].Title)
	assert.Equal(t, test_ad2.Ad.Title, result[1].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQuery_Offset_TooMuch(t *testing.T) {
	test_file0 := &File{
		Title:   "test AD0",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file1 := &File{
		Title:   "test AD1",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file2 := &File{
		Title:   "test AD2",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad0 := AdData{
		Ad: *test_file0,
	}
	test_ad1 := AdData{
		Ad: *test_file1,
	}
	test_ad2 := AdData{
		Ad: *test_file2,
	}

	// insert three test ads
	if err := StoreData(test_ad0); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad1); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad2); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   4,
		Limit:    5,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	_, err := QueryData(query)
	if err == nil {
		t.Fatalf("Failed with query in offset too much in query: %v", err)
	}
	assert.Equal(t, err, errors.New("offset is out of index in the query results"))
	

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// test query with limit
func TestQuery_Limit(t *testing.T) {
	test_file0 := &File{
		Title:   "test AD0",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file1 := &File{
		Title:   "test AD1",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file2 := &File{
		Title:   "test AD2",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad0 := AdData{
		Ad: *test_file0,
	}
	test_ad1 := AdData{
		Ad: *test_file1,
	}
	test_ad2 := AdData{
		Ad: *test_file2,
	}

	// insert three test ads
	if err := StoreData(test_ad0); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad1); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad2); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   1,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in limit in query: %v", err)
	}

	assert.Equal(t, 1, len(result))
	assert.Equal(t, test_ad1.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// test the result is sort by end time
func TestQuery_Sort(t *testing.T) {
	test_file0 := &File{
		Title:   "test AD0",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 3),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file1 := &File{
		Title:   "test AD1",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 2),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_file2 := &File{
		Title:   "test AD2",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad0 := AdData{
		Ad: *test_file0,
	}
	test_ad1 := AdData{
		Ad: *test_file1,
	}
	test_ad2 := AdData{
		Ad: *test_file2,
	}

	// insert three test ads
	if err := StoreData(test_ad0); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad1); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}
	if err := StoreData(test_ad2); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    3,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in sort in query: %v", err)
	}

	assert.Equal(t, 3, len(result))
	assert.Equal(t, test_ad2.Ad.Title, result[0].Title)
	assert.Equal(t, test_ad1.Ad.Title, result[1].Title)
	assert.Equal(t, test_ad0.Ad.Title, result[2].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// test query with age
func TestQueryData_Age_NoLimitInQuery(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 20,
				AgeEnd:   30,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in age no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Age_NoLimitInDB(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      50,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in age no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Age_BeforeStart(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 20,
				AgeEnd:   30,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      15,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in age before start in query: %v", err)
	}

	assert.Equal(t, 0, len(result))

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Age_BetweenStartAndEnd(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 20,
				AgeEnd:   30,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      25,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in age between in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Age_AfterEnd(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 20,
				AgeEnd:   30,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      35,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in age after end in query: %v", err)
	}

	assert.Equal(t, 0, len(result))

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// test query with gender
func TestQueryData_Gender_NoLimitInQuery(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   []string{"M", "F"},
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in gender no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Gender_NoLimitInDB(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "F",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in gender no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Gender_ConditionInTestData(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   []string{"M"},
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "M",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in gender in test in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Gender_ConditionNotInTestData(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   []string{"M"},
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "F",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in gender not in test in query: %v", err)
	}

	assert.Equal(t, 0, len(result))

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// test query with country
func TestQueryData_Country_NoLimtInQuery(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  []string{"TW", "JP"},
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in country no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Country_NoLimtInDB(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "TW",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in country no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Country_ConditionInTestData(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  []string{"TW", "JP"},
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "TW",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in country in test in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Country_ConditionNotInTestData(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  []string{"TW", "JP"},
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "USA",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in country not in test in query: %v", err)
	}

	assert.Equal(t, 0, len(result))

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// test query with platform
func TestQueryData_Platform_NoLimtInQuery(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: []string{"ios"},
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in platform no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Platform_NoLimtInDB(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: nil,
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "ios",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in platform no limit in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Platform_ConditionInTestData(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: []string{"ios", "MacOS"},
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "MacOS",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in platform in test in query: %v", err)
	}

	assert.Equal(t, test_ad.Ad.Title, result[0].Title)

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}
func TestQueryData_Platform_ConditionNotInTestData(t *testing.T) {
	test_file := &File{
		Title:   "test AD",
		StartAt: time.Now(),
		EndAt:   time.Now().AddDate(0, 0, 1),
		Conditions: []Condition{
			{
				AgeStart: 0,
				AgeEnd:   0,
				Gender:   nil,
				Country:  nil,
				Platform: []string{"ios"},
			},
		},
	}
	test_ad := AdData{
		Ad: *test_file,
	}

	// insert a test ad
	if err := StoreData(test_ad); err != nil {
		t.Fatalf("Fail to store test ad to MongoDB: %v", err)
	}

	// set query condition
	query := QueryRequest{
		ClientIP: "",
		Headers:  nil,
		Offset:   0,
		Limit:    1,
		Age:      0,
		Gender:   "",
		Country:  "",
		Platform: "MacOS",
	}

	// go to query
	result, err := QueryData(query)
	if err != nil {
		t.Fatalf("Failed with query in platform not in test in query: %v", err)
	}

	assert.Equal(t, 0, len(result))

	// set uri
	uri, database, _, err := SetUri("test.conf")
	if err != nil {
		t.Fatalf("SetUri returned an error: %v", err)
	}

	// establish a test client
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// drop this db
	err = client.Database(database).Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop test database: %v", err)
	}

	// defer to close it
	if err := client.Disconnect(context.TODO()); err != nil {
		t.Fatalf("Failes to disconnect the client: %v", err)
	}
}

// print the ad data get from client to log file
func printLogPostRequest(ad AdData) {
}

// print the query requirement from client to log file
func printLogGetRequest(query QueryRequest) {
}
