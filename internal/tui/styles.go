package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	danger    = lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"}

	// Styles
	senderStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true).
			MarginRight(1)

	botSenderStyle = lipgloss.NewStyle().
			Foreground(special).
			Bold(true).
			MarginRight(1)

	systemSenderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")). // Gray
			Bold(true).
			MarginRight(1)

	toolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("208")). // Orange
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1, 2)

	subtleStyle = lipgloss.NewStyle().
			Foreground(subtle)
)
