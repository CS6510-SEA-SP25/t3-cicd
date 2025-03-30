package routes

import (
	"cicd/pipeci/backend/cache"
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"
	JobService "cicd/pipeci/backend/services/job"
	PipelineService "cicd/pipeci/backend/services/pipeline"
	StageService "cicd/pipeci/backend/services/stage"
	"cicd/pipeci/backend/types"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	pipelineService *PipelineService.PipelineService
	stageService    *StageService.StageService
	jobService      *JobService.JobService
)

/* Get Pipeline Execution status */
func getExecutionStatus(pipeline models.Pipeline) (types.RequestExecutionStatus_ResponseBody, error) {
	var response types.RequestExecutionStatus_ResponseBody

	// set pipeline
	response.Pipeline = types.PipelineExecutionStatus{
		PipelineId: pipeline.PipelineId,
		Name:       pipeline.Name,
		Status:     string(pipeline.Status),
		StageOrder: pipeline.StageOrder,
	}

	// set stages and jobs
	response.Stages = make(map[string]types.StageExecutionStatus)
	var stageFilters = map[string]interface{}{
		"pipeline_id": pipeline.PipelineId,
	}
	// query stages
	stages, _ := stageService.QueryStages(stageFilters)
	for _, stage := range stages {
		// query jobs
		var jobFilters = map[string]interface{}{
			"stage_id": stage.StageId,
		}
		jobs, _ := jobService.QueryJobs(jobFilters)

		var jobResponse = make([]types.JobExecutionStatus, len(jobs))
		for i, job := range jobs {
			jobResponse[i] = types.JobExecutionStatus{
				JobId:  job.JobId,
				Name:   job.Name,
				Status: string(job.Status),
			}
		}

		response.Stages[stage.Name] = types.StageExecutionStatus{
			StageId: stage.StageId,
			Name:    stage.Name,
			Status:  string(stage.Status),
			Jobs:    jobResponse,
		}
	}
	return response, nil
}

/* Return API error */
func requestExecutionStatusError(c *gin.Context, err error) {
	log.Printf("RequestExecutionStatus %v", err)
	c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
}

/* Get pipeline execution status */
func RequestExecutionStatus(c *gin.Context) {
	pipelineService = PipelineService.NewPipelineService(db.Instance)
	stageService = StageService.NewStageService(db.Instance)
	jobService = JobService.NewJobService(db.Instance)

	var ctx context.Context = context.Background()
	var pipelineId string
	var body types.RequestExecutionStatus_RequestBody

	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}
	pipelineId, err = cache.Get(ctx, body.ExecutionId)
	if err != nil {
		requestExecutionStatusError(c, err)
		return
	}

	// Get pipeline by id
	var pipelineFilters = map[string]interface{}{
		"pipeline_id": pipelineId,
	}
	// query pipeline
	pipelines, err := pipelineService.QueryPipelines(pipelineFilters)
	if err != nil {
		requestExecutionStatusError(c, err)
	} else if len(pipelines) != 1 {
		requestExecutionStatusError(c, fmt.Errorf("there must exists only one pipeline of id %v", pipelineId))
	} else {
		response, err := getExecutionStatus(pipelines[0])
		if err != nil {
			requestExecutionStatusError(c, err)
		} else {
			c.IndentedJSON(http.StatusOK, response)
		}
	}
}
