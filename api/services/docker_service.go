package services

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/api/types/mount"
)

var DockerClient *client.Client

func InitDocker() error {
	var err error
	DockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	_, err = DockerClient.Ping(context.Background())
	return err
}

func PullImage(ctx context.Context, imageName string) error {
	log.Printf("[Docker] Pulling image: %s", imageName)
	out, err := DockerClient.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		log.Printf("[Docker Error] Failed to pull image %s: %v", imageName, err)
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	log.Printf("[Docker] Finished pulling image: %s", imageName)
	return nil
}

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

type VolumeMapping struct {
	HostPath      string
	ContainerPath string
}

func CreateContainer(ctx context.Context, name string, img string, ports []PortMapping, volumes []VolumeMapping, cmd []string, workingDir string) (string, error) {
	log.Printf("[Docker] Creating container '%s' with image '%s'", name, img)
	// Parse ports (we only expose them to docker network, NO host port bindings)
	exposedPorts := nat.PortSet{}
	for _, p := range ports {
		if p.ContainerPort == "" {
			continue
		}
		cp := nat.Port(p.ContainerPort + "/tcp")
		exposedPorts[cp] = struct{}{}
	}

	// Parse mounts
	mounts := []mount.Mount{}
	for _, v := range volumes {
		if v.HostPath == "" || v.ContainerPath == "" {
			continue
		}
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: v.HostPath,
			Target: v.ContainerPath,
		})
	}

	config := &container.Config{
		Image:        img,
		Tty:          false,
		ExposedPorts: exposedPorts,
	}

	if len(cmd) > 0 {
		config.Cmd = cmd
	}

	if workingDir != "" {
		config.WorkingDir = workingDir
	}

	resp, err := DockerClient.ContainerCreate(ctx, config, &container.HostConfig{
		Mounts:       mounts,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}, nil, nil, name)
	if err != nil {
		log.Printf("[Docker Error] Failed to create container '%s': %v", name, err)
		return "", err
	}
	log.Printf("[Docker] Successfully created container '%s' (ID: %s)", name, resp.ID)
	return resp.ID, nil
}

func StartContainer(ctx context.Context, containerID string) error {
	log.Printf("[Docker] Starting container: %s", containerID)
	err := DockerClient.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		log.Printf("[Docker Error] Failed to start container %s: %v", containerID, err)
		return err
	}
	log.Printf("[Docker] Container started successfully: %s", containerID)
	return nil
}

func StopContainer(ctx context.Context, containerID string) error {
	log.Printf("[Docker] Stopping container: %s", containerID)
	// using background context to let docker handle timeout
	timeout := 10
	err := DockerClient.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
	if err != nil {
		log.Printf("[Docker Error] Failed to stop container %s: %v", containerID, err)
		return err
	}
	log.Printf("[Docker] Container stopped successfully: %s", containerID)
	return nil
}

func RemoveContainer(ctx context.Context, containerID string) error {
	return DockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
}

func InspectContainer(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return DockerClient.ContainerInspect(ctx, containerID)
}

func ListContainers(ctx context.Context) ([]container.Summary, error) {
	return DockerClient.ContainerList(ctx, container.ListOptions{All: true})
}

func GetContainerLogs(ctx context.Context, containerID string) (string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       "500",
	}

	out, err := DockerClient.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Use stdcopy to demultiplex the docker log stream (which multiplexes stdout/stderr)
	// into a single buffer for simple returning.
	// Since we are reading directly from the stream, if Tty is false we MUST demux it.
	stdout := new(strings.Builder)
	stderr := new(strings.Builder)

	// Check if container was created with TTY
	inspect, err := DockerClient.ContainerInspect(ctx, containerID)
	if err == nil && inspect.Config.Tty {
		io.Copy(stdout, out)
	} else {
		stdcopy.StdCopy(stdout, stderr, out)
	}

	logs := stdout.String() + stderr.String()
	return logs, nil
}
