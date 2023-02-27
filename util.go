package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
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
)

func logHeader(s string) {
	fmt.Println(header.SetString(s))
}

func logInfo(s string) {
	fmt.Println(info.SetString(s))
}

func logAlert(s string) {
	fmt.Println(alert.SetString(s))
}
