package main

import (
	"context"
	"fmt"
	"os"
	"strings"

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
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &App{
		sandboxConfig: config,
		dockerCli:     cli,
	}, nil
}

func (a *App) RunContainer(ctx context.Context) {
	template, ok := a.sandboxConfig.Templates[os.Args[1]]
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
		logInfo(fmt.Sprintf("Removing existing '%s' container.", containerName))
		err := a.dockerCli.ContainerRemove(ctx, existingContainer.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			panic(err)
		}
	}

	// >>> Create and start container
	logInfo(fmt.Sprintf("Creating container '%s'.", containerName))

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
		logInfo(fmt.Sprintf("Mapping Container Port %s to Host Port %s.", port.ConatinerPort, port.HostPort))
	}

	container, err := a.dockerCli.ContainerCreate(ctx, &container.Config{
		Image:        template.Image,
		Cmd:          template.InitCommand,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			"dev-sandbox-template": template.Name,
		},
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	logInfo(fmt.Sprintf("Starting Container '%s'.", containerName))

	err = a.dockerCli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	logInfo(fmt.Sprintf("Container '%s' started successfully.", containerName))

	logAlert("\n" + template.Messages.PostStart)
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
