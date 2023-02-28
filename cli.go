package main

import (
	"github.com/urfave/cli/v2"
)

const (
	FlagDisablePorts = "disable-ports"
)

func BuildCli(a *App) *cli.App {
	return &cli.App{
		Name: "dev-sandbox",
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "list all the dev sandboxes",
				Action: func(ctx *cli.Context) error {
					return a.ListDevSandboxes(ctx.Context)
				},
			},
			{
				Name:    "run",
				Aliases: []string{"r"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  FlagDisablePorts,
						Usage: "don't expose any ports on the container",
						Value: false,
					},
				},
				Usage: "run a sandbox",
				Action: func(cCtx *cli.Context) error {
					template := cCtx.Args().First()
					return a.RunContainer(cCtx.Context, template, RunContainerOpts{
						DisablePorts: cCtx.Bool(FlagDisablePorts),
					})
				},
			},
			{
				Name:  "purge",
				Usage: "delete running sandboxes",
				Action: func(ctx *cli.Context) error {
					return a.PurgeDevSandboxes(ctx.Context)
				},
			},
		},
	}
}
