package docker

import (
	"context"
	"io"

	"github.com/charmbracelet/log"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

type ContainerCfg struct {
	Network *struct {
		HostPort      string // "8080"
		ContainerPort string // "8080/tcp"
	}

	Image string
}

func ContinerCreate(ctx context.Context, cfg *ContainerCfg) (string, error) {
	cli := getDockerClient()

	hostCfg := &container.HostConfig{}

	if cfg.Network != nil {
		hostCfg.PortBindings = nat.PortMap{
			nat.Port(cfg.Network.ContainerPort): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: cfg.Network.HostPort,
				},
			},
		}
	}

	containerCfg := &container.Config{
		Image: cfg.Image,
	}

	contain, err := cli.ContainerCreate(ctx, containerCfg, hostCfg, nil, nil, "")
	if err != nil {
		log.Fatalf("Failed to create container: %s", err)
	}

	cli.ContainerStart(ctx, contain.ID, types.ContainerStartOptions{})
	log.Infof("Container %s is started", contain.ID)

	return contain.ID, nil
}

func ContainerAttach(ctx context.Context, containerID string, w io.Writer) error {
	cli := getDockerClient()

	cli.ContainerAttach(ctx, containerID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
		Stdin:  true,
		Logs:   true,
	})

	return nil
}
