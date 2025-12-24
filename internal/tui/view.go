package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.state == StateConfirmTool {
		return m.confirmView()
	}

	var s strings.Builder

	// 1. Viewport (Chat History)
	s.WriteString(m.viewport.View())
	s.WriteString("\n")

	// 2. Suggestions or Spacer
	if m.showSuggestions {
		s.WriteString(m.suggestionsView() + "\n")
	} else if m.state != StateLoading {
		s.WriteString("\n")
	}

	// 3. Input or Loading
	if m.state == StateLoading {
		s.WriteString(fmt.Sprintf("  %s %s\n", m.spinner.View(), subtleStyle.Render("Thinking...")))
	} else {
		s.WriteString(m.textarea.View() + "\n")
	}

	// 4. Status Bar
	s.WriteString(m.statusBarView())

	return s.String()
}

func (m Model) statusBarView() string {
	cwd, _ := os.Getwd()
	// Shorten path
	home, _ := os.UserHomeDir()
	cwd = strings.Replace(cwd, home, "~", 1)

	status := "IDLE"
	if m.state == StateLoading {
		status = "BUSY"
	}

	w := m.viewport.Width
	if w <= 0 {
		w = 80
	}

	left := statusTextStyle.Render(" " + status + " ")
	left += statusKeyStyle.Render(" Context: ") + cwd

	help := " Esc: cancel • Ctrl+C: exit "
	if m.showSuggestions {
		help = " ↑/↓: navigate • Tab: apply "
	}
	
	spaces := w - lipgloss.Width(left) - lipgloss.Width(help)
	if spaces < 0 {
		spaces = 0
	}
	
	return statusBarStyle.Render(left + strings.Repeat(" ", spaces) + help)
}

func (m Model) suggestionsView() string {
	var s strings.Builder
	s.WriteString(subtleStyle.Render("  SUGGESTIONS:") + "\n")
	for i, suggestion := range m.suggestions {
		prefix := "  "
		line := suggestion
		if i == m.suggestionIdx {
			line = selectedSuggestionStyle.Render(" " + suggestion + " ")
		} else {
			line = suggestionStyle.Render(suggestion)
		}
		s.WriteString(prefix + line + "\n")
	}
	return s.String()
}

func (m Model) confirmView() string {
	var s strings.Builder
	s.WriteString(m.viewport.View())
	s.WriteString("\n")
	
	if m.currentReq != nil {
		argsJSON, _ := json.MarshalIndent(m.currentReq.Args, "", "  ")
		
		title := toolStyle.Render(" 🛠️  TOOL REQUEST: " + m.currentReq.Name)
		body := subtleStyle.Render(string(argsJSON))
		footer := "\n" + greenStyle.Render(" [Y] Approve ") + " " + redStyle.Render(" [N] Deny ")

		card := toolCardStyle.Render(title + "\n\n" + body + footer)
		s.WriteString(card + "\n")
	}

	s.WriteString(m.statusBarView())
	return s.String()
}
