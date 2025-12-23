package main

import (
	"context"
	"fmt"
	"go-agent/internal/geminiclient"
	"go-agent/internal/tools"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/glamour"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat with your codebase",
	Long: `Starts a chat session where you can ask questions about the code.
The agent uses tools to explore the file system and read files as needed.
Use @filename to include file content directly in your message.`,
	Run: func(cmd *cobra.Command, args []string) {
		startChat()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

var (
	fileCache *tools.ProjectFilesCache
	chatCtx   context.Context
	chatSess  *geminiclient.ChatSession
	client    *geminiclient.GeminiClient

	// Cancellation control
	currentCancel context.CancelFunc
	cancelMu      sync.Mutex
	lastSignal    time.Time

	// Slash commands list for autocomplete
	slashCommands = []prompt.Suggest{
		{Text: "/clear", Description: "Clear chat history and terminal"},
		{Text: "/save", Description: "Save chat history to a markdown file"},
		{Text: "/quit", Description: "Exit the chat"},
		{Text: "/exit", Description: "Exit the chat"},
	}
)

func startChat() {
	_ = godotenv.Load()
	chatCtx = context.Background()
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		for range sigChan {
			now := time.Now()

			// Double Ctrl+C check
			if now.Sub(lastSignal) < 500*time.Millisecond {
				fmt.Println("\nForce quitting...")
				os.Exit(0)
			}
			lastSignal = now

			cancelMu.Lock()
			if currentCancel != nil {
				fmt.Println("\nCancelling current operation...")
				currentCancel()
				currentCancel = nil
			} else {
				fmt.Println("\n(Press Ctrl+C again quickly to exit)")
			}
			cancelMu.Unlock()
		}
	}()

	fmt.Println("🤖 Initializing Agent...")
	
	var err error
	fileCache, err = tools.NewProjectFilesCache(".")
	if err != nil {
		fmt.Printf("Warning: Could not build file cache: %v\n", err)
	}

	client, err = geminiclient.New(chatCtx, debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	chatSess = client.StartChat()

	fmt.Println("✅ Agent Ready! Ask me anything about this project.")
	fmt.Println("   💡 Use @ to autocomplete files, / for commands.")
	fmt.Println("   (Type 'exit' or '/quit' to stop)")
	fmt.Println()

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("You > "),
		prompt.OptionPrefixTextColor(prompt.Blue),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
	)
	p.Run()
}

func completer(d prompt.Document) []prompt.Suggest {
	word := d.GetWordBeforeCursor()
	
	// Handle Slash Commands
	if strings.HasPrefix(word, "/") {
		var suggests []prompt.Suggest
		for _, c := range slashCommands {
			if strings.HasPrefix(c.Text, word) {
				suggests = append(suggests, c)
			}
		}
		return suggests
	}

	// Handle @ Files
	if strings.HasPrefix(word, "@") {
		filter := strings.TrimPrefix(word, "@")
		var suggests []prompt.Suggest
		for _, f := range fileCache.Files {
			if filter == "" || strings.Contains(strings.ToLower(f), strings.ToLower(filter)) {
				suggests = append(suggests, prompt.Suggest{Text: "@" + f})
			}
		}
		return suggests
	}

	return []prompt.Suggest{}
}

func executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Handle exit commands
	if input == "exit" || input == "quit" || input == "/quit" || input == "/exit" {
		fmt.Println("Bye!")
		os.Exit(0)
	}

	// Handle Slash Commands
	if strings.HasPrefix(input, "/") {
		handleSlashCommand(input)
		return
	}

	// Create cancellable context for this turn
	ctx, cancel := context.WithCancel(chatCtx)

	cancelMu.Lock()
	currentCancel = cancel
	cancelMu.Unlock()

	defer func() {
		cancel()
		cancelMu.Lock()
		// Only clear if it hasn't been replaced (though strictly it's sequential here)
		currentCancel = nil
		cancelMu.Unlock()
	}()

	// Process @files
	finalPrompt := processFilesContext(ctx, input)

	// Check if cancelled during file processing
	if ctx.Err() != nil {
		fmt.Println("\nOperation cancelled.")
		return
	}

	stopSpinner := startSpinner()
	resp, err := chatSess.SendMessage(ctx, finalPrompt)
	stopSpinner()

	if err != nil {
		if err == context.Canceled {
			fmt.Println("\nCancelled.")
		} else {
			fmt.Printf("\033[31mError: %v\033[0m\n", err)
		}
		return
	}

	fmt.Printf("\033[1;32mAgent:\033[0m\n")
	
	// Render Markdown response
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		// Fallback to plain text if renderer fails
		fmt.Printf("%s\n\n", resp)
	} else {
		out, err := renderer.Render(resp)
		if err != nil {
			fmt.Printf("%s\n\n", resp)
		} else {
			fmt.Print(out)
			fmt.Println()
		}
	}
}

func handleSlashCommand(input string) {
	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "/clear":
		chatSess = client.StartChat()
		fmt.Print("\033[H\033[2J") // Clear screen and reset cursor
		fmt.Println("✅ Session cleared. Starting fresh.")
		fmt.Println()
	case "/save":
		filename := "chat_history.md"
		if len(parts) > 1 {
			filename = parts[1]
			if !strings.HasSuffix(filename, ".md") {
				filename += ".md"
			}
		}
		
		history := chatSess.History()
		var sb strings.Builder
		sb.WriteString("# Chat History - " + time.Now().Format("2006-01-02 15:04:05") + "\n\n")
		
		for _, h := range history {
			role := "Unknown"
			if h.Role == "user" {
				role = "### User"
			} else if h.Role == "model" {
				role = "### Agent"
			}
			
			sb.WriteString(role + ":\n")
			for _, p := range h.Parts {
				if t, ok := p.(genai.Text); ok {
					sb.WriteString(string(t) + "\n\n")
				}
			}
		}
		
		err := os.WriteFile(filename, []byte(sb.String()), 0644)
		if err != nil {
			fmt.Printf("\033[31mError saving history: %v\033[0m\n", err)
		} else {
			fmt.Printf("✅ History saved to %s\n", filename)
		}
	default:
		fmt.Printf("\033[31mUnknown command: %s. Type / to see available commands.\033[0m\n", cmd)
	}
}

func startSpinner() func() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Print("\r\033[K") // Clear line
				close(done)
				return
			case <-ticker.C:
				fmt.Printf("\r\033[33m%s Thinking...\033[0m", frames[i%len(frames)])
				i++
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

func processFilesContext(ctx context.Context, input string) string {
	re := regexp.MustCompile(`@([^\s]+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return input
	}

	var sb strings.Builder
	sb.WriteString(input)
	sb.WriteString("\n\n---\nContexto de arquivos selecionados:\n")

	processed := make(map[string]bool)
	for _, match := range matches {
		if ctx.Err() != nil {
			break
		}

		fileName := match[1]
		if processed[fileName] {
			continue
		}

		// 1. Basic security and existence checks
		info, err := os.Stat(fileName)
		if err != nil {
			fmt.Printf("\n\033[31m⚠️  File not found: %s\033[0m\n", fileName)
			continue
		}

		if info.IsDir() {
			fmt.Printf("\n\033[33m⚠️  Skipping directory: %s (support coming soon)\033[0m\n", fileName)
			continue
		}

		// 2. Size limit check
		if info.Size() > 100*1024 { // 100KB limit
			fmt.Printf("\n\033[33m⚠️  File too large (>100KB), skipping: %s\033[0m\n", fileName)
			continue
		}

		// 3. Read and Binary check
		content, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Printf("\n\033[31m⚠️  Error reading file: %s\033[0m\n", fileName)
			continue
		}

		if isBinary(content) {
			fmt.Printf("\n\033[33m⚠️  Binary file detected, skipping: %s\033[0m\n", fileName)
			continue
		}

		sb.WriteString(fmt.Sprintf("\nArquivo: %s\n```\n%s\n```\n", fileName, string(content)))
		processed[fileName] = true
	}

	return sb.String()
}

func isBinary(content []byte) bool {
	// Check for null byte in the first 1024 bytes to detect binary files
	limit := 1024
	if len(content) < limit {
		limit = len(content)
	}
	for i := 0; i < limit; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
