package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
)

var (
	colorPurple  = lipgloss.Color("#7D56F4")
	colorYellow  = lipgloss.Color("#FFFF00")
	colorOrgange = lipgloss.Color("#FFAA33")
	colorGreen   = lipgloss.Color("#04B575")

	header = lipgloss.NewStyle().
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		// BorderBackground(lipgloss.Color("#7D56F4")).
		Foreground(lipgloss.Color("#7D56F4")).
		// Background(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00"))

	alert = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA33"))

	ErrCommandNotFound = errors.New("command not found")
)

func logHeader(s string) {
	fmt.Println(header.SetString(s))
}

func logMessage(s string, color lipgloss.Color) {
	style := lipgloss.NewStyle().
		Foreground(color).
		SetString(s)

	fmt.Println(style)
}

func hasCommand(cmd string) error {
	err := exec.Command("which", cmd).Run()
	if err != nil {
		return ErrCommandNotFound
	}
	return nil
}

func renderTemplate(templateData string, data any) (string, error) {
	// >>>>> Render post run message using go templates.
	t, err := template.New("template").Parse(templateData)
	if err != nil {
		return "", err
	}

	buff := bytes.NewBuffer([]byte{})
	err = t.Execute(buff, data)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}
