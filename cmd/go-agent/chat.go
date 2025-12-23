package main

import (
	"context"
	"fmt"
	"go-agent/internal/geminiclient"
	"go-agent/internal/tui"
	"os"

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

// Variables needed for chat command
var (
	chatCtx   context.Context
	client    *geminiclient.RawClient
)

func startChat() {
	_ = godotenv.Load()
	chatCtx = context.Background()
	debug, _ := rootCmd.PersistentFlags().GetBool("debug")

	// Channel for tool confirmation requests
	confirmReqChan := make(chan tui.ToolConfirmRequest)

	// Initialize Client with special ConfirmFn
	client = geminiclient.NewRawClient(debug)
	client.ConfirmFn = func(name string, args map[string]interface{}) bool {
		respChan := make(chan bool)
		// Send request to TUI
		confirmReqChan <- tui.ToolConfirmRequest{
			Name: name,
			Args: args,
			Resp: respChan,
		}
		// Wait for response from TUI
		return <-respChan
	}

	fmt.Println("🤖 Initializing TUI...")

	// Start the Bubble Tea program
	if err := tui.StartTUI(client, confirmReqChan); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

