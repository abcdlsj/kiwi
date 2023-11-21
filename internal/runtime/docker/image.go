package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/term"
)

type ImageCfg struct {
	Tags []string
	Out  io.Writer
	In   io.Reader
}

func ImageBuild(ctx context.Context, cfg *ImageCfg) error {
	cli := getDockerClient()

	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Dockerfile:     "Dockerfile",
		BuildArgs:      nil,
		Tags:           cfg.Tags,
	}

	resp, err := cli.ImageBuild(ctx, cfg.In, options)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	termFd, isTerm := term.GetFdInfo(cfg.Out)
	jsonmessage.DisplayJSONMessagesStream(resp.Body, cfg.Out, termFd, isTerm, nil)

	return nil
}

func ImageRemove(ctx context.Context, imageID string) error {
	_, err := getDockerClient().ImageRemove(context.Background(),
		imageID, types.ImageRemoveOptions{Force: true, PruneChildren: true})

	return err
}
