package main

type SandboxTemplate struct {
	Name        string   `yaml:"name"`
	Image       string   `yaml:"image"`
	InitCommand []string `yaml:"initCommand"`
	Ports       []struct {
		ConatinerPort string `yaml:"containerPort"`
		HostPort      string `yaml:"hostPort"`
	} `yaml:"ports"`
}

type SandboxConfig struct {
	Templates map[string]SandboxTemplate `yaml:"templates"`
}
