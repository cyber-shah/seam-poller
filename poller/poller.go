package main

// TODO: logging everywhere

import (
	// "log"
	"github.com/streadway/amqp"
)

const (
	pollerQueueName = "pollerQueue"
	rabbitMQURL     = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// Connect to RabbitMQ server
	conn, _ := amqp.Dial(rabbitMQURL)
	// Create a channel
	ch, _ := conn.Channel()
	// Consume messages from the poller queue
	consumePollerQueue(ch)
}

// Function to consume messages from the poller queue
func consumePollerQueue(ch *amqp.Channel) {
}
