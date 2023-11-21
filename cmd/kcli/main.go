package main

import (
	"context"
	"io"
	"os"

	"github.com/abcdlsj/kiwi/internal/runtime/docker"
	"github.com/abcdlsj/kiwi/internal/tarball"
	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "kcli",
		Usage: "kcli",
		Commands: []*cli.Command{
			{
				Name:  "tar",
				Usage: "tar",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "dir",
						Aliases:  []string{"d"},
						Usage:    "dir",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "out",
						Aliases:  []string{"o"},
						Usage:    "out",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					out := c.String("out")
					dir := c.String("dir")

					outw, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
					if err != nil {
						log.Errorf("Failed to open %s: %s", out, err)
						return err
					}

					defer outw.Close()

					r, err := tarball.TarDir(context.Background(), dir)
					if err != nil {
						log.Errorf("Failed to tar %s: %s", dir, err)
						return err
					}

					if _, err := io.Copy(outw, r); err != nil {
						log.Errorf("Failed to copy response: %s", err)
						return err
					}

					return nil
				},
			},
			{
				Name:  "runtime",
				Usage: "runtime",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "type",
						Usage:       "type",
						DefaultText: "docker",
					},
					&cli.BoolFlag{
						Name:  "quiet",
						Usage: "quiet",
						Value: false,
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:  "build",
						Usage: "build",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "in",
								Aliases: []string{"i"},
								Usage:   "in",
							},
							&cli.StringFlag{
								Name:  "image",
								Usage: "image",
							},
						},
						Action: func(c *cli.Context) error {
							in := c.String("in")
							if in == "" {
								log.Errorf("input is required")
								return nil
							}

							r, err := os.Open(in)
							if err != nil {
								log.Errorf("Failed to open %s: %s", in, err)
								return err
							}

							defer r.Close()

							cfg := &docker.ImageCfg{
								In:  r,
								Out: os.Stdout,
							}

							if c.Bool("quiet") {
								cfg.Out = io.Discard
							}

							image := c.String("image")
							if image != "" {
								cfg.Tags = []string{image}
							}

							err = docker.ImageBuild(context.Background(), cfg)
							if err != nil {
								log.Errorf("Failed to build image: %s", err)
								return err
							}

							log.Infof("Built image %s successfully!", image)

							return nil
						},
					},

					{
						Name:  "remove",
						Usage: "remove",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "image",
								Usage:    "image",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							image := c.String("image")

							err := docker.ImageRemove(context.Background(), image)
							if err != nil {
								log.Errorf("Failed to remove image %s: %s", image, err)
								return err
							}

							log.Infof("Removed image %s successfully!", image)
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
