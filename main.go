package main

import (
	"polling_service/helpers"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"
)

// TODO: storing the active polling jobs in a persistent storage mechanism
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

// -----------------------------------------------------------------------------------------------
// API ENDPOINT POST /POLLING-JOBS
// -----------------------------------------------------------------------------------------------
func create(c *gin.Context, channel *amqp.Channel, queue *amqp.Queue) {
	jsonBody := helpers.ConvertToJson(c)

	// Publish the JSON message to the queue
	helpers.Publish(channel, queue, jsonBody)

	// Send a success message
	c.JSON(200, gin.H{"success": "Polling job created successfully"})
}
