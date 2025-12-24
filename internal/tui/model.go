package tui

import (
	"fmt"
	"go-agent/internal/geminiclient"
	"go-agent/internal/tools"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type SessionState int

const (
	StateChat SessionState = iota
	StateLoading
	StateConfirmTool
)

type ChatMessage struct {
	Role    string
	Content string
}

type ToolConfirmRequest struct {
	Name string
	Args map[string]interface{}
	Resp chan bool
}

type Model struct {
	client     *geminiclient.RawClient
	state      SessionState
	viewport   viewport.Model
	textarea   textarea.Model
	spinner    spinner.Model
	messages   []ChatMessage
	err        error
	
	// Channels for communication with the AI runner
	confirmReqChan <-chan ToolConfirmRequest // Receive requests from AI
	currentReq     *ToolConfirmRequest       // The request currently being asked
	
	// Autocomplete
	fileCache      *tools.ProjectFilesCache
	suggestions    []string
	suggestionIdx  int
	showSuggestions bool
	filterText     string // The text being typed for filter (@main.go -> main.go)
	triggerType    string // "@" or "/"
}

func NewModel(client *geminiclient.RawClient, reqChan <-chan ToolConfirmRequest) Model {
	ta := textarea.New()
	ta.Placeholder = "Ask a question about your code..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 0
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(0, 0)
	vp.SetContent("Welcome to the Gemini Code Agent!\nType a message to start.")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Initialize File Cache
	cache, err := tools.NewProjectFilesCache(".")
	if err != nil {
		// Log error to viewport or ignore
		vp.SetContent(fmt.Sprintf("Warning: Failed to load file cache: %v", err))
	}

	return Model{
		client:         client,
		state:          StateChat,
		textarea:       ta,
		viewport:       vp,
		spinner:        sp,
		messages:       []ChatMessage{},
		confirmReqChan: reqChan,
		fileCache:      cache,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		waitForConfirmation(m.confirmReqChan),
	)
}

// Helper to render markdown safely
func renderMarkdown(text string, width int) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return out
}
