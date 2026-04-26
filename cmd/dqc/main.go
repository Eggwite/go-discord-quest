//go:build windows

// Package main is the entry point for the go-discord-quest (dqc) CLI application.
// It initializes the Bubble Tea TUI model and manages critical application signals
// (like SIGINT) to ensure running child processes (game stubs) are cleanly terminated.
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Eggwite/go-discord-quest/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := tui.New()
	p := tea.NewProgram(model, tea.WithAltScreen())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		model.KillRunner()
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
