package routes

import (
	containers "cicd/pipeci/backend/containers/docker"
	"cicd/pipeci/backend/models"
	"log"
	"net/http"

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

	err = containers.Execute(body.Pipeline, body.Repository)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{"success": true})
	}
	log.Print("reach here!\n")
}
