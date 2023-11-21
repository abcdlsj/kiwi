package main

import (
	"github.com/abcdlsj/kiwi/internal/git"
)

func main() {
	gitServer := git.App{
		Port:        22223,
		Host:        "localhost",
		RepoDir:     ".repos",
		HostKeyPath: ".ssh/kiwi_git_server_ed25519",
	}

	gitServer.Serve()
}
