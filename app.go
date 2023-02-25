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
	SandboxConfig SandboxConfig
}

func NewApp(config SandboxConfig) *App {
	return &App{
		SandboxConfig: config,
	}
}

func (a *App) RunContainer(ctx context.Context) {
	template, ok := a.SandboxConfig.Templates[os.Args[1]]
	if !ok {
		fmt.Printf("Template '%s' doesn't exists.\n", os.Args[1])
		os.Exit(1)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	ping, err := cli.Ping(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("API Version: %s\n", ping.APIVersion)

	// >>> Delete existing container
	containerName := fmt.Sprintf("dev_sandbox_%s", strings.Join(strings.Split(template.Name, " "), "_"))
	existingContainer, err := a.getContainerByName(ctx, cli, containerName)
	if err != nil && err != ErrNotFound {
		panic(err)
	}
	if err == nil {
		fmt.Printf("Removing existing '%s' container\n", containerName)
		err := cli.ContainerRemove(ctx, existingContainer.ID, types.ContainerRemoveOptions{Force: true})
		if err != nil {
			panic(err)
		}
	}

	// >>> Create and start container
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
	}

	container, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        template.Image,
		Cmd:          template.InitCommand,
		ExposedPorts: exposedPorts,
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}
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
