package process_test

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"dcard/storage"

	"github.com/gin-gonic/gin"
)

type item struct {
	Title string    `json:"title"`
	Endat time.Time `json:"endAt"`
}

func ProcessGet(c *gin.Context) {
	var query storage.QueryRequest

	// record the clientIP
	query.ClientIP = c.ClientIP()

	// record the clientIP
	query.Headers = make(map[string][]string)
	for key, vals := range c.Request.Header {
		query.Headers[key] = make([]string, 0)
		query.Headers[key] = append(query.Headers[key], vals...)
	}

	// parse the query requirement
	query.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "5"))
	if query.Offset < 1 || query.Offset > 100 {
		err := errors.New("offset should be in this interval: [1, 100]")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	query.Offset -= 1
	query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "5"))
	if query.Limit < 1 || query.Limit > 100 {
		err := errors.New("limit should be in this interval: [1, 100]")
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	query.Age, _ = strconv.Atoi(c.DefaultQuery("age", "0"))
	query.Gender = c.DefaultQuery("gender", "")
	query.Country = c.DefaultQuery("country", "")
	query.Platform = c.DefaultQuery("platform", "")

	// call function to query data
	// results, err := storage.QueryData(query)
	// if err != nil {
	// 	log.Println(err)
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// }

	// get title and endat, then store as an array
	// items := make([]item, len(results))
	// for i, result := range results {
	// 	items[i].Title = result.Title
	// 	items[i].Endat = result.EndAt
	// 	log.Println(items[i])
	// }

	test_items := make([]item, 0)

	// return the query results with json format
	c.JSON(http.StatusOK, gin.H{"items": test_items})
}

// test processget with offest < 1 || offset > 100
func TestProcessGet_Offset_OutRange(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the endpoint handler
	router.GET("/test-offset", ProcessGet)

	// Perform GET requests with out of offset == 0 || 101
	req0, err := http.NewRequest("GET", "/test-offset?offset=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	req101, err := http.NewRequest("GET", "/test-offset?offset=101", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	recorder0 := httptest.NewRecorder()
	recorder101 := httptest.NewRecorder()

	// Serve the request to the recorder
	router.ServeHTTP(recorder0, req0)
	router.ServeHTTP(recorder101, req101)


	// Check the response status code
	if recorder0.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder0.Code)
	}
	if recorder101.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder101.Code)
	}

	// Decode the response body
	var responseBody0 map[string]string
	if err := json.Unmarshal(recorder0.Body.Bytes(), &responseBody0); err != nil {
		t.Fatal(err)
	}
	var responseBody101 map[string]string
	if err := json.Unmarshal(recorder0.Body.Bytes(), &responseBody101); err != nil {
		t.Fatal(err)
	}

	// Check if the error message is as expected
	expectedErrorMessage := "offset should be in this interval: [1, 100]"
	if responseBody0["error"] != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, responseBody0["error"])
	}
	if responseBody101["error"] != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, responseBody101["error"])
	}
}

// test processget with offest == [1, 100]
func TestProcessGet_Offset_InRange(t *testing.T) {
    // Create a new Gin router
    router := gin.Default()

    // Set up the endpoint handler
    router.GET("/test-offset", ProcessGet)

    // Perform a GET request with valid offset
    req, err := http.NewRequest("GET", "/test-offset?offset=55", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Create a response recorder
    recorder := httptest.NewRecorder()

    // Serve the request to the recorder
    router.ServeHTTP(recorder, req)

    // Check the response status code
    if recorder.Code != http.StatusOK {
        t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
    }

    // Decode the response body
    var responseBody map[string][]item
    if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
        t.Fatal(err)
    }

    // Check if the response contains items array
    if _, ok := responseBody["items"]; !ok {
        t.Errorf("Expected response to contain 'items' key, got %v", responseBody)
    }
}

// test processget with limit < 1 || limit > 100
func TestProcessGet_Limit_OutRange(t *testing.T) {
	// Create a new Gin router
	router := gin.Default()

	// Set up the endpoint handler
	router.GET("/test-limit", ProcessGet)

	// Perform GET requests with out of limit == 0 || 101
	req0, err := http.NewRequest("GET", "/test-limit?limit=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	req101, err := http.NewRequest("GET", "/test-limit?limit=101", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	recorder0 := httptest.NewRecorder()
	recorder101 := httptest.NewRecorder()

	// Serve the request to the recorder
	router.ServeHTTP(recorder0, req0)
	router.ServeHTTP(recorder101, req101)


	// Check the response status code
	if recorder0.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder0.Code)
	}
	if recorder101.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder101.Code)
	}

	// Decode the response body
	var responseBody0 map[string]string
	if err := json.Unmarshal(recorder0.Body.Bytes(), &responseBody0); err != nil {
		t.Fatal(err)
	}
	var responseBody101 map[string]string
	if err := json.Unmarshal(recorder0.Body.Bytes(), &responseBody101); err != nil {
		t.Fatal(err)
	}

	// Check if the error message is as expected
	expectedErrorMessage := "limit should be in this interval: [1, 100]"
	if responseBody0["error"] != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, responseBody0["error"])
	}
	if responseBody101["error"] != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, responseBody101["error"])
	}
}

// test processget with limit == [1, 100]
func TestProcessGet_Limit_InRange(t *testing.T) {
    // Create a new Gin router
    router := gin.Default()

    // Set up the endpoint handler
    router.GET("/test-limit", ProcessGet)

    // Perform a GET request with valid offset
    req, err := http.NewRequest("GET", "/test-limit?limit=55", nil)
    if err != nil {
        t.Fatal(err)
    }

    // Create a response recorder
    recorder := httptest.NewRecorder()

    // Serve the request to the recorder
    router.ServeHTTP(recorder, req)

    // Check the response status code
    if recorder.Code != http.StatusOK {
        t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
    }

    // Decode the response body
    var responseBody map[string][]item
    if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
        t.Fatal(err)
    }

    // Check if the response contains items array
    if _, ok := responseBody["items"]; !ok {
        t.Errorf("Expected response to contain 'items' key, got %v", responseBody)
    }
}
