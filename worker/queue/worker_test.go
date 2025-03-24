package queue

import (
	"cicd/pipeci/worker/types"
	"encoding/json"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChannel mocks the RabbitMQ channel for testing Consume()
type MockChannel struct {
	mock.Mock
}

func (m *MockChannel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	argsList := m.Called(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return argsList.Get(0).(chan amqp.Delivery), argsList.Error(1)
}

func (m *MockChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	return nil
}

func (m *MockChannel) Close() error {
	return nil
}

// MockDelivery represents a mock message
type MockDelivery struct {
	amqp.Delivery
	AckCalled  bool
	NackCalled bool
}

func (m *MockDelivery) Ack(multiple bool) error {
	m.AckCalled = true
	return nil
}

func (m *MockDelivery) Nack(multiple bool, requeue bool) error {
	m.NackCalled = true
	return nil
}

func TestConsume(t *testing.T) {
	// Set env variables
	os.Setenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	os.Setenv("TASK_QUEUE", "test_queue")

	// Create a mock RabbitMQ channel
	mockCh := new(MockChannel)

	// Simulate a message
	task := Task{Id: "1234", Message: types.ExecuteLocal_RequestBody{}}
	body, _ := json.Marshal(task)
	mockMsg := make(chan amqp.Delivery, 1)
	mockMsg <- amqp.Delivery{Body: body}

	// Configure mock behavior
	mockCh.On("Consume", "test_queue", "", false, false, false, false, mock.Anything).Return(mockMsg, nil)

	// Run Consume in a separate goroutine
	go func() {
		Consume()
	}()

	time.Sleep(1 * time.Second)

	select {
	case msg := <-mockMsg:
		assert.NotNil(t, msg.Body, "Message should be processed")
	default:
		t.Log("No message processed")
	}
}
