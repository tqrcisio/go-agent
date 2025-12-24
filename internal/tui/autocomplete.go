package tui

import (
	"sort"
	"strings"
)

// List of slash commands for autocomplete
var slashCommands = []string{
	"/help",
	"/clear",
	"/quit",
	"/exit",
	"/tools",
}

// updateSuggestions recalculates the suggestion list based on input
func (m *Model) updateSuggestions(input string) {
	// 1. Detect Trigger
	// We need to find the word being typed at the cursor.
	// For simplicity, we'll check the last word or if the input starts with /
	
	// Case 1: Slash commands (must be at start)
	if strings.HasPrefix(input, "/") && !strings.Contains(input, " ") {
		m.triggerType = "/"
		m.filterText = input
		m.suggestions = filterStrings(slashCommands, input)
		m.showSuggestions = len(m.suggestions) > 0
		m.suggestionIdx = 0
		return
	}

	// Case 2: File mentions (@)
	// Find the last "@" index
	lastAt := strings.LastIndex(input, "@")
	if lastAt != -1 {
		// Check if it's a valid trigger (start of string or preceded by space)
		if lastAt == 0 || input[lastAt-1] == ' ' {
			// Extract text after @
			filter := input[lastAt+1:]
			// Ensure we don't have spaces after @ (autocomplete ends at space)
			if !strings.Contains(filter, " ") {
				m.triggerType = "@"
				m.filterText = filter
				
				if m.fileCache != nil {
					m.suggestions = filterStrings(m.fileCache.Files, filter)
				}
				
				m.showSuggestions = len(m.suggestions) > 0
				m.suggestionIdx = 0
				return
			}
		}
	}

	m.showSuggestions = false
	m.suggestions = nil
}

func filterStrings(candidates []string, pattern string) []string {
	var matches []string
	pattern = strings.ToLower(pattern)
	
	// Prioritize prefix matches, then contains
	// Actually for now just contains or fuzzy
	for _, c := range candidates {
		if strings.Contains(strings.ToLower(c), pattern) {
			matches = append(matches, c)
		}
	}
	
	// Sort by length then alpha (simple heuristic)
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i]) < len(matches[j])
	})
	
	if len(matches) > 5 {
		matches = matches[:5] // Limit to 5 suggestions
	}
	
	return matches
}

func (m *Model) applySuggestion() {
	if !m.showSuggestions || len(m.suggestions) == 0 {
		return
	}

	selected := m.suggestions[m.suggestionIdx]
	val := m.textarea.Value()

	if m.triggerType == "/" {
		// Replace whole input for slash commands
		m.textarea.SetValue(selected + " ")
	} else if m.triggerType == "@" {
		// Replace the part after the last @
		lastAt := strings.LastIndex(val, "@")
		if lastAt != -1 {
			prefix := val[:lastAt]
			m.textarea.SetValue(prefix + "@" + selected + " ")
		}
	}
	
	// Move cursor to end
	m.textarea.SetCursor(len(m.textarea.Value()))
	m.showSuggestions = false
}
