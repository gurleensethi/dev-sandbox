package main

import (
	"bytes"
	_ "embed"
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	ErrNotFound = errors.New("not found")
)

//go:embed sandbox-config.yaml
var config string

func main() {
	sandboxConfig, err := parseSandboxFromConfig(config)
	if err != nil {
		panic(err)
	}

	app, err := NewApp(sandboxConfig)
	if err != nil {
		panic(err)
	}

	err = BuildCli(app).Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func parseSandboxFromConfig(config string) (SandboxConfig, error) {
	var sandbox SandboxConfig
	decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(config)))
	err := decoder.Decode(&sandbox)
	return sandbox, err
}
