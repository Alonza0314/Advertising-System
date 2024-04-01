package process_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dcard/storage"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func ProcessPost(c *gin.Context) {
	var ad storage.AdData

	// parse the data into json struct
	if err := c.ShouldBindJSON(&ad.Ad); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// if title is nil, return an error
	if ad.Ad.Title == "" {
		err := errors.New("title is nil")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// if start and end is nil, return an error
	if ad.Ad.StartAt == (time.Time{}) {
		err := errors.New("start time is nil")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ad.Ad.EndAt == (time.Time{}) {
		err := errors.New("end time is nil")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// if condition is not set, give it an all nil condition
	if len(ad.Ad.Conditions) == 0 {
		ad.Ad.Conditions = append(ad.Ad.Conditions, storage.Condition{
			AgeStart: 0,
			AgeEnd:   0,
			Gender:   []string{},
			Country:  []string{},
			Platform: []string{},
		})
	}

	// record the clientIP
	ad.ClientIP = c.ClientIP()

	// record the clientIP
	ad.Headers = make(map[string][]string)
	for key, vals := range c.Request.Header {
		ad.Headers[key] = make([]string, 0)
		ad.Headers[key] = append(ad.Headers[key], vals...)
	}

	// call store function to store data
	// if err := storage.StoreData(ad); err != nil {
	// 	log.Println(err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// }

	// return a success feedback
	c.JSON(http.StatusOK, gin.H{ad.Ad.Title: "POST successfully"})
}

// test processpost with nil title
func TestProcessPost_Title(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the endpoint handler
	router.POST("/test-title", ProcessPost)

	// Create a request body with nil title
	requestBody := string("{\"startAt\": \"2024-01-21T16:00:00.000Z\", \"endAt\": \"2024-06-21T16:00:00.000Z\"}")

	// Perform a POST request with the test payload
	req, err := http.NewRequest("POST", "/test-title", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Serve the request to the recorder
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	// Decode the response body
	var responseBody map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the error message is as expected
	assert.Equal(t, "title is nil", responseBody["error"])
}

// test processpost with nil start time
func TestProcessPost_StartTime(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the endpoint handler
	router.POST("/test-start", ProcessPost)

	// Create a request body with nil title
	requestBody := string("{\"title\": \"test AD\",\"endAt\": \"2024-06-21T16:00:00.000Z\"}")

	// Perform a POST request with the test payload
	req, err := http.NewRequest("POST", "/test-start", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Serve the request to the recorder
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	// Decode the response body
	var responseBody map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the error message is as expected
	assert.Equal(t, "start time is nil", responseBody["error"])
}

// test processpost with nil start time
func TestProcessPost_EndTime(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the endpoint handler
	router.POST("/test-end", ProcessPost)

	// Create a request body with nil title
	requestBody := string("{\"title\": \"test AD\",\"startAt\": \"2024-01-21T16:00:00.000Z\"}")

	// Perform a POST request with the test payload
	req, err := http.NewRequest("POST", "/test-end", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Serve the request to the recorder
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	// Decode the response body
	var responseBody map[string]string
	err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatal(err)
	}

	// Check if the error message is as expected
	assert.Equal(t, "end time is nil", responseBody["error"])
}

// test processpost success
func TestProcess_Success(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the endpoint handler
	router.POST("/test-success", ProcessPost)

	// Create a request body with nil title
	requestBody := string("{\"Title\": \"test AD\", \"startAt\": \"2024-01-21T16:00:00.000Z\", \"endAt\": \"2024-06-21T16:00:00.000Z\"}")

	// Perform a POST request with the test payload
	req, err := http.NewRequest("POST", "/test-success", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	recorder := httptest.NewRecorder()

	// Serve the request to the recorder
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check if the error message is as expected
	assert.Equal(t, "{\"test AD\":\"POST successfully\"}", recorder.Body.String())
}
