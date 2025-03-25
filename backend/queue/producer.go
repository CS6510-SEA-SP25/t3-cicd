package queue

import (
	"cicd/pipeci/backend/types"
	"encoding/json"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Task struct
type Task struct {
	Id      string                         `json:"id"`
	Message types.ExecuteLocal_RequestBody `json:"message"`
}

// Connects to RabbitMQ and returns the connection and channel.
func ConnectRabbitMQ() (*amqp.Connection, *amqp.Channel, error) {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
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
	taskQueue := os.Getenv("TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "task_queue"
	}
	q, err := ch.QueueDeclare(
		taskQueue,
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

// EnqueueTask publishes a single task to RabbitMQ
func EnqueueTask(ch *amqp.Channel, queueName string, task Task) error {
	body, err := json.Marshal(task)
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

	fmt.Printf("Task enqueued: %+v\n", task)
	return nil
}
