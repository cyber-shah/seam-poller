package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

/**
1. User specific polling - diff jobs for diff users
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

Why ?
Allows higher scalability and performance as the Poller might need to scale horizontally
while requests might not.
*/
type PollingRequest struct {
	userId          string
	apiEndpoint     url.URL
	pollingInterval time.Timer
}

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

func main() {
	fmt.Println("hello world")
	router := gin.Default()
	router.GET("/albums", getAlbums)

	router.Run("localhost:8080")
}
