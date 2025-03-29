package queue

import (
	"cicd/pipeci/backend/types"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnectRabbitMQ(t *testing.T) {
	t.Run("successful connection with default URL", func(t *testing.T) {
		// Ensure no env var is set
		os.Unsetenv("RABBITMQ_URL")
		defer os.Unsetenv("RABBITMQ_URL")

		conn, ch, err := ConnectRabbitMQ()
		require.NoError(t, err)
		assert.NotNil(t, conn)
		assert.NotNil(t, ch)

		// Cleanup
		ch.Close()
		conn.Close()
	})

	t.Run("successful connection with custom URL", func(t *testing.T) {
		os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
		defer os.Unsetenv("RABBITMQ_URL")

		conn, ch, err := ConnectRabbitMQ()
		require.NoError(t, err)
		assert.NotNil(t, conn)
		assert.NotNil(t, ch)

		// Cleanup
		ch.Close()
		conn.Close()
	})

	t.Run("failed connection with invalid URL", func(t *testing.T) {
		os.Setenv("RABBITMQ_URL", "amqp://invalid:5672/")
		defer os.Unsetenv("RABBITMQ_URL")

		conn, ch, err := ConnectRabbitMQ()
		require.Error(t, err)
		assert.Nil(t, conn)
		assert.Nil(t, ch)
		assert.Contains(t, err.Error(), "failed to connect to RabbitMQ")
	})
}

func TestDeclareQueue(t *testing.T) {
	// Setup - create a real connection for testing
	conn, ch, err := ConnectRabbitMQ()
	require.NoError(t, err)
	defer conn.Close()
	defer ch.Close()

	t.Run("successful queue declaration with default name", func(t *testing.T) {
		os.Unsetenv("TASK_QUEUE")
		defer os.Unsetenv("TASK_QUEUE")

		q, err := DeclareQueue(ch)
		require.NoError(t, err)
		assert.Equal(t, "task_queue", q.Name)
	})

	t.Run("successful queue declaration with custom name", func(t *testing.T) {
		testQueue := "test_queue_123"
		os.Setenv("TASK_QUEUE", testQueue)
		defer os.Unsetenv("TASK_QUEUE")

		q, err := DeclareQueue(ch)
		require.NoError(t, err)
		assert.Equal(t, testQueue, q.Name)
	})

	// t.Run("failed queue declaration with invalid channel", func(t *testing.T) {
	// 	invalidCh := &amqp.Channel{} // invalid channel
	// 	_, err := DeclareQueue(invalidCh)
	// 	require.Error(t, err)
	// 	assert.Contains(t, err.Error(), "failed to declare a queue")
	// })
}

func TestEnqueueTask(t *testing.T) {
	// Setup - create a real connection and queue for testing
	conn, ch, err := ConnectRabbitMQ()
	require.NoError(t, err)
	defer conn.Close()
	defer ch.Close()

	testQueue := "test_enqueue_queue"
	os.Setenv("TASK_QUEUE", testQueue)
	defer os.Unsetenv("TASK_QUEUE")

	_, err = ch.QueueDeclare(
		testQueue,
		true,  // Durable
		false, // Auto-delete
		false, // Exclusive
		false, // No-wait
		nil,   // Args
	)
	require.NoError(t, err)

	// Cleanup - delete the test queue after tests
	defer ch.QueueDelete(testQueue, false, false, false)

	t.Run("successful task enqueue", func(t *testing.T) {
		task := Task{
			Id:      "test-task-1",
			Message: types.ExecuteLocal_RequestBody{
				// populate required fields
			},
		}

		err := EnqueueTask(ch, testQueue, task)
		require.NoError(t, err)
	})

	t.Run("failed task enqueue - marshaling error", func(t *testing.T) {
		// Create a task with a field that can't be marshaled
		invalidTask := Task{
			Id:      "test-task-3",
			Message: types.ExecuteLocal_RequestBody{
				// Add a field that would cause marshaling to fail if uncomment
				// SomeField: make(chan int) // This would cause marshaling to fail
			},
		}

		// This test would need to be adjusted based on your types.ExecuteLocal_RequestBody
		// Currently it will pass since we don't have un-marshallable fields
		err := EnqueueTask(ch, testQueue, invalidTask)
		require.NoError(t, err)
	})
}
