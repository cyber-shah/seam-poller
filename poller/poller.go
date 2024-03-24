package main

// TODO: logging everywhere

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"polling_service/helpers"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

func main() {
	// Connect to RabbitMQ server
	conn, _ := amqp.Dial(helpers.RabbitMQURL)
	// Create a channel
	ch, _ := conn.Channel()

	db := helpers.ConnectToDB()

	// Consume messages from the poller queue
	consumePollerQueue(ch, db)
}

// Function to consume messages from the poller queue
func consumePollerQueue(ch *amqp.Channel, db *sql.DB) {
	pollerQueue := helpers.SetupQ(ch, helpers.PollerQueueName)

	msgs, _ := ch.Consume(pollerQueue.Name, "", true, false, false, false, nil)

	// for each message in message call the API
	for msg := range msgs {
		// spawn a concurrent go quque
		go processMessage(msg.Body, db)
	}
}

// =======================================================================================
// 1. PARSES THE MESSAGE
// 2. CALLS THE API
// 3. CHECKS DUPLICATES
// 4. ADDS TO DB
// =======================================================================================
func processMessage(msg []byte, db *sql.DB) []byte {
	var requestBody helpers.PollingRequest

	err := json.Unmarshal(msg, &requestBody)
	helpers.LogError(err, "unable to unmarshal")

	// Call the API
	response, err := http.Get(requestBody.APIEndpoint)
	helpers.LogError(err, "failed to make HTTP GET request")

	body, err := io.ReadAll(response.Body)
	helpers.LogError(err, "failed to read response body")

	// compute the hash of the current item
	hash := md5.New()
	hash.Write([]byte(body))
	hashValue := hex.EncodeToString(hash.Sum(nil))

	// Check if the hash already exists in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM api_responses WHERE hash = $1", hashValue).Scan(&count)
	helpers.LogError(err, "failed to query database")
	if count > 0 {
		log.Println("Duplicate data detected, skipping insertion")
		return body
	}

	// Prepare SQL statement
	statement, err := db.Prepare("INSERT INTO api_responses (user_id, response_body, hash) VALUES ($1, $2, $3)")
	helpers.LogError(err, "unable to prepare SQL statement")

	// Execute SQL statement
	_, err = statement.Exec(requestBody.UserID, string(body), hashValue)
	helpers.LogError(err, "unable to execute SQL statement")

	log.Println("Response data inserted into the database")

	return body
}
