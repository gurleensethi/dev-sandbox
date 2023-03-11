package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	configFolderPath    = flag.String("config-folder", "", "Path to read config from")
	templateFileOutPath = flag.String("config-file-out", "", "Path to write the config file")
)

func main() {
	flag.Parse()

	if *configFolderPath == "" {
		panic("Config Folder Path is required")
	}

	if *templateFileOutPath == "" {
		panic("Config File Out path is required")
	}

	fmt.Println("Sandbox config folder path:", *configFolderPath)
	fmt.Println("Template out file:", *templateFileOutPath)

	// Represent YAML:
	// templates:
	//		template-key:
	//			<config>
	sandboxTemplates := make(map[string]map[string]any)
	sandboxTemplates["templates"] = make(map[string]any)

	outFile, err := os.Stat(*templateFileOutPath)
	if err != nil {
		panic(err)
	}

	if outFile.IsDir() {
		panic("Out file must be a file.")
	}

	err = filepath.WalkDir(*configFolderPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && filepath.Ext(d.Name()) == ".yaml" || filepath.Ext(d.Name()) == ".yml" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fileNameWithoutExt := strings.Split(d.Name(), ".")[0] // golang.yaml -> golang
			baseFolder := filepath.Base(filepath.Dir(path))       // sandbox-configs/golang/golang.yaml -> sandbox-configs/golang -> golang

			sandboxConfigKey := baseFolder // golang
			if fileNameWithoutExt != baseFolder {
				sandboxConfigKey += "-" + fileNameWithoutExt // golang -> golang-web-app
			}

			m := make(map[string]any)
			err = yaml.NewDecoder(bytes.NewBuffer(data)).Decode(m)
			if err != nil {
				return err
			}

			sandboxTemplates["templates"][sandboxConfigKey] = m
		}

		return err
	})
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(*templateFileOutPath, os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}

	err = file.Truncate(0)
	if err != nil {
		panic(err)
	}

	err = yaml.NewEncoder(file).Encode(sandboxTemplates)
	if err != nil {
		panic(err)
	}
}
