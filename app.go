package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

type App struct {
	sandboxConfig SandboxConfig
	dockerCli     *client.Client
}

func NewApp(config SandboxConfig) (*App, error) {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &App{
		sandboxConfig: config,
		dockerCli:     dockerCli,
	}, nil
}

type RunContainerOpts struct {
	DisablePorts bool
}

func (a *App) RunContainer(ctx context.Context, templateName string, opts RunContainerOpts) error {
	sandboxTemplate, ok := a.sandboxConfig.Templates[templateName]
	if !ok {
		fmt.Printf("Sandbox template '%s' doesn't exists.\n", os.Args[1])
		os.Exit(1)
	}

	ping, err := a.dockerCli.Ping(ctx)
	if err != nil {
		return err
	}

	logHeader(fmt.Sprintf("Docker API Version: %s\nTemplate: %s", ping.APIVersion, sandboxTemplate.Name))

	uniqueID := uuid.NewString()[:6]
	containerName := fmt.Sprintf("dev-sandbox-%s-%s", strings.Join(strings.Split(sandboxTemplate.Name, " "), "_"), uniqueID)

	// >>>>> Create and start container
	logMessage(fmt.Sprintf("Creating container '%s'.", containerName), colorYellow)

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	if !opts.DisablePorts {
		for _, port := range sandboxTemplate.Ports {
			exposedPorts[nat.Port(port.ConatinerPort)] = struct{}{}
			portBindings[nat.Port(port.ConatinerPort)] = []nat.PortBinding{
				{
					HostIP:   "localhost",
					HostPort: port.HostPort,
				},
			}
			logMessage(fmt.Sprintf("Mapping Container Port %s to Host Port %s.", port.ConatinerPort, port.HostPort), colorYellow)
		}
	}

	container, err := a.dockerCli.ContainerCreate(ctx, &container.Config{
		Image:        sandboxTemplate.Image,
		Cmd:          sandboxTemplate.InitCommand,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"dev.sandbox.container": "true",
			"dev.sandbox.id":        uniqueID,
			"dev.sandbox.template":  sandboxTemplate.Name,
		},
		Env: sandboxTemplate.Environment,
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
	if err != nil {
		return err
	}

	logMessage(fmt.Sprintf("Starting Container '%s'.", containerName), colorYellow)

	err = a.dockerCli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		a.dockerCli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
		return err
	}

	logMessage(fmt.Sprintf("Container '%s' started successfully.", containerName), colorYellow)

	// >>>>> Render post run message using go templates.
	t, err := template.New("post_message_render").Parse(sandboxTemplate.Messages.PostStart)
	if err != nil {
		return err
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, map[string]string{
		"ContainerName": containerName,
	})
	if err != nil {
		return err
	}

	logMessage("\n"+buff.String(), colorOrgange)

	return nil
}

func (a *App) ListDevSandboxes(ctx context.Context) error {
	writer := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', tabwriter.TabIndent)

	containers, err := a.dockerCli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "dev.sandbox.template",
		}),
	})
	if err != nil {
		return err
	}

	logHeader(fmt.Sprintf("Total Sandboxes: %d", len(containers)))

	for _, c := range containers {
		templateName := c.Labels["dev.sandbox.template"]

		textStyle := lipgloss.NewStyle()

		meta := []string{
			"template:" + templateName,
		}
		metaLine := fmt.Sprintf("[%s]", strings.Join(meta, " "))

		line := strings.Join(
			[]string{
				textStyle.Foreground(colorYellow).SetString(strings.Join(c.Names, " ")).String(),
				textStyle.Foreground(colorGreen).SetString(metaLine).String(),
			}, "\t")

		writer.Write([]byte(line + "\n"))
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) PurgeDevSandboxes(ctx context.Context) error {
	conatiners, err := a.dockerCli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "dev.sandbox.template",
		}),
	})
	if err != nil {
		return err
	}

	for _, c := range conatiners {
		logMessage(fmt.Sprintf("Deleting sandbox %s", strings.Join(c.Names, "")), colorYellow)

		err := a.dockerCli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type DeleteSandboxOpts struct {
	SandboxName string
}

func (a *App) DeleteSandbox(ctx context.Context, opts DeleteSandboxOpts) error {
	if strings.TrimSpace(opts.SandboxName) == "" {
		return errors.New("provide a sandbox name to delete")
	}

	containers, err := a.dockerCli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{
				Key:   "name",
				Value: "^" + opts.SandboxName + "$",
			},
			filters.KeyValuePair{
				Key:   "label",
				Value: "dev.sandbox.container",
			},
		),
	})
	if err != nil {
		return err
	}

	if len(containers) == 0 {
		return errors.New("no container found with the provided name")
	}

	if len(containers) > 1 {
		return errors.New("multiple containers found with name")
	}

	container := containers[0]

	logMessage(fmt.Sprintf("Deleteing sandbox '%s'", strings.Join(container.Names, "")), colorYellow)

	err = a.dockerCli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}

	return nil
}
