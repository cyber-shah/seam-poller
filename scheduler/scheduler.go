package scheduler

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"polling_service/helpers"

	"github.com/streadway/amqp"
)

type PollingRequest struct {
	UserID          string `json:"userId"`
	APIEndpoint     string `json:"apiEndpoint"`
	PollingInterval int    `json:"pollingInterval"`
}

const (
	schedulerQueueName = "schedulerQueue"
	pollerQueueName    = "pollerQueue"
	rabbitMQURL        = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial(rabbitMQURL)
	helpers.LogError(err, "")

	// Create a channel
	channel, pollerQueue := helpers.SetupQueue(conn, pollerQueueName)

	// Consume messages from the scheduler queue
	go consumeSchedulerQueue(channel)

	log.Println("Scheduler service started. Waiting for messages...")
	// Block the main goroutine
	select {}
}

// Function to consume messages from the scheduler queue
func consumeSchedulerQueue(ch *amqp.Channel) {
	// Declare the scheduler queue
	q, err := ch.QueueDeclare(
		schedulerQueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Consume messages from the scheduler queue
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	// Process incoming messages
	for msg := range msgs {
		fmt.Println("Received a message:", string(msg.Body))

		// Spawn a goroutine to publish messages to the poller queue
		go publishPollerQueue(ch, msg.Body)
	}
}

// Function to publish message to the poller queue
func publishPollerQueue(ch *amqp.Channel, body []byte) {
	// Parse the message body as JSON
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	// Extract polling interval from the JSON data
	pollingInterval, ok := data["pollingInterval"].(float64)
	if !ok {
		log.Println("PollingInterval not found or not a valid number")
		return
	}
	fmt.Println(pollingInterval)

	// Declare the poller queue
	q, err := ch.QueueDeclare(
		pollerQueueName, // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %v", err)
		return
	}

	// Publish messages to the poller queue at regular intervals
	ticker := time.NewTicker(time.Duration(pollingInterval) * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		err := ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			},
		)
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
		}
	}
}
