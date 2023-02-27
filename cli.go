package main

import (
	"github.com/urfave/cli/v2"
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
				Usage:   "run a sandbox",
				Action: func(ctx *cli.Context) error {
					template := ctx.Args().First()
					a.RunContainer(ctx.Context, template)
					return nil
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
