package git

// An example git server. This will list all available repos if you ssh
// directly to the server. To test `ssh -p 23233 localhost` once it's running.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	gm "github.com/charmbracelet/wish/git"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type App struct {
	Port        int
	Host        string
	RepoDir     string
	HostKeyPath string
}

type hook struct {
	access      gm.AccessLevel
	repoRootDir string
}

func (v hook) AuthRepo(repo string, pk ssh.PublicKey) gm.AccessLevel {
	return v.access
}

func (v hook) Push(repo string, pk ssh.PublicKey) {
	r, err := git.PlainOpen(filepath.Join(v.repoRootDir, repo))
	if err != nil {
		log.Infof("PlainOpen error: %s", err)
		return
	}

	head, err := r.Head()
	if err != nil {
		log.Infof("Head error: %s", err)
		return
	}

	commit, err := r.CommitObject(head.Hash())
	if err != nil {
		log.Infof("CommitObject error: %s", err)
		return
	}

	tree, err := commit.Tree()
	if err != nil {
		log.Infof("Tree error: %s", err)
		return
	}

	tree.Files().ForEach(func(f *object.File) error {
		log.Infof("pushing %s", f.Name)
		if f.Name == "kiwi.toml" {
			content, err := f.Contents()
			if err != nil {
				log.Infof("Contents error: %s", err)
				return nil
			}

			fmt.Println(string(content))
		}
		return nil
	})
}

func (v hook) Fetch(repo string, pk ssh.PublicKey) {
	log.Info("fetch", "repo", repo)
}

func passHandler(ctx ssh.Context, password string) bool {
	return false
}

func pkHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	return true
}

func (a App) Serve() {
	s, err := wish.NewServer(
		ssh.PublicKeyAuth(pkHandler),
		ssh.PasswordAuth(passHandler),
		wish.WithAddress(fmt.Sprintf("%s:%d", a.Host, a.Port)),
		wish.WithHostKeyPath(a.HostKeyPath),
		wish.WithMiddleware(
			gm.Middleware(a.RepoDir, hook{gm.ReadWriteAccess, a.RepoDir}),
			a.handlerMiddleware,
			lm.Middleware(),
		),
	)
	if err != nil {
		log.Error("could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", a.Host, "port", a.Port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("could not stop server", "error", err)
	}
}

func (a App) handlerMiddleware(h ssh.Handler) ssh.Handler {
	return func(s ssh.Session) {
		if len(s.Command()) == 0 {
			des, err := os.ReadDir(a.RepoDir)
			if err != nil && err != fs.ErrNotExist {
				log.Error("invalid repository", "error", err)
			}
			if len(des) > 0 {
				fmt.Fprintf(s, "\n### Repo Menu ###\n\n")
			}
			for _, de := range des {
				fmt.Fprintf(s, "• %s - ", de.Name())
				fmt.Fprintf(s, "git clone ssh://%s:%d/%s\n", a.Host, a.Port, de.Name())
			}
			fmt.Fprintf(s, "\n\n### Add some repos! ###\n\n")
			fmt.Fprintf(s, "> cd some_repo\n")
			fmt.Fprintf(s, "> git remote add wish_test ssh://%s:%d/some_repo\n", a.Host, a.Port)
			fmt.Fprintf(s, "> git push wish_test\n\n\n")
		}
		h(s)
	}
}
