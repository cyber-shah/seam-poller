package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"
)

type PollingRequest struct {
	UserID          string `json:"userId"`
	APIEndpoint     string `json:"apiEndpoint"`
	PollingInterval int    `json:"pollingInterval"`
}

func logError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %s", err, message)
	}
}

func main() {
	// 1. connect to rabbitmq and log any errors
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	logError(err, "")
	channel, queue := setupQueue(connection)

	// 2. start the server and create the endpoints
	router := gin.Default()
	router.Handle("POST", "/polling-jobs", func(c *gin.Context) {
		create(c, channel, queue)
	})
	router.Run("localhost:8080")
}

/*
Sets up the connection and returns channel and queue
*/
func setupQueue(connection *amqp.Connection) (*amqp.Channel, *amqp.Queue) {
	// 2. create a channel
	channel, err := connection.Channel()
	logError(err, "")

	// 3. create a queue
	queue, err := channel.QueueDeclare(
		"schedulerQueue", // name
		true,             // persistant
		false,
		false,
		false,
		nil)
	logError(err, "")

	return channel, &queue
}

func create(c *gin.Context, channel *amqp.Channel, queue *amqp.Queue) {
	// Create a struct
	var requestBody PollingRequest

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
	err = channel.Publish(
		"",
		queue.Name,
		true,
		false,
		amqp.Publishing{
			ContentType: "application/json", // Set content type to JSON
			Body:        jsonBody,           // Use the JSON-encoded message body
		})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to publish message to queue"})
		return
	}
	fmt.Println(jsonBody)

	// Send a success message
	c.JSON(200, gin.H{"success": "Polling job created successfully"})
}
