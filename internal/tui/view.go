package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.state == StateConfirmTool {
		return m.confirmView()
	}

	var s strings.Builder

	// Header?
	// s.WriteString("Gemini Code Agent\n\n")

	// Viewport (Chat History)
	s.WriteString(m.viewport.View())
	s.WriteString("\n")

	// Footer Area
	if m.state == StateLoading {
		s.WriteString(fmt.Sprintf("\n%s Thinking...\n", m.spinner.View()))
	} else {
		// Render suggestions if available
		if m.showSuggestions {
			s.WriteString("\n" + m.suggestionsView() + "\n")
		} else {
			s.WriteString("\n") // Spacer
		}
		
		// Input Area
		s.WriteString(m.textarea.View())
		s.WriteString("\n")
		s.WriteString(subtleStyle.Render("Press Enter to send • Esc to quit"))
	}

	return s.String()
}

func (m Model) suggestionsView() string {
	var s strings.Builder
	for i, suggestion := range m.suggestions {
		if i == m.suggestionIdx {
			s.WriteString(selectedSuggestionStyle.Render(suggestion))
		} else {
			s.WriteString(suggestionStyle.Render(suggestion))
		}
		s.WriteString(" ") // Space between suggestions (horizontal list? or vertical?)
		// Let's do horizontal for compact look or vertical?
		// Vertical is better for file paths
		s.WriteString("\n")
	}
	return s.String()
}

func (m Model) confirmView() string {
	// Overlay or full screen replace? For simplicity, replace/append at bottom
	
	// We want to show the chat history still? 
	// Ideally yes, but maybe simpler to show just the confirmation for now to be safe with layout.
	// Or we can construct a string that includes the viewport.
	
	var s strings.Builder
	s.WriteString(m.viewport.View())
	s.WriteString("\n")
	
	if m.currentReq != nil {
		argsJSON, _ := json.MarshalIndent(m.currentReq.Args, "", "  ")
		
		panel := boxStyle.Render(fmt.Sprintf(
			"%s requests to run tool: %s\n\nArguments:\n%s\n\n%s",
			botSenderStyle.Render("🤖 Agent"),
			toolStyle.Render(m.currentReq.Name),
			lipgloss.NewStyle().Foreground(subtle).Render(string(argsJSON)),
			"Allow execution? (y/n)",
		))
		s.WriteString("\n" + panel + "\n")
	}

	return s.String()
}
