package main

import (
	// "fmt"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"
)

/**
	1. User specifizac polling - diff jobs for diff users
	2. Polling job creation - accept the job API and interval
	3. Poll the API at interval, interval is specified and must be configurable
	4. Store response from the third party -- inside a Database
	5. Duplicate data handling - agnostic of DS & content sent by 3rd party
	6. Persistence of jobs

	Example request:
	```json
	POST /polling-jobs
	{
	"userId": "user123",
	"apiEndpoint": "docs.dummyapi.online/api/endpoint",
	"pollingInterval": 60000
	}
```

TODO:		error handling
				logging
				Multiple polling jobs running concurrently
				Instructions of how to run and test the service

*/

/*
	Design:
	Two microservices :
	1. Requests that handles our first line of APIs and database -- the user works with.
	2. Poller that actually polls the requested service.
	# A message queue that sits in between both the services
	#
	Why ?
	Allows higher scalability and performance as the Poller might need to scale horizontally
	while requests might not.
	More Pollers can be added if the queue gets filled up quickly
*/

type PollingRequest struct {
	UserId          string
	ApiEndpoint     string
	PollingInterval time.Duration
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
	router.Handle("POST", "/create", func(c *gin.Context) {
		create(c, channel, queue)
	})
	router.Handle("GET", "/read", func(c *gin.Context) {
		read(c, channel, queue)
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
		"apiQueue", // name
		true,       // persistant
		false,
		false,
		false,
		nil)
	logError(err, "")

	return channel, &queue
}

func create(connection *gin.Context, channel *amqp.Channel, queue *amqp.Queue) {
	// create a struct
	var requestBody struct {
		UserID          string `json:"userId"`
		APIEndpoint     string `json:"apiEndpoint"`
		PollingInterval int    `json:"pollingInterval"`
	}

	// parse json to struct
	connection.ShouldBindJSON(&requestBody)

	messageBody := fmt.Sprintf("userId : %s, APIEndpoint: %s, PollingInterval: %d",
		requestBody.UserID,
		requestBody.APIEndpoint,
		requestBody.PollingInterval)

	channel.Publish(
		"",
		queue.Name,
		true,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(messageBody),
		})
}
