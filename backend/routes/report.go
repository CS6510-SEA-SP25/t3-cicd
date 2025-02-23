package routes

import (
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"
	PipelineService "cicd/pipeci/backend/services/pipeline"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReportPastExecutionsLocal_RequestBody struct {
	Repository   models.Repository `json:"repository"`
	IPAddress    string            `json:"ip_address"`
	PipelineName string            `json:"pipeline_name"`
}

/* Report all local pipeline executions */
func ReportPastExecutionsLocal(c *gin.Context) {
	var body ReportPastExecutionsLocal_RequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}

	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	filters := map[string]interface{}{"repository": body.Repository.Url, "ip_address": body.IPAddress}

	// Add commit hash matching if specified
	if body.Repository.CommitHash != "" {
		filters["commit_hash"] = body.Repository.CommitHash
	}

	pipelines, err := pipelineService.QueryPipelines(filters)

	if err != nil {
		log.Printf("ReportPastExecutionsLocal %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
	} else {
		c.IndentedJSON(http.StatusOK, pipelines)
	}
	log.Print("reach here!\n")
}

/* Query local pipeline executions by conditions */
func QueryPastExecutionsLocal(c *gin.Context) {
	var body ReportPastExecutionsLocal_RequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}

	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	filters := map[string]interface{}{"repository": body.Repository.Url, "ip_address": body.IPAddress}

	// Add conditions if specified
	if body.Repository.CommitHash != "" {
		filters["commit_hash"] = body.Repository.CommitHash
	}
	if body.PipelineName != "" {
		filters["name"] = body.PipelineName
	}

	pipelines, err := pipelineService.QueryPipelines(filters)

	if err != nil {
		log.Printf("ReportPastExecutionsLocal %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
	} else {
		c.IndentedJSON(http.StatusOK, pipelines)
	}
	log.Print("reach here!\n")
}
