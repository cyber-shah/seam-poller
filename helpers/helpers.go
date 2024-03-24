package helpers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

// ------------------------------------------------------------------------------
//
//	RABBIT MQ CONST
//
// ------------------------------------------------------------------------------
const (
	SchedulerQueueName = "schedulerQueue"
	PollerQueueName    = "pollerQueue"
	RabbitMQURL        = "amqp://guest:guest@localhost:5672/"
)

// ----------------------------------------------------------------------------------
//
//	POSTGRE SQL CONSTS
//
// ---------------------------------------------------------------------------------
const (
	host     = "localhost"
	port     = 5432
	user     = "myuser"
	password = "mypassword"
	dbname   = "mydatabase"
)

// ----------------------------------------------------------------------------------
//
// ---------------------------------------------------------------------------------
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

// ----------------------------------------------------------------------------------
// connect to the database
// ---------------------------------------------------------------------------------
func ConnectToDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	LogError(err, "cant open postgres")

	err = db.Ping()
	LogError(err, "cant open postgres")

	log.Printf("Connected to DB")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS api_responses (
		id SERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		requested_endpoint TEXT NOT NULL,
		response_body TEXT NOT NULL,
		hash TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	LogError(err, "can't create table")

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS apis_created (
		id SERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		requested_endpoint TEXT NOT NULL,
		polling_interval INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	LogError(err, "can't create table")

	return db
}
