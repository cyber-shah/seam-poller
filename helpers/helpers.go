package helpers

import (
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

const (
	SchedulerQueueName = "schedulerQueue"
	PollerQueueName    = "pollerQueue"
	RabbitMQURL        = "amqp://guest:guest@localhost:5672/"
)

func LogError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %s", err, message)
	}
}

type PollingRequest struct {
	UserID          string `json:"userId"`
	APIEndpoint     string `json:"apiEndpoint"`
	PollingInterval int    `json:"pollingInterval"`
}

func Publish(channel *amqp.Channel, queue *amqp.Queue, body []byte) {
	channel.Publish(
		"",
		queue.Name,
		true,
		false,
		amqp.Publishing{
			ContentType: "application/json", // Set content type to JSON
			Body:        body,               // Use the JSON-encoded message body
		})
}

// queue setup to remain modular
func SetupQ(channel *amqp.Channel, queueName string) *amqp.Queue {
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // persistant
		false,
		false,
		false,
		nil)
	LogError(err, "Failed to declare queue")

	return &queue
}

func ConvertToJson(context *gin.Context) []byte {
	var requestBody PollingRequest
	// Parse request body to struct
	if err := context.ShouldBind(&requestBody); err != nil {
		context.JSON(400, gin.H{"error": err.Error()})
	}

	// Marshal the struct to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		context.JSON(500, gin.H{"error": "Failed to marshal JSON"})
	}

	return jsonBody
}
