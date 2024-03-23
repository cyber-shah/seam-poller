package main

import (
	"encoding/json"

	"polling_service/helpers"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. connect to rabbitmq and log any errors
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	helpers.LogError(err, "")
	channel, _ := connection.Channel()
	queue := helpers.SetupQ(channel, "schedulerQueue")

	// 2. start the server and create the endpoints
	router := gin.Default()
	router.Handle("POST", "/polling-jobs", func(c *gin.Context) {
		create(c, channel, queue)
	})
	router.Run("localhost:8080")
}

func create(c *gin.Context, channel *amqp.Channel, queue *amqp.Queue) {
	// Create a struct
	var requestBody helpers.PollingRequest

	// Parse request body to struct
	if err := c.ShouldBind(&requestBody); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Marshal the struct to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal JSON"})
		return
	}

	// Publish the JSON message to the queue
	helpers.Publish(channel, queue, jsonBody)

	// Send a success message
	c.JSON(200, gin.H{"success": "Polling job created successfully"})
}
