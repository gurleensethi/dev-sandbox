package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	ErrNotFound = errors.New("not found")
)

//go:embed sandbox-config.yaml
var config string

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Use the command line: dev-sandbox <template>\nExample: dev-sandbox react")
		os.Exit(1)
	}

	ctx := context.Background()

	sandboxConfig, err := parseSandboxFromConfig(config)
	if err != nil {
		panic(err)
	}

	app, err := NewApp(sandboxConfig)
	if err != nil {
		panic(err)
	}

	app.RunContainer(ctx)
}

func parseSandboxFromConfig(config string) (SandboxConfig, error) {
	var sandbox SandboxConfig
	decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(config)))
	err := decoder.Decode(&sandbox)
	return sandbox, err
}
