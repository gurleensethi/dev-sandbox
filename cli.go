package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

const (
	FlagDisablePorts = "disable-ports"
	FlagOpenVSCode   = "open-vscode"
)

func BuildCli(a *App) *cli.App {
	return &cli.App{
		Name:  "dev-sandbox",
		Usage: "run predefined sandbox templates in docker containers",
		Commands: []*cli.Command{
			{
				Name:  "doctor",
				Usage: "check for all requirements on your system",
				Action: func(ctx *cli.Context) error {
					return a.Doctor(ctx.Context)
				},
			},
			{
				Name:    "list-templates",
				Aliases: []string{"lst"},
				Usage:   "list all the available templates",
				Action: func(ctx *cli.Context) error {
					return a.ListTemplates(ctx.Context)
				},
			},
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
					&cli.BoolFlag{
						Name:  FlagOpenVSCode,
						Usage: "open container code in VSCode",
						Value: false,
					},
				},
				Usage: "run a sandbox",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() == 0 {
						return fmt.Errorf("no sandbox template specified provided")
					}

					template := cCtx.Args().First()
					return a.RunSandbox(cCtx.Context, template, RunSandboxOpts{
						DisablePorts: cCtx.Bool(FlagDisablePorts),
						OpenVSCode:   cCtx.Bool(FlagOpenVSCode),
					})
				},
			},
			{
				Name:  "purge",
				Usage: "delete all running sandboxes",
				Action: func(ctx *cli.Context) error {
					return a.PurgeDevSandboxes(ctx.Context)
				},
			},
			{
				Name:    "delete",
				Usage:   "delete a sandbox",
				Aliases: []string{"rm"},
				Action: func(ctx *cli.Context) error {
					sandboxName := ctx.Args().Get(0)
					return a.DeleteSandbox(ctx.Context, DeleteSandboxOpts{
						SandboxName: sandboxName,
					})
				},
			},
		},
	}
}
