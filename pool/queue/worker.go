package queue

import (
	"cicd/pipeci/pool/types"
	"encoding/json"
	"fmt"
	"log"
	"os"

	// "os/exec"
	KubernetesService "cicd/pipeci/pool/k8s"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Task struct
type Task struct {
	Id      string                         `json:"id"`
	Message types.ExecuteLocal_RequestBody `json:"message"`
}

// Declares a durable queue in RabbitMQ.
func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	queue, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare a queue: %v", err)
	}
	return queue, nil
}

/*
Create worker instance
Returns an UUID to match on status command
*/
func createWorker(id int, taskChan <-chan amqp.Delivery) {
	for msg := range taskChan {
		var task Task
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("[Worker %d] Error JSON Unmarshalling: %v", id, err)
			//nolint
			msg.Nack(false, false) // Reject message without requeueing
			continue
		}

		log.Printf("[Worker %d] Processing task: %s\n", id, task.Id)

		jsonBody, err := json.Marshal(task)
		if err != nil {
			log.Fatalf("Error marshaling struct to JSON: %v", err)
		}

		// Process jobs
		/* ---------- Local execution ---------- */
		// cmd := exec.Command("../worker/worker", "--task", string(jsonBody))
		// // Capture the output
		// _, err = cmd.CombinedOutput()
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// }
		/*--------------------------------------*/

		/* ---------- Kubernetes in-cluster execution ---------- */
		KubernetesService.CreateWorkerInstance([]string{"--task", string(jsonBody)})
		/*-------------------------------------------------------*/

		msg.Ack(false) // Acknowledge successful processing
		log.Printf("[Worker %d] Task %s completed successfully", id, task.Id)
	}
}

// Process messages from RabbitMQ in the background
func Consume() {
	// get env vars
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}
	taskQueue := os.Getenv("TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "task_queue"
	}
	workerCount := 5 // Number of concurrent workers (adjustable)

	// connect
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Printf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// declare queue
	queue, err := declareQueue(ch, taskQueue)
	if err != nil {
		log.Printf("DeclareQueue Error: %v", err)
	}

	// Set QoS (Quality of Service) to distribute work fairly
	err = ch.Qos(workerCount, 0, false)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	// Start consuming messages
	msgs, err := ch.Consume(
		queue.Name,
		"",
		false, // Manual acknowledgment
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	// Create worker pool
	taskChan := make(chan amqp.Delivery, workerCount)
	for i := 1; i <= workerCount; i++ {
		go createWorker(i, taskChan)
	}

	log.Printf("Worker pool started with %d workers. Listening for tasks...", workerCount)

	// Dispatch messages to worker pool
	for msg := range msgs {
		taskChan <- msg
	}
}
