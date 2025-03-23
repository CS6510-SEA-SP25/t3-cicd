package routes

import (
	"log"
	"net/http"

	// DockerService "cicd/pipeci/backend/containers/docker"
	queue "cicd/pipeci/backend/queue"
	types "cicd/pipeci/backend/types"

	"github.com/gin-gonic/gin"
)

func enqueue(body types.ExecuteLocal_RequestBody) error {
	// Connect to RabbitMQ
	conn, ch, err := queue.ConnectRabbitMQ()
	if err != nil {
		return err
	}
	defer conn.Close()
	defer ch.Close()

	// Declare the queue
	q, err := queue.DeclareQueue(ch)
	if err != nil {
		return err
	}

	task := queue.Task{Id: 1, Message: types.ExecuteLocal_RequestBody(body)}
	if err := queue.EnqueueTask(ch, q.Name, task); err != nil {
		log.Printf("Error enqueuing task: %v", err)
		return err
	}

	return nil
}

/* Execute pipeline for local repo */
func ExecuteLocal(c *gin.Context) {
	var body types.ExecuteLocal_RequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}

	err = enqueue(body)

	// err = DockerService.Execute(body.Pipeline, body.Repository)
	if err != nil {
		log.Printf("ExecuteLocal %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
	}
	log.Print("reach here!\n")
}
