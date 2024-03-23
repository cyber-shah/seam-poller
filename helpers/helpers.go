package helpers

import (
	"log"

	"github.com/streadway/amqp"
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

/*
Sets up the connection and returns channel and queue
*/
func SetupQueue(connection *amqp.Connection, queueName string) (*amqp.Channel, *amqp.Queue) {
	// 2. create a channel
	channel, err := connection.Channel()
	LogError(err, "")

	// 3. create a queue
	queue, err := channel.QueueDeclare(
		queueName, // name
		true,      // persistant
		false,
		false,
		false,
		nil)
	LogError(err, "")

	return channel, &queue
}
