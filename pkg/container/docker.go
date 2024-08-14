package container

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerDeployer struct {
	client *client.Client
}

func NewDockerDeployer() (*DockerDeployer, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	return &DockerDeployer{client: cli}, nil
}

func (d *DockerDeployer) checkContainerNameConflict(ctx context.Context, name string) error {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return fmt.Errorf("container with name %s already exists", name)
			}
		}
	}
	return nil
}

func (d *DockerDeployer) checkPortConflicts(ctx context.Context, portMappings []PortMapping) error {
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	for _, mapping := range portMappings {
		for _, c := range containers {
			for _, p := range c.Ports {
				if p.PublicPort == uint16(nat.Port(mapping.HostPort).Int()) {
					return fmt.Errorf("port %s is already in use by container %s", mapping.HostPort, c.ID[:12])
				}
			}
		}
	}
	return nil
}

func (d *DockerDeployer) DeployService(options ServiceOptions, writer OutputWriter) error {
	ctx := context.Background()

	// Check for container name conflict
	err := d.checkContainerNameConflict(ctx, options.ContainerName)
	if err != nil {
		writer.Write("Conflict", "Container name conflict: %v", err)
		return err
	}

	// Check for port conflicts
	err = d.checkPortConflicts(ctx, options.PortMappings)
	if err != nil {
		writer.Write("Conflict", "Port conflict: %v", err)
		return err
	}

	_, _, err = d.client.ImageInspectWithRaw(ctx, options.ImageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			writer.Write("ImagePull", "Pulling image: %s", options.ImageName)
			reader, err := d.client.ImagePull(ctx, options.ImageName, types.ImagePullOptions{All: false})
			if err != nil {
				return fmt.Errorf("failed to pull image: %v", err)
			}

			io.Copy(io.Discard, reader)

			writer.Write("ImagePull", "Image pulled successfully: %s", options.ImageName)
		} else {
			return fmt.Errorf("failed to inspect image: %v", err)
		}
	} else {
		writer.Write("ImageCheck", "Image %s already exists locally", options.ImageName)
	}

	writer.Write("ContainerConfig", "Creating container configuration")
	containerConfig := &container.Config{
		Image:        options.ImageName,
		ExposedPorts: nat.PortSet{},
	}

	portBindings := nat.PortMap{}
	for _, mapping := range options.PortMappings {
		containerPort := nat.Port(mapping.ContainerPort)
		containerConfig.ExposedPorts[containerPort] = struct{}{}
		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: mapping.HostPort,
			},
		}
		writer.Write("PortMapping", "Mapping port %s to %s", mapping.HostPort, mapping.ContainerPort)
	}

	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			NanoCPUs: int64(options.CPU * 1e9),
			Memory:   options.Memory,
		},
		PortBindings: portBindings,
		NetworkMode:  container.NetworkMode(options.NetworkMode),
		AutoRemove:   options.AutoRemove,
	}

	if options.RestartPolicy != "" {
		hostConfig.RestartPolicy = container.RestartPolicy{
			Name: container.RestartPolicyMode(options.RestartPolicy),
		}
		writer.Write("RestartPolicy", "Setting restart policy to: %s", options.RestartPolicy)
	}

	for _, volume := range options.VolumeBinds {
		if volume.HostPath == "/path/to/host" || volume.HostPath == "" ||
			volume.ContainerPath == "/path/to/container" || volume.ContainerPath == "" {
			continue
		}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s", volume.HostPath, volume.ContainerPath))
		writer.Write("VolumeBinding", "Binding volume %s to %s", volume.HostPath, volume.ContainerPath)
	}

	writer.Write("ContainerCreate", "Creating container: %s", options.ContainerName)
	resp, err := d.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, options.ContainerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %v", err)
	}

	writer.Write("ContainerStart", "Starting container: %s", resp.ID)
	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}

	writer.Write("ContainerStarted", "Container %s is running", resp.ID)
	return nil
}
