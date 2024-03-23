
// all microservices are in go and therefore its better to use structs instead of json for inner communication

/**
	1. User specifizac polling - diff jobs for diff users
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
	# A message queue that sits in between both the services
	#
	Why ?
	Allows higher scalability and performance as the Poller might need to scale horizontally
	while requests might not.
	More Pollers can be added if the queue gets filled up quickly
*/