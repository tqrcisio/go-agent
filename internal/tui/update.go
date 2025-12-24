package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Msg types
type errMsg error
type responseMsg string

// Command to wait for tool confirmation requests
func waitForConfirmation(ch <-chan ToolConfirmRequest) tea.Cmd {
	return func() tea.Msg {
		req, ok := <-ch
		if !ok {
			return nil
		}
		return req
	}
}

// Command to send message to Gemini
func sendMessageCmd(m *Model, msg string) tea.Cmd {
	return func() tea.Msg {
		// Process @files expansion before sending
		// We use a background context or a specific timeout context if needed
		fullMsg := processFilesContext(context.Background(), msg)
		
		resp, err := m.client.SendMessageRaw(context.Background(), fullMsg)
		if err != nil {
			return errMsg(err)
		}
		return responseMsg(resp)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		cmd   tea.Cmd
	)

	switch msg := msg.(type) {

		case tea.KeyMsg:

			switch msg.Type {

			case tea.KeyCtrlC, tea.KeyEsc:

				if m.showSuggestions {

					m.showSuggestions = false

					return m, nil

				}

				return m, tea.Quit

				

			case tea.KeyTab:

				if m.showSuggestions {

					m.applySuggestion()

					return m, nil

				}

				

			case tea.KeyUp:

				if m.showSuggestions {

					if m.suggestionIdx > 0 {

						m.suggestionIdx--

					}

					return m, nil

				}

				

			case tea.KeyDown:

				if m.showSuggestions {

					if m.suggestionIdx < len(m.suggestions)-1 {

						m.suggestionIdx++

					}

					return m, nil

				}

	

			case tea.KeyEnter:

				if m.showSuggestions {

					m.applySuggestion()

					return m, nil

				}

				

				if m.state == StateChat {

					if m.textarea.Value() == "" {

						return m, nil

					}

					

					userInput := m.textarea.Value()

	

					// Handle Slash Commands

					if strings.HasPrefix(userInput, "/") {

						m.textarea.Reset()

						// Add the command itself to history for visibility? 

						// Maybe just execute it. Usually commands are ephemeral or shown as system actions.

						// Let's NOT add the user's slash command to the chat history to keep it clean,

						// except maybe as a log? "You > /help"

						// Let's add it for clarity.

						m.messages = append(m.messages, ChatMessage{Role: "user", Content: userInput})

						m.viewport.SetContent(m.renderConversation()) // Update view with command

						

						cmd := m.handleSlashCommand(userInput)

						return m, cmd

					}

	

					m.messages = append(m.messages, ChatMessage{Role: "user", Content: userInput})

					

					// Update viewport with new message immediately

					m.viewport.SetContent(m.renderConversation())

					m.viewport.GotoBottom()

					

					m.textarea.Reset()

					m.showSuggestions = false // Clear suggestions on send

					m.state = StateLoading

					return m, tea.Batch(

						m.spinner.Tick,

						sendMessageCmd(&m, userInput),

					)

				} else if m.state == StateConfirmTool {

					// Handle Y/N for tool confirmation if user types manually?

					// Actually, we'll listen for specific keys 'y' or 'n' below

				}

			}

			

			// Handle specific keys for confirmation

			if m.state == StateConfirmTool {

				switch msg.String() {

				case "y", "Y":

					if m.currentReq != nil {

						m.currentReq.Resp <- true

						m.currentReq = nil

						m.state = StateLoading // Go back to loading (waiting for tool execution)

						return m, waitForConfirmation(m.confirmReqChan)

					}

				case "n", "N":

					if m.currentReq != nil {

						m.currentReq.Resp <- false

						m.currentReq = nil

						m.state = StateLoading // Go back to loading (waiting for result/next step)

						return m, waitForConfirmation(m.confirmReqChan)

					}

				}

			}

	

		case responseMsg:

			m.state = StateChat

			m.messages = append(m.messages, ChatMessage{Role: "model", Content: string(msg)})

			m.viewport.SetContent(m.renderConversation())

			m.viewport.GotoBottom()

			return m, nil

	

		case errMsg:

			m.state = StateChat

			m.err = msg

			return m, nil

	

		case ToolConfirmRequest:

			m.state = StateConfirmTool

			m.currentReq = &msg

			return m, nil // Stop ticking spinner? Or keep it? Maybe keep it but overlay dialog.

	

		case tea.WindowSizeMsg:

			m.viewport.Width = msg.Width

			m.viewport.Height = msg.Height - m.textarea.Height() - 4 // Leave room for header/footer

			m.textarea.SetWidth(msg.Width)

			m.viewport.SetContent(m.renderConversation())

		}

	

		// Update components

		if m.state == StateChat {

			m.textarea, tiCmd = m.textarea.Update(msg)

			// Update suggestions after textarea update

			if msg, ok := msg.(tea.KeyMsg); ok {

				// Don't update on navigation keys if we consumed them

				if !m.showSuggestions || (msg.Type != tea.KeyUp && msg.Type != tea.KeyDown && msg.Type != tea.KeyEnter && msg.Type != tea.KeyTab) {

					m.updateSuggestions(m.textarea.Value())

				}

			}

		}

	
	
m.viewport, vpCmd = m.viewport.Update(msg)
	
	if m.state == StateLoading || m.state == StateConfirmTool {
		var sCmd tea.Cmd
		m.spinner, sCmd = m.spinner.Update(msg)
		cmd = tea.Batch(cmd, sCmd)
	}

	return m, tea.Batch(cmd, tiCmd, vpCmd)
}

func (m Model) renderConversation() string {
	var s string
	width := m.viewport.Width - 4 // Padding
	if width < 10 {
		width = 10
	}

	for _, msg := range m.messages {
		content := renderMarkdown(msg.Content, width)
		if msg.Role == "user" {
			s += fmt.Sprintf("%s\n%s\n\n", senderStyle.Render("You:"), content)
		} else if msg.Role == "model" {
			s += fmt.Sprintf("%s\n%s\n\n", botSenderStyle.Render("Gemini:"), content)
		} else if msg.Role == "system" {
			s += fmt.Sprintf("%s\n%s\n\n", systemSenderStyle.Render("System:"), content)
		}
	}
	return s
}
