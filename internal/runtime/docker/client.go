package docker

import (
	"sync"

	"github.com/charmbracelet/log"
	"github.com/docker/docker/client"
)

var (
	clientMutex sync.Mutex
	cli         *client.Client
)

func getDockerClient() *client.Client {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if cli != nil {
		return cli
	}

	var err error
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v\n", err)
	}

	log.Infof("Using Docker client: %s\n", cli.DaemonHost())

	return cli
}
