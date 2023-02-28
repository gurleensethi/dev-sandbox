package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type App struct {
	sandboxConfig SandboxConfig
	dockerCli     *client.Client
}

func NewApp(config SandboxConfig) (*App, error) {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &App{
		sandboxConfig: config,
		dockerCli:     dockerCli,
	}, nil
}

func (a *App) RunContainer(ctx context.Context, templateName string) {
	template, ok := a.sandboxConfig.Templates[templateName]
	if !ok {
		fmt.Printf("Template '%s' doesn't exists.\n", os.Args[1])
		os.Exit(1)
	}

	ping, err := a.dockerCli.Ping(ctx)
	if err != nil {
		panic(err)
	}

	logHeader(fmt.Sprintf("Docker API Version: %s\nTemplate: %s", ping.APIVersion, template.Name))

	// >>> Delete existing container
	containerName := fmt.Sprintf("dev_sandbox_%s", strings.Join(strings.Split(template.Name, " "), "_"))
	existingContainer, err := a.getContainerByName(ctx, a.dockerCli, containerName)
	if err != nil && err != ErrNotFound {
		panic(err)
	}
	if err == nil {
		logMessage(fmt.Sprintf("Removing existing '%s' container.", containerName), colorYellow)
		err := a.dockerCli.ContainerRemove(ctx, existingContainer.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			panic(err)
		}
	}

	// >>> Create and start container
	logMessage(fmt.Sprintf("Creating container '%s'.", containerName), colorYellow)

	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	for _, port := range template.Ports {
		exposedPorts[nat.Port(port.ConatinerPort)] = struct{}{}
		portBindings[nat.Port(port.ConatinerPort)] = []nat.PortBinding{
			{
				HostIP:   "localhost",
				HostPort: port.HostPort,
			},
		}
		logMessage(fmt.Sprintf("Mapping Container Port %s to Host Port %s.", port.ConatinerPort, port.HostPort), colorYellow)
	}

	container, err := a.dockerCli.ContainerCreate(ctx, &container.Config{
		Image:        template.Image,
		Cmd:          template.InitCommand,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"dev.sandbox.template": template.Name,
		},
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	logMessage(fmt.Sprintf("Starting Container '%s'.", containerName), colorYellow)

	err = a.dockerCli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	logMessage(fmt.Sprintf("Container '%s' started successfully.", containerName), colorYellow)

	logMessage("\n"+template.Messages.PostStart, colorOrgange)
}

func (a *App) getContainerByName(ctx context.Context, cli *client.Client, name string) (types.Container, error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: name,
		}),
	})
	if err != nil {
		panic(err)
	}

	if len(containers) == 0 {
		return types.Container{}, ErrNotFound
	}

	return containers[0], nil
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
		logMessage(fmt.Sprintf("Removing container %s", strings.Join(c.Names, " ")), colorYellow)

		err := a.dockerCli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
