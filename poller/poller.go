package main

// TODO: logging everywhere

import (
	// "log"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"polling_service/helpers"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

// TODO: is there a better way to do this??
const (
	host     = "localhost"
	port     = 5432
	user     = "myuser"
	password = "mypassword"
	dbname   = "mydatabase"
)

func main() {
	// Connect to RabbitMQ server
	conn, _ := amqp.Dial(helpers.RabbitMQURL)
	// Create a channel
	ch, _ := conn.Channel()

	db := connectToDB()

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
		fmt.Println("Duplicate data detected, skipping insertion")
		return body
	}

	statement, err := db.Prepare("INSERT INTO api_responses (user_id, response_body) VALUES ($1, $2)")
	// _, err = statement.Exec(userID, body)
	helpers.LogError(err, "unable to execute statement")

	return body
}

func connectToDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	helpers.LogError(err, "cant open postgres")

	err = db.Ping()
	helpers.LogError(err, "cant open postgres")

	log.Printf("Connected to DB")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS api_responses (
		id SERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		requested_endpoint TEXT NOT NULL,
		response_body TEXT NOT NULL,
		hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	helpers.LogError(err, "can't create table")

	return db
}
