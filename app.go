package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
)

var (
	MsgOpenVSCodeSteps = `> Open in VSCode:
1. Open VSCode.
2. Cmd+Shift+P
3. Search for 'Dev Containers: Attach to a running container...'
4. Select '{{.ContainerName}}' container.`
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

type RunSandboxOpts struct {
	DisablePorts bool
	OpenVSCode   bool
}

func (a *App) RunSandbox(ctx context.Context, templateName string, opts RunSandboxOpts) error {
	sandboxTemplate, ok := a.sandboxConfig.Templates[templateName]
	if !ok {
		fmt.Printf("Sandbox template '%s' doesn't exists\n", os.Args[1])
		os.Exit(1)
	}

	ping, err := a.dockerCli.Ping(ctx)
	if err != nil {
		return err
	}

	logHeader(fmt.Sprintf("Docker API Version: %s\nTemplate: %s", ping.APIVersion, sandboxTemplate.Name))

	// >>>>> Pull Image
	images, err := a.dockerCli.ImageList(ctx, types.ImageListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "reference",
			Value: sandboxTemplate.Image,
		}),
	})
	if err != nil {
		return err
	}

	if len(images) == 0 {
		logMessage(fmt.Sprintf("Pulling image %s...", sandboxTemplate.Image), colorGreen)

		reader, err := a.dockerCli.ImagePull(ctx, sandboxTemplate.Image, types.ImagePullOptions{})
		if err != nil {
			return err
		}

		_, err = io.ReadAll(reader)
		if err != nil {
			return err
		}
	}

	// >>>>> Create and start container

	uniqueID := uuid.NewString()[:6]
	containerName := fmt.Sprintf("dev-sandbox-%s-%s", strings.Join(strings.Split(sandboxTemplate.Name, " "), "_"), uniqueID)

	logMessage(fmt.Sprintf("Creating container '%s'...", containerName), colorGreen)

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
			logMessage(fmt.Sprintf("Mapping Container Port %s to Host Port %s...", port.ConatinerPort, port.HostPort), colorGreen)
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

	logMessage(fmt.Sprintf("Starting Container '%s'...", containerName), colorGreen)

	err = a.dockerCli.ContainerStart(ctx, container.ID, types.ContainerStartOptions{})
	if err != nil {
		a.dockerCli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{})
		return err
	}

	logMessage(fmt.Sprintf("Container '%s' started successfully...", containerName), colorGreen)

	templateData := map[string]string{
		"ContainerName": containerName,
	}

	vscodeSteps, err := renderTemplate(MsgOpenVSCodeSteps, templateData)
	if err != nil {
		return err
	}

	// >>>>> Open VSCode attaching the remote container
	if opts.OpenVSCode {
		if sandboxTemplate.VSCodeConfig != nil {
			logMessage("Opening container application in VSCode...", colorGreen)
			applicationPath := path.Join("/", sandboxTemplate.VSCodeConfig.ApplicationFolder)

			if hasCommand("code") == nil {
				err := exec.Command(
					"code",
					"--folder-uri",
					fmt.Sprintf("vscode-remote://attached-container+%x%s", containerName, applicationPath),
				).Run()
				if err != nil {
					return err
				}
			} else {
				logMessage("Unable to open container in VSCode, 'code' command not found!", colorYellow)
				logMessage(vscodeSteps, colorYellow)
			}
		}
	} else {
		logMessage("\n"+vscodeSteps, colorOrgange)
	}

	// >>>>> Render post run message using go templates.
	postStartMsg, err := renderTemplate(sandboxTemplate.Messages.PostStart, templateData)
	if err != nil {
		return err
	}

	logMessage("\n"+postStartMsg, colorOrgange)

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

func (a *App) ListTemplates(ctx context.Context) error {
	writer := tabwriter.NewWriter(os.Stdin, 4, 4, 4, ' ', 0)

	logHeader(fmt.Sprintf("Total Sandbox Templates: %d", len(a.sandboxConfig.Templates)))

	fmt.Fprintln(writer, "Name\tDescription")
	fmt.Fprintln(writer, "----\t-----------")

	rows := []string{}

	for key, value := range a.sandboxConfig.Templates {
		rows = append(rows, strings.Join([]string{key, value.Description}, "\t"))
	}

	sort.Strings(rows)

	for _, row := range rows {
		_, err := fmt.Fprintln(writer, row)
		if err != nil {
			return err
		}
	}

	err := writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) Doctor(ctx context.Context) error {
	commands := map[string]string{
		"docker": "Visit 'https://www.docker.com/products/docker-desktop/' to install Docker.",
		"code":   "Visit 'https://code.visualstudio.com/docs/setup/mac#_launching-from-the-command-line' to install VSCode shell command.",
	}

	for cmd, instructions := range commands {
		err := hasCommand(cmd)
		if err != nil {
			if err == ErrCommandNotFound {
				logMessage(fmt.Sprintf("Command `%s` not found! %s", cmd, instructions), colorYellow)
			} else {
				return err
			}
		} else {
			logMessage(fmt.Sprintf("Command `%s` found.", cmd), colorGreen)
		}
	}

	return nil
}
