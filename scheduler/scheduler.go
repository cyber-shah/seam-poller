package scheduler

import (
	"encoding/json"
	"log"
	"time"

	"polling_service/helpers"

	"github.com/streadway/amqp"
)

func Init() {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial(helpers.RabbitMQURL)
	helpers.LogError(err, "")

	// Create a channel
	channel, _ := conn.Channel()
	schedulerQueue := helpers.SetupQ(channel, helpers.SchedulerQueueName)

	// Consume messages from the scheduler queue
	go consumeSchedulerQueue(channel, schedulerQueue)

	log.Println("Scheduler service started. Waiting for messages...")
	// Block the main goroutine
	select {}
}

// -----------------------------------------------------------------------------------------------
// CONSUME FROM THE SCHEDULER
// -----------------------------------------------------------------------------------------------
func consumeSchedulerQueue(ch *amqp.Channel, schedulerQueue *amqp.Queue) {
	// Consume messages from the scheduler queue
	msgs, _ := ch.Consume(
		schedulerQueue.Name, // queue
		"",                  // consumer
		true,                // auto-ack
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	// Process incoming messages
	for msg := range msgs {
		log.Println("Received a message:", string(msg.Body))
		// Spawn a goroutine to publish messages to the poller queue
		go publishPollerQueue(ch, msg.Body)
	}
}

// -----------------------------------------------------------------------------------------------
// PUBLISH TO THE POLLER
// -----------------------------------------------------------------------------------------------
func publishPollerQueue(ch *amqp.Channel, body []byte) {
	// Parse the message body as JSON
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	helpers.LogError(err, "Failed to unmarshal message")

	// Extract polling interval from the JSON data
	pollingInterval, ok := data["pollingInterval"].(float64)
	if !ok {
		log.Println("PollingInterval not found or not a valid number")
		return
	}

	q := helpers.SetupQ(ch, helpers.PollerQueueName)

	if pollingInterval > 0 {
		// Publish messages to the poller queue at regular intervals
		ticker := time.NewTicker(time.Duration(pollingInterval) * time.Millisecond)
		defer ticker.Stop()

		// at each time interval, publish to the queue
		for range ticker.C {
			helpers.Publish(ch, q, body)
		}
	}
}
