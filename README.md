
# Main Features

Uses Go, RabbitMQ and PostgreSQL

1. **User-Specific Polling Jobs:** Each user can have their own unique polling jobs, so they can get data tailored to their needs.

2. **Create Polling Jobs:** Users can create new polling jobs by providing the API endpoint and the interval at which they want the data to be fetched.

3. **Configure Polling Intervals:** The polling interval for each job can be set by the user, so they have full control over how often the data is fetched.

4. **Store API Responses:** All the responses received from the third-party APIs are stored in a database for easy access and analysis.

5. **Handle Duplicate Data:** The system uses MD5 hashing (a non-cryptographic, fast hashing algorithm) to check for duplicate data received from the APIs, ensuring that redundant information is not stored.

6. **Persistent Jobs:** Once a polling job is created, it's stored in the database, so it can be picked up and continued even if the system crashes or restarts.

# The Design

To make this polling service scalable and efficient, I went with a microservices architecture. Here's how it all fits together:

1. **Initiator Service**

   - This service handles user requests for creating new polling jobs.
   - It takes the job details (API endpoint and polling interval) and stores them in a database.
   - It publishes a message to a message queue, letting the Scheduler know about the new job.

2. **Scheduler Service**

   - The Scheduler listens for messages from the queue about new or updated polling jobs.
   - Its job is to schedule these jobs by sending messages to the Poller service at the specified intervals, telling it to fetch data from the API.

3. **Poller Service**

   - This is the workhorse of the system.
   - The Poller receives messages from the Scheduler with the API endpoint and job details.
   - It sends requests to the API, retrieves the responses, and stores them in the database.
   - Before storing the responses, the Poller calculates an MD5 hash of the response data and checks if a duplicate hash already exists in the database. This helps in preventing the storage of redundant data.

**Message Queue**

- To ensure efficient communication between Scheduler and Poller services, I used RabbitMQ as the message queue.
- This ensures reliable message delivery and processing, even in the event of temporary service disruptions or failures.

**Database**

- For storing polling jobs and API responses, I used a PostgreSQL database.
- This provides a persistent storage solution, ensuring that all data is safely stored and can be retrieved even after system restarts or crashes.

**Why this design?**

- **Scalability:** Each microservice can be scaled independently based on demand.
- **Reliability:** The message queue ensures that no job gets lost, even if one of the services goes down temporarily.
- **Separation of Concerns:** Each microservice has a specific job, making the code more modular, maintainable, and easier to test.
- **Flexibility:** Adding new polling jobs or integrating with different APIs is straightforward.
- **Monitoring and Logging:** With a distributed system, it's easier to monitor and log the health and performance of each component, providing better visibility into the overall system.

# Getting Started

All the dependencies can be found inside dependencies.txt.

To run this project:

1. Clone this repository to your local machine.
2. Navigate to `helpers/helpers.go`.
3. Update PostgreSQL and RabbitMQ configs if needed.
4. Build and run the project using the following commands:

```bash
go build .
./polling_service

