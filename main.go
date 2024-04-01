package main

import (
	"log"
	"os"

	"dcard/process"
	"github.com/gin-gonic/gin"
)

func main() {
	// set a log file to monitor the conditions
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// set a router
	router := gin.Default()

	router.POST("/api/v1/ad", process.ProcessPost)
	router.GET("/api/v1/ad", process.ProcessGet)

	router.Run()
}
