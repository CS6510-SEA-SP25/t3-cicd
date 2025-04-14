package queue

import (
	"cicd/pipeci/worker/cache"
	JobService "cicd/pipeci/worker/services/job"
	"cicd/pipeci/worker/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ctx = context.Background()

// RabbitMQ Item
type QueueItem struct {
	Id         string                        `json:"id"`
	JobId      int                           `json:"jobId"`
	StageId    int                           `json:"stageId"`
	PipelineId int                           `json:"pipelineId"`
	Message    types.JobExecutor_RequestBody `json:"message"`
	Dependency map[string][]string           `json:"dependency"`
}

// Checks if a job is completed
func isJobDone(jobExecutionId string, jobService *JobService.JobService) bool {
	// Get JobId in Redis
	jobId, err := cache.Get(ctx, jobExecutionId)
	if err != nil {
		log.Printf("isJobDone error get cache %v", err)
		return false
	}
	jobIdAsNum, err := strconv.Atoi(jobId)
	if err != nil {
		log.Printf("isJobDone error get converting string to int %v", err)
		return false
	}

	// Query the database for the job status
	status, err := jobService.GetJobStatus(jobIdAsNum)
	if err != nil {
		log.Printf("isJobDone error get job status %v", err)
		return false
	}

	return status == "SUCCESS"
}

// Checks if all dependencies are completed
func areDependenciesMet(dependencies []string, jobService *JobService.JobService) bool {
	for _, dep := range dependencies {
		if !isJobDone(dep, jobService) {
			return false
		}
	}
	return true
}

// Connects to RabbitMQ and returns the connection and channel.
func ConnectRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	rabbitMQURL := os.Getenv("JOB_QUEUE_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5673/"
	}

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open a channel: %v", err)
	}

	return conn, ch, nil
}

// Declares a durable queue in RabbitMQ.
func DeclareQueue(ch *amqp.Channel) (amqp.Queue, error) {
	jobQueue := os.Getenv("JOB_QUEUE_NAME")
	if jobQueue == "" {
		jobQueue = "job_queue"
	}
	q, err := ch.QueueDeclare(
		jobQueue,
		true,  // Durable
		false, // Auto-delete
		false, // Exclusive
		false, // No-wait
		nil,   // Args
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare a queue: %v", err)
	}
	return q, nil
}

// Publishes a job to RabbitMQ
func EnqueueJob(ch *amqp.Channel, queueName string, job QueueItem, jobService *JobService.JobService) error {
	for {
		if areDependenciesMet(job.Dependency[job.Id], jobService) {
			body, err := json.Marshal(job)
			if err != nil {
				return fmt.Errorf("failed to marshal task: %v", err)
			}

			err = ch.Publish(
				"",
				queueName,
				false,
				false,
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "application/json",
					Body:         body,
				},
			)
			if err != nil {
				return fmt.Errorf("failed to publish message: %v", err)
			}

			fmt.Printf("Job enqueued: %+v\n", job)
			return nil
		}
		// If dependencies are not met, retry after 10 seconds
		fmt.Printf("Dependencies not met for job: %s. Retrying in 10 seconds...\n", job.Id)
		time.Sleep(10 * time.Second)
	}
}
