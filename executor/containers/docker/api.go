//nolint:errcheck
package DockerService

import (
	"bytes"
	"cicd/pipeci/executor/cache"
	"cicd/pipeci/executor/db"
	"cicd/pipeci/executor/models"
	JobService "cicd/pipeci/executor/services/job"
	PipelineService "cicd/pipeci/executor/services/pipeline"
	StageService "cicd/pipeci/executor/services/stage"
	"cicd/pipeci/executor/storage"
	"context"
	"fmt"
	"io"
	"log"
	"os"

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
func initDockerClient() (*DockerClient, error) {
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
func (dc *DockerClient) pullImage(imageName string) error {
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
func (dc *DockerClient) createContainer(containerName string, imageName string, commands []string) (string, error) {
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

// deleteContainer deletes a Docker container by its Id
func (dc *DockerClient) deleteContainer(containerId string) error {
	options := container.RemoveOptions{
		Force: true, // Force removal if the container is running
	}
	err := dc.cli.ContainerRemove(dc.ctx, containerId, options)
	return err
}

/* Start container */
func (dc *DockerClient) startContainer(containerId string) error {
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

/* Retrieve container logs as byte stream for easier MinIO upload */
func (dc *DockerClient) getContainerLogs(containerID string) (*bytes.Buffer, error) {
	log.Printf("START getContainerLogs")
	// Get container logs
	out, err := dc.cli.ContainerLogs(dc.ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	defer out.Close()

	// Read logs into a buffer
	var logBuffer bytes.Buffer
	_, err = io.Copy(&logBuffer, out)
	if err != nil {
		return nil, fmt.Errorf("failed to read container logs: %w", err)
	}

	return &logBuffer, nil
}

// Actions on Docker
func (dc *DockerClient) initContainer(job models.JobConfiguration, repository models.Repository) (string, error) {
	log.Printf("Running stage `%v`, job: `%v`", job.Stage.Value, job.Name.Value)

	if err := dc.pullImage(job.Image.Value); err != nil {
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

	// Create container with image and commands to run at start
	containerId, err := dc.createContainer(containerName, job.Image.Value, cmds)
	if err != nil {
		return containerId, err
	}
	log.Printf("Container Id for job %v: %v", job.Name.Value, containerId)

	// Start container
	if err := dc.startContainer(containerId); err != nil {
		return containerId, err
	}

	// Wait for completion
	if err := dc.WaitContainer(containerId); err != nil {
		return containerId, err
	}

	// Done
	log.Printf("Execution done for Container Id %v", containerId)
	return containerId, nil
}

/* Take actions after executions: get logs, upload to Minio, then delete containers */
func (dc *DockerClient) handlePostExecution(containerId string) error {
	log.Printf("START handlePostExecution")
	// Retrieve container logs
	containerLogs, err := dc.getContainerLogs(containerId)
	if err != nil {
		return err
	}

	// Upload logs to MinIO
	var minioBucket string = os.Getenv("DEFAULT_BUCKET")
	err = storage.UploadLogsToMinIO(minioBucket, fmt.Sprintf("containers/%v", containerId), containerLogs)
	if err != nil {
		return err
	}

	// Delete container after saving logs
	dc.deleteContainer(containerId)

	// Done
	log.Printf("handlePostExecution done for Container Id %v", containerId)
	return nil
}

/* Match execution key-value pair  */
func matchExecutionIdToJob(executionId string, jobId int) {
	ctx := context.Background()
	cache.Set(ctx, executionId, jobId, 0)
}

/*
Execute jobs in Docker containers
Revisions:
  - Feb 15: Linear execution
    TODO #1: Parallel execution for single-graph pipeline
    TODO #2: Parallel execution for multiple-graphs pipeline
    TODO #3: continue-on-error
*/
func executeJob(job models.JobConfiguration, repository models.Repository) (string, error) {
	log.Printf("START executeJob")
	dc, err := initDockerClient()
	if err != nil {
		return "", err
	}
	defer dc.Close()

	containerId, initErr := dc.initContainer(job, repository)
	postExecErr := dc.handlePostExecution(containerId)

	// If both initContainer and handlePostExecution fail, combine errors
	if initErr != nil && postExecErr != nil {
		return containerId, fmt.Errorf(
			"terminating pipeline execution, caused by failure in running job %#v\nCaused by: %v\nAdditionally, post-execution failed: %v",
			job.Name.Value, initErr.Error(), postExecErr.Error(),
		)
	}

	// Prioritize initErr
	if initErr != nil {
		return containerId, fmt.Errorf("terminating pipeline execution, caused by failure in running job %#v\nCaused by: %v",
			job.Name.Value, initErr.Error(),
		)
	}

	// postExecErr
	if postExecErr != nil {
		return containerId, fmt.Errorf("post-execution failed for container %s: %v", containerId, postExecErr)
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
func Execute(pipelineReportId, stageReportId, jobReportId int, executionId string, job models.JobConfiguration, repository models.Repository) error {
	// Service instance
	var pipelineService = PipelineService.NewPipelineService(db.Instance)
	var stageService = StageService.NewStageService(db.Instance)
	var jobService = JobService.NewJobService(db.Instance)

	// Put K-V pair to Redis
	matchExecutionIdToJob(executionId, pipelineReportId)

	// * Execute job and update job/stage/pipeline execution status
	if containerId, err := executeJob(job, repository); err != nil {
		err = jobService.UpdateJobStatusAndEndTime(jobReportId, containerId, models.FAILED)
		if err != nil {
			log.Printf("%v\n", err)
		}
		log.Printf("REPORT: Job `%v` run failed!\nCaused by: %v", job.Name.Value, err)
	} else {
		if err = jobService.UpdateJobStatusAndEndTime(jobReportId, containerId, models.SUCCESS); err != nil {
			log.Printf("%v\n", err)
		}
		if err = stageService.UpdateStageStatusAndEndTime(stageReportId, models.FAILED); err != nil {
			log.Printf("%v\n", err)
		}
		if err = pipelineService.UpdatePipelineStatusAndEndTime(pipelineReportId, models.FAILED); err != nil {
			log.Printf("%v\n", err)
		}
		log.Printf("REPORT: Job `%v` run success!\n", job.Name.Value)
	}

	return nil
}

// Remove Personal Access Token from URL if exists
// func removeTokenFromURL(url string) string {
// 	// Split the URL by "@"
// 	parts := strings.Split(url, "@")
// 	if len(parts) > 1 {
// 		// If there's a token, return the part after "@"
// 		return "https://" + parts[1]
// 	}
// 	return url
// }
