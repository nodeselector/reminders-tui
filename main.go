package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nodeselector/reminders-tui/internal/model"
	"github.com/nodeselector/reminders-tui/internal/remindctl"
)

func main() {
	// Look for remindctl in ~/.local/bin first, then PATH
	bin := "remindctl"
	home, err := os.UserHomeDir()
	if err == nil {
		localBin := filepath.Join(home, ".local", "bin", "remindctl")
		if _, err := os.Stat(localBin); err == nil {
			bin = localBin
		}
	}

	client := remindctl.New(bin)
	m := model.New(client)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
