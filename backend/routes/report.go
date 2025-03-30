package routes

import (
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"
	JobService "cicd/pipeci/backend/services/job"
	PipelineService "cicd/pipeci/backend/services/pipeline"
	StageService "cicd/pipeci/backend/services/stage"
	"cicd/pipeci/backend/types"
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* Report all local pipeline executions */
func ReportPastExecutionsLocal_CurrentRepo(c *gin.Context) {
	var body types.ReportPastExecutionsLocal_CurrentRepo_RequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}

	// SQL Filter
	pipelineFilters := map[string]interface{}{"repository": body.Repository.Url, "ip_address": body.IPAddress}

	// Add commit hash matching if specified
	if body.Repository.CommitHash != "" {
		pipelineFilters["commit_hash"] = body.Repository.CommitHash
	}

	gatherPipelineReport(c, pipelineFilters)
	log.Println("ReportPastExecutionsLocal_CurrentRepo: Done report summary!")
}

/* Query local executions by conditions */
func ReportPastExecutionsLocal_ByCondition(c *gin.Context) {
	var body types.ReportPastExecutionsLocal_CurrentRepo_RequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		return
	}

	// SQL Filter
	pipelineFilters := map[string]interface{}{"repository": body.Repository.Url, "ip_address": body.IPAddress}
	stageFilters := map[string]interface{}{}
	jobFilters := map[string]interface{}{}

	/*
		Represent which execution components to be reported
		0: Pipeline (default)
		1: Stage
		2: Job
	*/
	var reportOption int = 0

	// Add conditions if specified
	// CommitHash
	if body.Repository.CommitHash != "" {
		pipelineFilters["commit_hash"] = body.Repository.CommitHash
	}
	// Pipeline name
	if body.PipelineName != "" {
		pipelineFilters["name"] = body.PipelineName
	}
	// Stage name
	if body.StageName != "" {
		stageFilters["name"] = body.StageName
		reportOption = 1
	}
	// Job name
	if body.JobName != "" {
		jobFilters["name"] = body.JobName
		reportOption = 2
	}

	// Gather report based on reportOption
	if reportOption == 0 { // Pipeline
		gatherPipelineReport(c, pipelineFilters)
	} else if reportOption == 1 { // Stage
		gatherStageReport(c, pipelineFilters, stageFilters)
	} else { // Job
		gatherJobReport(c, pipelineFilters, stageFilters, jobFilters)
	}

	log.Println("ReportPastExecutionsLocal_ByCondition: Done report summary!")
}

// ------------ QUERY DATABASE FOR EXECUTION REPORTS ---------------- //
/* Get execution reports for pipeline */
func gatherPipelineReport(c *gin.Context, pipelineFilters map[string]interface{}) {
	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	var reports []types.Report_ResponseBody

	pipelines, err := pipelineService.QueryPipelines(pipelineFilters)
	parsePipelineReports(pipelines, &reports)

	if err != nil {
		log.Printf("gatherPipelineReport %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
	} else {
		c.IndentedJSON(http.StatusOK, reports)
	}
}

/* Get execution reports for stage */
func gatherStageReport(c *gin.Context, pipelineFilters, stageFilters map[string]interface{}) {
	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	var stageService = StageService.NewStageService(db.Instance)
	var reports []types.Report_ResponseBody
	var err error

	// Get PipelineId then GetStagesByPipelineId
	pipelines, err := pipelineService.QueryPipelines(pipelineFilters)
	if err != nil {
		log.Printf("gatherStageReport %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
	}

	// Append MATCHING stages to reports
	for _, pipeline := range pipelines {
		var pipelineId int = pipeline.PipelineId
		stageFilters["pipeline_id"] = pipelineId
		stages, err := stageService.QueryStages(stageFilters)
		if err != nil {
			log.Printf("gatherStageReport %v", err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
		} else {
			parseStageReports(stages, &reports)
		}
	}

	// return
	c.IndentedJSON(http.StatusOK, reports)
}

/* Get execution reports for job */
func gatherJobReport(c *gin.Context, pipelineFilters, stageFilters, jobFilters map[string]interface{}) {
	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	var stageService = StageService.NewStageService(db.Instance)
	var jobService = JobService.NewJobService(db.Instance)
	var reports []types.Report_ResponseBody
	var err error

	// Get PipelineId then GetStagesByPipelineId
	pipelines, err := pipelineService.QueryPipelines(pipelineFilters)
	if err != nil {
		log.Printf("gatherStageReport %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
	}

	// Get StageId then QueryJobs
	for _, pipeline := range pipelines {
		stageFilters["pipeline_id"] = pipeline.PipelineId
		stages, err := stageService.QueryStages(stageFilters)
		if err != nil {
			log.Printf("gatherStageReport %v", err)
			c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
		} else {
			// Append MATCHING jobs to reports
			for _, stage := range stages {
				jobFilters["stage_id"] = stage.StageId
				jobs, err := jobService.QueryJobs(jobFilters)
				if err != nil {
					log.Printf("gatherJobReport %v", err)
					c.IndentedJSON(http.StatusBadRequest, gin.H{"success": false})
				} else {
					parseJobReports(jobs, &reports)
				}
			}
		}
	}

	// return
	c.IndentedJSON(http.StatusOK, reports)
}

// ------------ PARSE DATABASE SCHEMA TO REPORTS ---------------- //
/* Parse pipelines and append to general reports list */
func parsePipelineReports(pipelines []models.Pipeline, reports *[]types.Report_ResponseBody) {
	for _, pipeline := range pipelines {
		report := types.Report_ResponseBody{
			Id:        pipeline.PipelineId,
			Name:      pipeline.Name,
			StartTime: pipeline.StartTime,
			Status:    string(pipeline.Status),
		}
		// EndTime
		if pipeline.EndTime.Valid {
			report.EndTime = pipeline.EndTime
		} else {
			report.EndTime = sql.NullTime{Valid: false}
		}
		*reports = append(*reports, report)
	}
}

/* Parse stages and append to general reports list */
func parseStageReports(stages []models.Stage, reports *[]types.Report_ResponseBody) {
	for _, stage := range stages {
		report := types.Report_ResponseBody{
			Id:        stage.StageId,
			Name:      stage.Name,
			StartTime: stage.StartTime,
			Status:    string(stage.Status),
		}
		// EndTime
		if stage.EndTime.Valid {
			report.EndTime = stage.EndTime
		} else {
			report.EndTime = sql.NullTime{Valid: false}
		}
		*reports = append(*reports, report)
	}
}

/* Parse jobs and append to general reports list */
func parseJobReports(jobs []models.Job, reports *[]types.Report_ResponseBody) {
	for _, job := range jobs {
		report := types.Report_ResponseBody{
			Id:        job.JobId,
			Name:      job.Name,
			StartTime: job.StartTime,
			Status:    string(job.Status),
		}
		// EndTime
		if job.EndTime.Valid {
			report.EndTime = job.EndTime
		} else {
			report.EndTime = sql.NullTime{Valid: false}
		}
		*reports = append(*reports, report)
	}
}
