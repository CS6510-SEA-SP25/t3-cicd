//nolint:errcheck
package DockerService

import (
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/models"
	JobService "cicd/pipeci/backend/services/job"
	PipelineService "cicd/pipeci/backend/services/pipeline"
	StageService "cicd/pipeci/backend/services/stage"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// Docker Client
type DockerClient struct {
	cli *client.Client  // Docker API Client
	ctx context.Context // Context
}

/* Initialize Docker client */
func InitDockerClient() (*DockerClient, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{cli: cli, ctx: ctx}, nil
}

/* Close the transport used by the client */
func (dc *DockerClient) Close() {
	dc.cli.Close()
}

/* Pull image from Docker hub */
func (dc *DockerClient) PullImage(imageName string) error {
	log.Printf("Installing image: %v\n", imageName)
	reader, err := dc.cli.ImagePull(dc.ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	// Copy logs to client stdout
	_, err = io.ReadAll(reader)
	// _, err = io.Copy(os.Stdout, reader)
	return err
}

/*
Create Docker container from input image and commands

	Rev #1: Define the bind mount for the workspace
	* Rev #2: Git clone inside container instead of mounting from local dir
*/
func (dc *DockerClient) CreateContainer(containerName string, imageName string, commands []string) (string, error) {
	// Combine multiple commands into a single shell command
	joinedCmd := []string{"sh", "-c", ""}
	for _, cmd := range commands {
		joinedCmd[2] += cmd + " && "
	}
	joinedCmd[2] = joinedCmd[2][:len(joinedCmd[2])-4]

	resp, err := dc.cli.ContainerCreate(dc.ctx, &container.Config{
		Image: imageName,
		Cmd:   joinedCmd,
	}, &container.HostConfig{
		// Mounts: mounts,
	}, nil, nil, "")
	if err != nil {
		return "", err
	}

	// Rename the container with the given prefix
	newName := containerName + "_" + resp.ID
	err = dc.cli.ContainerRename(dc.ctx, resp.ID, newName)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// DeleteContainer deletes a Docker container by its Id
func (dc *DockerClient) DeleteContainer(containerId string) error {
	options := container.RemoveOptions{
		Force: true, // Force removal if the container is running
	}
	err := dc.cli.ContainerRemove(dc.ctx, containerId, options)
	return err
}

/* Start container */
func (dc *DockerClient) StartContainer(containerId string) error {
	return dc.cli.ContainerStart(dc.ctx, containerId, container.StartOptions{})
}

/* Wait container */
func (dc *DockerClient) WaitContainer(containerId string) error {
	statusCh, errCh := dc.cli.ContainerWait(dc.ctx, containerId, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return err
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with non-zero status: %d", status.StatusCode)
		}
		return nil
	}
}

// Actions on Docker
func initContainer(job models.JobConfiguration, repository models.Repository) (string, error) {
	log.Printf("Running stage `%v`, job: `%v`", job.Stage.Value, job.Name.Value)

	dc, err := InitDockerClient()
	if err != nil {
		return "", err
	}
	defer dc.Close()

	if err := dc.PullImage(job.Image.Value); err != nil {
		return "", err
	}

	// Create container
	containerName := "pipeline_"

	// Checkout code from Github - checkout to specific commit hash
	var cmds []string
	cmds = append(cmds, "git clone --no-checkout "+repository.Url+" /tmp/repo")
	cmds = append(cmds, "cd /tmp/repo")
	cmds = append(cmds, "git checkout "+repository.CommitHash)
	cmds = append(cmds, job.Script.Value...)

	log.Printf("repository.Url %v", repository.Url)
	log.Printf("repository.CommitHash %v", repository.CommitHash)

	// Create container with image and commands to run at start
	containerId, err := dc.CreateContainer(containerName, job.Image.Value, cmds)
	if err != nil {
		return containerId, err
	}
	log.Printf("Container Id for job %v: %v", job.Name.Value, containerId)

	// Start container
	if err := dc.StartContainer(containerId); err != nil {
		return containerId, err
	}

	// Wait for completion
	if err := dc.WaitContainer(containerId); err != nil {
		return containerId, err
	}

	// if err := dc.GetContainerLogs(containerId); err != nil {
	// 	return err
	// }
	log.Printf("Execution done for Container Id %v", containerId)
	return containerId, nil
}

/*
Execute jobs in Docker containers
Revisions:
  - Feb 15: Linear execution
    TODO #1: Parallel execution for single-graph pipeline
		partially done, need tests
    TODO #2: Parallel execution for multiple-graphs pipeline
	TODO #3: continue-on-error
*/

func executeJob(job models.JobConfiguration, repository models.Repository) (string, error) {
	containerId, err := initContainer(job, repository)
	if err != nil {
		return containerId, fmt.Errorf("terminating pipeline execution, caused by failure in running job %#v\nCaused by: %v", job.Name.Value, err.Error())
	}
	return containerId, nil
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
func Execute(pipeline models.PipelineConfiguration, repository models.Repository) error {
	// Service instance
	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	var stageService = StageService.NewStageService(db.Instance)
	var jobService = JobService.NewJobService(db.Instance)

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
		log.Println("debug 220")
		return err
	}

	var pipelineHasFailedStages bool = false
	var pipelineHasCanceledStages bool = false

	// Execute stage
	for _, stage := range stageOrder {
		// Stage execution report
		var stageReport models.Stage = models.Stage{
			PipelineId: pipelineReportId,
			Name:       stage,
			Status:     models.PENDING,
		}
		var stageReportId, err = stageService.CreateStage(stageReport)
		if err != nil {
			log.Println("debug 237")
			return err
		}

		var stageHasFailedJobs bool = false
		var stageHasCanceledJobs bool = false

		levels := execOrder[stage]
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
						// * Update job execution status
						if containerId, err := executeJob(job, repository); err != nil {
							errCh <- JobExecResult{Job: job, Err: err} // Send result to channel
							err = jobService.UpdateJobStatusAndEndTime(jobReportId, containerId, models.FAILED)
							if err != nil {
								log.Printf("%v\n", err)
								errCh <- JobExecResult{Job: job, Err: errors.New("unexpected UpdateJobStatusAndEndTime failed")}
							}
							log.Printf("REPORT: Job `%v` run failed!\nCaused by: %v", job.Name.Value, err)
						} else {
							err = jobService.UpdateJobStatusAndEndTime(jobReportId, containerId, models.SUCCESS)
							if err != nil {
								log.Printf("%v\n", err)
								errCh <- JobExecResult{Job: job, Err: errors.New("unexpected UpdateJobStatusAndEndTime failed")}
							}
							log.Printf("REPORT: Job `%v` run success!\n", job.Name.Value)
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
					log.Println("debug 330")
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

	log.Println("debug 357")
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
