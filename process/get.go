package process

import (
	"errors"
	"log"
	"net/http"
	"strconv"
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
	results, err := storage.QueryData(query)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// get title and endat, then store as an array
	items := make([]item, len(results))
	for i, result := range results {
		items[i].Title = result.Title
		items[i].Endat = result.EndAt
		log.Println(items[i])
	}

	// return the query results with json format
	c.JSON(http.StatusOK, gin.H{"items": items})
}
