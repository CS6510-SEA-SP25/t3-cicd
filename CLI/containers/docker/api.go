package containers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	// "github.com/docker/docker/pkg/stdcopy"

	schema "cicd/pipeci/schema"
)

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

/* Create Docker container from input image and commands */
func (dc *DockerClient) CreateContainer(containerName string, imageName string, commands []string, hostWorkspaceDir string, containerWorkspaceDir string) (string, error) {
	// Combine multiple commands into a single shell command
	joinedCmd := []string{"sh", "-c", ""}
	for _, cmd := range commands {
		joinedCmd[2] += cmd + " && "
	}
	joinedCmd[2] = joinedCmd[2][:len(joinedCmd[2])-4]

	// Define the bind mount for the workspace
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: hostWorkspaceDir,      // Host directory
			Target: containerWorkspaceDir, // Container directory
		},
	}

	resp, err := dc.cli.ContainerCreate(dc.ctx, &container.Config{
		Image:      imageName,
		Cmd:        joinedCmd,
		WorkingDir: containerWorkspaceDir,
	}, &container.HostConfig{
		Mounts: mounts,
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

// DeleteContainer deletes a Docker container by its ID
func (dc *DockerClient) DeleteContainer(containerID string) error {
	options := container.RemoveOptions{
		Force: true, // Force removal if the container is running
	}
	err := dc.cli.ContainerRemove(dc.ctx, containerID, options)
	return err
}

/* Start container */
func (dc *DockerClient) StartContainer(containerID string) error {
	return dc.cli.ContainerStart(dc.ctx, containerID, container.StartOptions{})
}

/* Wait container */
func (dc *DockerClient) WaitContainer(containerID string) error {
	statusCh, errCh := dc.cli.ContainerWait(dc.ctx, containerID, container.WaitConditionNotRunning)
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

/* Retrieve container logs */
// func (dc *DockerClient) GetContainerLogs(containerID string) error {
// 	out, err := dc.cli.ContainerLogs(dc.ctx, containerID, container.LogsOptions{ShowStdout: true})
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()
// 	_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, out)
// 	return err
// }

// Actions on Docker
func initContainer(job schema.JobConfiguration) error {
	log.Printf("Running stage `%v`, job: `%v`", job.Stage.Value, job.Name.Value)
	dc, err := InitDockerClient()
	if err != nil {
		return err
	}
	defer dc.Close()

	if err := dc.PullImage(job.Image.Value); err != nil {
		return err
	}

	hostWorkspaceDir, err := os.Getwd()
	if err != nil {
		return err
	}
	containerWorkspaceDir := "/workspace"
	// containerName := fmt.Sprintf("p_%v.s_%v.j_%v", pipelineName, job.Stage.Value, job.Name.Value)
	containerName := "pipeline_"
	containerID, err := dc.CreateContainer(containerName, job.Image.Value, job.Script.Value, hostWorkspaceDir, containerWorkspaceDir)
	if err != nil {
		return err
	}
	log.Printf("Container ID for job %v: %v", job.Name.Value, containerID)

	if err := dc.StartContainer(containerID); err != nil {
		return err
	}

	if err := dc.WaitContainer(containerID); err != nil {
		return err
	}

	// if err := dc.GetContainerLogs(containerID); err != nil {
	// 	return err
	// }
	log.Printf("Execution done for Container ID %v", containerID)
	return nil
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

func executeJob(job schema.JobConfiguration) error {
	err := initContainer(job)
	if err != nil {
		return fmt.Errorf("terminating pipeline execution, caused by failure in running job %#v\nCaused by: %v", job.Name.Value, err.Error())
	}
	return nil
}

/* Job Execution Result schema */
type JobExecResult struct {
	Job schema.JobConfiguration
	Err error
}

/* Execute a pipeline */
func Execute(pipeline schema.PipelineConfiguration) error {
	execOrder := pipeline.ExecOrder
	stageOrder := pipeline.StageOrder
	var terminatedJobs []string = make([]string, 0)

	for _, stage := range stageOrder {
		levels := execOrder[stage]

		for _, level := range levels {
			/* Parallel execution */
			var wg sync.WaitGroup
			// Buffered channel to avoid blocking
			errCh := make(chan JobExecResult, len(level))
			for _, name := range level {
				wg.Add(1)
				go func(name string) {
					defer wg.Done()
					var job schema.JobConfiguration = *pipeline.Stages.Value[stage].Value[name]
					// Check if parent jobs is terminated
					var isTerminated bool = false
					if job.Dependencies != nil {
						for _, dep := range job.Dependencies.Value {
							if slices.Contains(terminatedJobs, dep) {
								errCh <- JobExecResult{Job: job, Err: errors.New("parent job failed or terminated")} // Send error to channel
								isTerminated = true
								break
							}
						}
					}

					if isTerminated {
						log.Printf("REPORT: Job `%v` is terminated!", job.Name.Value)
					} else {
						// Execute job then send result to channel
						if err := executeJob(job); err != nil {
							errCh <- JobExecResult{Job: job, Err: err} // Send result to channel
							log.Printf("REPORT: Job `%v` run failed!", job.Name.Value)
						} else {
							log.Printf("REPORT: Job `%v` run success!", job.Name.Value)
						}
					}
				}(name)
			}

			wg.Wait()    // Wait for all goroutines to finish
			close(errCh) // Close the channel after all goroutines are done

			for result := range errCh {
				if result.Err != nil {
					// return err // Return the first error encountered
					terminatedJobs = append(terminatedJobs, result.Job.Name.Value)
					return result.Err
				}
			}

			// /* Sequential execution */
			// for _, name := range level {
			// 	var job schema.JobConfiguration = *pipeline.Stages.Value[stage].Value[name]
			// 	err := executeJob(job)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
		}
	}
	return nil
}
