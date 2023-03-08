package main

import (
	"errors"
	"strings"
)

type SandboxTemplate struct {
	Name         string   `yaml:"name"`
	Image        string   `yaml:"image"`
	Description  string   `yaml:"description"`
	InitCommand  []string `yaml:"initCommand"`
	VSCodeConfig *struct {
		ApplicationFolder string `yaml:"applicationFolder"`
	} `yaml:"vscodeConfig"`
	Environment []string `yaml:"environment"`
	Ports       []struct {
		ConatinerPort string `yaml:"containerPort"`
		HostPort      string `yaml:"hostPort"`
	} `yaml:"ports"`
	Messages struct {
		PostStart string `yaml:"postStart"`
	} `yaml:"messages"`
}

func (st SandboxTemplate) Validate() error {
	if strings.TrimSpace(st.Name) == "" {
		return errors.New("sandbox name cannot be empty")
	}

	if strings.TrimSpace(st.Image) == "" {
		return errors.New("sandbox image cannot be empty")
	}

	return nil
}

type SandboxConfig struct {
	Templates map[string]SandboxTemplate `yaml:"templates"`
}
