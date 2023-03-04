package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

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
		if os.Getenv("PANIC_ON_ERROR") == "1" {
			panic(err)
		}

		errStr := err.Error()
		errStr = strings.ToUpper(errStr[:1]) + errStr[1:]
		if errStr[len(errStr)-1] != '.' {
			errStr += "."
		}

		fmt.Println(errStr)
		os.Exit(1)
	}
}

func parseSandboxFromConfig(config string) (SandboxConfig, error) {
	var sandbox SandboxConfig
	decoder := yaml.NewDecoder(bytes.NewBuffer([]byte(config)))
	err := decoder.Decode(&sandbox)
	return sandbox, err
}
