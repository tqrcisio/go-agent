package geminiclient

import (
	"context"
	"fmt"
	"go-agent/internal/tools"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// ChatSession wraps the genai.ChatSession.
type ChatSession struct {
	Session *genai.ChatSession
	Client  *GeminiClient
}

// StartChat initializes a new chat session with tools configured.
func (c *GeminiClient) StartChat() *ChatSession {
	// Configure tools on the model
	c.model.Tools = []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "list_files",
					Description: "List files and directories in a given path. Use this to explore the project structure.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "The relative path to list (e.g., '.', 'cmd/'). Defaults to '.'",
							},
						},
					},
				},
				{
					Name:        "read_file",
					Description: "Read the content of a specific file. Use this to examine code.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"path": {
								Type:        genai.TypeString,
								Description: "The relative path of the file to read.",
							},
						},
						Required: []string{"path"},
					},
				},
				{
					Name:        "search_files",
					Description: "Search for a text pattern in the codebase using grep.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"pattern": {
								Type:        genai.TypeString,
								Description: "The string pattern to search for.",
							},
							"path": {
								Type:        genai.TypeString,
								Description: "The directory to search in. Defaults to '.'",
							},
						},
						Required: []string{"pattern"},
					},
				},
			},
		},
	}

	cs := c.model.StartChat()

	return &ChatSession{
		Session: cs,
		Client:  c,
	}
}

// SendMessage sends a message to the chat and handles tool calls loop.
func (s *ChatSession) SendMessage(ctx context.Context, msg string) (string, error) {
	if s.Client.Debug {
		log.Printf("[DEBUG] Sending message: %s", msg)
	}

	resp, err := s.Session.SendMessage(ctx, genai.Text(msg))
	if err != nil {
		return "", err
	}

	// Loop to handle multi-turn tool interactions
	for {
		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			return "", fmt.Errorf("empty response from model")
		}

		content := resp.Candidates[0].Content
		var toolResponses []genai.Part
		var textResponse strings.Builder
		hasToolCalls := false

		for _, part := range content.Parts {
			if funcCall, ok := part.(genai.FunctionCall); ok {
				hasToolCalls = true
				if s.Client.Debug {
					log.Printf("[DEBUG] Tool call: %s(%v)", funcCall.Name, funcCall.Args)
				}

				// User feedback
				fmt.Printf("🛠️  Executando: %s...\n", funcCall.Name)

				var result map[string]interface{}
				var toolErr error

				switch funcCall.Name {
				case "list_files":
					result, toolErr = tools.ToolListFiles(ctx, funcCall.Args)
				case "read_file":
					result, toolErr = tools.ToolReadFile(ctx, funcCall.Args)
				case "search_files":
					result, toolErr = tools.ToolSearchFiles(ctx, funcCall.Args)
				default:
					result = map[string]interface{}{"error": "Unknown function"}
				}

				if toolErr != nil {
					result = map[string]interface{}{"error": toolErr.Error()}
				}

				toolResponses = append(toolResponses, genai.FunctionResponse{
					Name:     funcCall.Name,
					Response: result,
				})
			} else if text, ok := part.(genai.Text); ok {
				textResponse.WriteString(string(text))
			}
		}

		if hasToolCalls {
			// Send all tool responses back to the model
			if s.Client.Debug {
				log.Printf("[DEBUG] Sending %d tool responses", len(toolResponses))
			}
			resp, err = s.Session.SendMessage(ctx, toolResponses...)
			if err != nil {
				return "", fmt.Errorf("error sending tool responses: %w", err)
			}
			continue // Loop to get the next response from the model
		}

		// If no tool calls, return the accumulated text
		return textResponse.String(), nil
	}
}

// History returns the chat history.
func (s *ChatSession) History() []*genai.Content {
	return s.Session.History
}