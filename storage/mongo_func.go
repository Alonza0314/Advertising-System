package storage

import (
	"context"
	"errors"
	"log"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	uri, database, collection, err := SetUri("project.conf")
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

// query ad from db
func QueryData(query QueryRequest) ([]File, error) {
	printLogGetRequest(query)

	// set mongodb connection
	uri, database, collection, err := SetUri("project.conf")
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
			{"conditions.Platform": nil},
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

// print the ad data get from client to log file
func printLogPostRequest(ad AdData) {
	log.Println("POST from:", ad.ClientIP)
	log.Println("POST Headers:")
	for key, vals := range ad.Headers {
		log.Print("\t", key, ": ", vals, "\n")
	}
	log.Println("POST Body:")
	log.Print("\t", ad.Ad, "\n")
}

// print the query requirement from client to log file
func printLogGetRequest(query QueryRequest) {
	log.Println("GET from:", query.ClientIP)
	log.Println("GET Headers:")
	for key, vals := range query.Headers {
		log.Print("\t", key, ": ", vals, "\n")
	}
	log.Println("GET Requirement:")
	log.Println("\t", "offset:", query.Offset)
	log.Println("\t", "limit:", query.Limit)
	log.Println("\t", "age:", query.Age)
	log.Println("\t", "gender:", query.Gender)
	log.Println("\t", "country:", query.Country)
	log.Println("\t", "platform:", query.Platform)
}
