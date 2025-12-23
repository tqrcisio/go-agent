package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSlashCommand processes commands starting with /
func (m *Model) handleSlashCommand(input string) tea.Cmd {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}
	cmd := parts[0]

	switch cmd {
	case "/quit", "/exit":
		return tea.Quit

	case "/clear":
		m.messages = []ChatMessage{}
		m.client.ClearHistory()
		m.viewport.SetContent("Chat cleared. Starting fresh context.")
		return nil

	case "/help":
		helpText := `
## Available Commands

- **/clear**: Clear chat history and reset context.
- **/quit**, **/exit**: Exit the application.
- **/tools**: List available tools.
- **/help**: Show this help message.

💡 **Tip**: Use **@filename** to include file content in your message.
`
		m.messages = append(m.messages, ChatMessage{Role: "system", Content: helpText})
		m.viewport.SetContent(m.renderConversation())
		m.viewport.GotoBottom()
		return nil

	case "/tools":
		var sb strings.Builder
		sb.WriteString("## Available Tools\n\n")
		
		for _, tool := range m.client.Tools {
			if fd, ok := tool["function_declarations"].([]map[string]interface{}); ok {
				for _, f := range fd {
					name, _ := f["name"].(string)
					desc, _ := f["description"].(string)
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, desc))
				}
			}
		}
		
		m.messages = append(m.messages, ChatMessage{Role: "system", Content: sb.String()})
		m.viewport.SetContent(m.renderConversation())
		m.viewport.GotoBottom()
		return nil

	default:
		m.messages = append(m.messages, ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Unknown command: **%s**. Type **/help** for a list of commands.", cmd),
		})
		m.viewport.SetContent(m.renderConversation())
		m.viewport.GotoBottom()
		return nil
	}
}
