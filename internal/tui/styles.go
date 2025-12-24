package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors (Modern/Claude-inspired)
	white       = lipgloss.Color("#FFFFFF")
	gray        = lipgloss.Color("#626262")
	darkGray    = lipgloss.Color("#353535")
	lightGray   = lipgloss.Color("#D9DCCF")
	purple      = lipgloss.Color("#874BFD")
	cyan        = lipgloss.Color("#00ADD8")
	green       = lipgloss.Color("#43BF6D")
	orange      = lipgloss.Color("#FFA500")
	red         = lipgloss.Color("#F25D94")

	// Base Styles
	senderStyle = lipgloss.NewStyle().
			Foreground(purple).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	botSenderStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	systemSenderStyle = lipgloss.NewStyle().
			Foreground(gray).
			Italic(true).
			PaddingLeft(1)

	// Message Containers
	userMsgStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(purple).
			PaddingLeft(2).
			MarginBottom(1)

	botMsgStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(cyan).
			PaddingLeft(2).
			MarginBottom(1)

	// Tool & System Styles
	toolStyle = lipgloss.NewStyle().
			Foreground(orange).
			Bold(true)

	greenStyle = lipgloss.NewStyle().Foreground(green)
	redStyle   = lipgloss.NewStyle().Foreground(red)

	toolCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(orange).
			Padding(0, 1).
			MarginLeft(2).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Background(darkGray).
			Padding(0, 1)

	statusTextStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Bold(true)

	statusKeyStyle = lipgloss.NewStyle().
			Inherit(statusBarStyle).
			Foreground(purple)

	// Autocomplete Styles
	suggestionStyle = lipgloss.NewStyle().
			Foreground(gray).
			Padding(0, 1)

	selectedSuggestionStyle = lipgloss.NewStyle().
			Foreground(white).
			Background(purple).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	subtleStyle = lipgloss.NewStyle().
			Foreground(gray)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2)
)

