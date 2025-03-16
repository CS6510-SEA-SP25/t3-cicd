package routes

import (
	"cicd/pipeci/backend/models"
	"log"
	"net/http"

	DockerService "cicd/pipeci/backend/containers/docker"

	"github.com/gin-gonic/gin"
)

type ExecuteLocal_RequestBody struct {
	Pipeline   models.PipelineConfiguration `json:"pipeline"`
	Repository models.Repository            `json:"repository"`
}

/* Execute pipeline for local repo */
func ExecuteLocal(c *gin.Context) {
	var body ExecuteLocal_RequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}

	err = DockerService.Execute(body.Pipeline, body.Repository)
	if err != nil {
		log.Printf("ExecuteLocal %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "error": err})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
	}
	log.Print("reach here!\n")
}
