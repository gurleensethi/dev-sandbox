package main

type SandboxTemplate struct {
	Name        string   `yaml:"name"`
	Image       string   `yaml:"image"`
	InitCommand []string `yaml:"initCommand"`
	Ports       []struct {
		ConatinerPort string `yaml:"containerPort"`
		HostPort      string `yaml:"hostPort"`
	} `yaml:"ports"`
	Messages struct {
		PostStart string `yaml:"postStart"`
	} `yaml:"messages"`
}

type SandboxConfig struct {
	Templates map[string]SandboxTemplate `yaml:"templates"`
}
