package main

import (
	"database/sql"
	"encoding/json"

	"polling_service/helpers"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. connect to rabbitmq and log any errors
	connection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	helpers.LogError(err, "")
	channel, _ := connection.Channel()
	queue := helpers.SetupQ(channel, "schedulerQueue")

	// 3. check if there are any jobs in the database start them first
	db := helpers.ConnectToDB()
	checkPrevious(db, channel, queue)

	// 2. start the server and create the endpoints
	router := gin.Default()
	router.Handle("POST", "/polling-jobs", func(c *gin.Context) {
		create(c, channel, queue, db)
	})

	router.Run("localhost:8080")
}

// -----------------------------------------------------------------------------------------------
// API ENDPOINT POST /POLLING-JOBS
// -----------------------------------------------------------------------------------------------
func create(c *gin.Context, channel *amqp.Channel, queue *amqp.Queue, db *sql.DB) {
	// add to DB
	_, err := db.Exec(`INSERT INTO apis_created (user_id, requested_endpoint, polling_interval) 
VALUES ($1, $2, $3)`, job.UserID, job.APIEndpoint, job.PollingInterval)

	jsonBody := helpers.ConvertToJson(c)

	// Publish the JSON message to the queue
	helpers.Publish(channel, queue, jsonBody)

	// Send a success message
	c.JSON(200, gin.H{"success": "Polling job created successfully"})
}

func checkPrevious(db *sql.DB, channel *amqp.Channel, queue *amqp.Queue) {
	rows, err := db.Query("SELECT COUNT(*) FROM apis_created")
	helpers.LogError(err, "failed to retrieve existing jobs from database")

	var activeJobs []helpers.PollingRequest
	for rows.Next() {
		var userID, apiEndpoint string
		var pollingInterval int
		rows.Scan(&userID, &apiEndpoint, &pollingInterval)
		activeJobs = append(activeJobs, helpers.PollingRequest{
			UserID:          userID,
			APIEndpoint:     apiEndpoint,
			PollingInterval: pollingInterval,
		})
	}

	for job := range activeJobs {
		jsonBody, err := json.Marshal(job)
		helpers.LogError(err, "can't convert to json")
		helpers.Publish(channel, queue, jsonBody)
	}
}
