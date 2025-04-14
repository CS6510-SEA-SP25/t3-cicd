//nolint:errcheck
package DockerService

import (
	"cicd/pipeci/worker/cache"
	"cicd/pipeci/worker/db"
	"cicd/pipeci/worker/models"
	"cicd/pipeci/worker/queue"
	JobService "cicd/pipeci/worker/services/job"
	PipelineService "cicd/pipeci/worker/services/pipeline"
	StageService "cicd/pipeci/worker/services/stage"
	"cicd/pipeci/worker/types"
	"context"
	"errors"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
)

/* Enqueue task into job_queue */
func enqueue(jobExecutionId string, pipelineId, stageId, jobId int,
	dependency map[string][]string, body types.JobExecutor_RequestBody,
	jobService *JobService.JobService) error {
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

	// Generate UUID as Task ID
	queueItem := queue.QueueItem{
		Id:         jobExecutionId,
		PipelineId: pipelineId,
		StageId:    stageId,
		JobId:      jobId,
		Message:    body,
		Dependency: dependency,
	}

	// dq := queue.NewDependencyQueue(ch, "job_queue")
	// dq.EnqueueWithDependencies(queueItem)

	if err := queue.EnqueueJob(ch, q.Name, queueItem, jobService); err != nil {
		log.Printf("Error enqueuing task: %v", err)
		return err
	}

	return nil
}

/* Match execution key-value pair  */
func matchExecutionIdToPipeline(executionId string, pipelineId int) {
	ctx := context.Background()
	cache.Set(ctx, executionId, pipelineId, 0)
}

/* Job Execution Result models */
type JobExecResult struct {
	Job models.JobConfiguration
	Err error
}

/*
Execute a pipeline and store reports
TODO #1: Allow failures and update status for failed jobs
TODO #2: Force stop job(s)
*/
func Execute(pipelineExecutionId string, pipeline models.PipelineConfiguration, repository models.Repository) error {
	// Service instance
	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	var stageService = StageService.NewStageService(db.Instance)
	var jobService = JobService.NewJobService(db.Instance)

	// Topological order
	var execOrder = pipeline.ExecOrder
	var stageOrder = pipeline.StageOrder

	// Jobs being canceled because their parents failed
	var terminatedJobs []string = make([]string, 0)

	// Pipeline execution report
	var pipelineReport models.Pipeline = models.Pipeline{
		Repository: removeTokenFromURL(repository.Url),
		CommitHash: repository.CommitHash,
		IPAddress:  "0.0.0.0",
		Name:       pipeline.Pipeline.Value.Name.Value,
		StageOrder: strings.Join(pipeline.StageOrder, ","),
		Status:     models.PENDING,
	}
	var pipelineReportId, err = pipelineService.CreatePipeline(pipelineReport)
	if err != nil {
		return err
	}

	// Put K-V pair to Redis
	matchExecutionIdToPipeline(pipelineExecutionId, pipelineReportId)

	var pipelineHasFailedStages bool = false
	var pipelineHasCanceledStages bool = false

	// Execute stage
	for _, stage := range stageOrder {
		// Job's execution id dependency map.
		// This map is in the exact order as job dependency map, but instead using the generated execution uuid.
		// REASON: The mapping/queueing in operator is asynchronous events -> can't check for real job id.
		var jobExecIdDependency map[string][]string = make(map[string][]string) // execId1 -> [execId2, execId3]
		var jobExecIdMap map[string]string = make(map[string]string)            // execId1 -> compile, execId2 -> build

		// Stage execution report
		var stageReport models.Stage = models.Stage{
			PipelineId: pipelineReportId,
			Name:       stage,
			Status:     models.PENDING,
		}
		var stageReportId, err = stageService.CreateStage(stageReport)
		if err != nil {
			return err
		}

		var stageHasFailedJobs bool = false
		var stageHasCanceledJobs bool = false

		levels := execOrder[stage]

		// Allocate Job Execution ID
		for _, level := range levels {
			for _, name := range level {
				// Job Async-execution id
				var jobExecutionId = "job_" + uuid.New().String()
				jobExecIdDependency[jobExecutionId] = make([]string, 0)

				var job models.JobConfiguration = *pipeline.Stages.Value[stage].Value[name]
				jobExecIdMap[job.Name.Value] = jobExecutionId

				if job.Dependencies != nil {
					for _, dep := range job.Dependencies.Value {
						jobExecIdDependency[jobExecutionId] = append(jobExecIdDependency[jobExecutionId], jobExecIdMap[dep])
					}
				}
			}
		}

		for _, level := range levels {
			/* Parallel execution */
			var wg sync.WaitGroup
			// Buffered channel to avoid blocking
			errCh := make(chan JobExecResult, len(level))
			// Execute job
			for _, name := range level {
				wg.Add(1)
				go func(name string) {
					defer wg.Done()
					var job models.JobConfiguration = *pipeline.Stages.Value[stage].Value[name]
					// Check if parent jobs is terminated
					var isTerminated bool = false
					if job.Dependencies != nil {
						for _, dep := range job.Dependencies.Value {
							if slices.Contains(terminatedJobs, dep) {
								isTerminated = true
								break
							}
						}
					}

					// Job execution report
					var jobReport models.Job = models.Job{
						StageId:     stageReportId,
						Name:        job.Name.Value,
						Image:       job.Image.Value,
						Script:      strings.Join(job.Script.Value, " && "),
						Status:      models.PENDING,
						ContainerId: "",
					}
					var jobReportId, err = jobService.CreateJob(jobReport)
					if err != nil {
						log.Printf("JobExecResult: %v", err)
						errCh <- JobExecResult{Job: job, Err: errors.New("insert job report into database failed")}
						// ? How to handle this error?
						// ? What to do if failed because of database insertion?
					}

					if isTerminated {
						stageHasCanceledJobs = true
						errCh <- JobExecResult{Job: job, Err: errors.New("parent job failed or terminated")} // Send error to channel
						err = jobService.UpdateJobStatusAndEndTime(jobReportId, "", models.CANCELED)
						if err != nil {
							log.Printf("%v\n", err)
							errCh <- JobExecResult{Job: job, Err: errors.New("unexpected UpdateJobStatusAndEndTime failed")}
						}
						log.Printf("REPORT: Job `%v` is terminated!", job.Name.Value)
					} else {

						// Execute job then send result to channel
						err := enqueue(
							jobExecIdMap[job.Name.Value],
							pipelineReportId,
							stageReportId,
							jobReportId,
							jobExecIdDependency,
							types.JobExecutor_RequestBody{
								Job:        job,
								Repository: repository,
							},
							jobService,
						)

						if err != nil {
							errCh <- JobExecResult{Job: job, Err: err} // Send result to channel
							log.Printf("REPORT: Job `%v` enqueue failed!\nCaused by: %v", job.Name.Value, err)
						} else {
							log.Printf("REPORT: Job `%v` enqueue successfully!\n", job.Name.Value)
						}
					}
				}(name)
			}

			wg.Wait()    // Wait for all goroutines to finish
			close(errCh) // Close the channel after all goroutines are done

			for result := range errCh {
				if result.Err != nil {
					terminatedJobs = append(terminatedJobs, result.Job.Name.Value)
					// stageHasFailedJobs = true // golangci-lint: ineffassign
					// ! NOTE: If no job's `allow-failure`, update stage & pipeline status before quit
					if err = stageService.UpdateStageStatusAndEndTime(stageReportId, models.FAILED); err != nil {
						log.Printf("%v\n", err)
					}
					if err = pipelineService.UpdatePipelineStatusAndEndTime(pipelineReportId, models.FAILED); err != nil {
						log.Printf("%v\n", err)
					}
					return result.Err // Return the first error encountered
				}
			}
		}

		// * Update stage execution status
		if stageHasFailedJobs {
			stageService.UpdateStageStatusAndEndTime(stageReportId, models.FAILED)
			pipelineHasFailedStages = true
		} else if stageHasCanceledJobs {
			stageService.UpdateStageStatusAndEndTime(stageReportId, models.CANCELED)
			pipelineHasCanceledStages = true
		} else {
			stageService.UpdateStageStatusAndEndTime(stageReportId, models.SUCCESS)
		}
	}

	// * Update pipeline execution status
	if pipelineHasFailedStages {
		pipelineService.UpdatePipelineStatusAndEndTime(pipelineReportId, models.FAILED)
	} else if pipelineHasCanceledStages {
		pipelineService.UpdatePipelineStatusAndEndTime(pipelineReportId, models.CANCELED)
	} else {
		pipelineService.UpdatePipelineStatusAndEndTime(pipelineReportId, models.SUCCESS)
	}

	return nil
}

// Remove Personal Access Token from URL if exists
func removeTokenFromURL(url string) string {
	// Split the URL by "@"
	parts := strings.Split(url, "@")
	if len(parts) > 1 {
		// If there's a token, return the part after "@"
		return "https://" + parts[1]
	}
	return url
}
