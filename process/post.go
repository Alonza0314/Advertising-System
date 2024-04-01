package process

import (
	"errors"
	"log"
	"net/http"
	"time"

	"dcard/storage"

	"github.com/gin-gonic/gin"
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
	if err := storage.StoreData(ad); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	// return a success feedback
	c.JSON(http.StatusOK, gin.H{ad.Ad.Title: "POST successfully"})
}
