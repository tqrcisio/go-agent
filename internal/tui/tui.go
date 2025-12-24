package tui

import (
	"fmt"
	"go-agent/internal/geminiclient"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func StartTUI(client *geminiclient.RawClient, reqChan <-chan ToolConfirmRequest) error {
	p := tea.NewProgram(
		NewModel(client, reqChan),
		tea.WithAltScreen(), // Use full screen
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting TUI: %v\n", err)
		os.Exit(1)
	}
	return nil
}

