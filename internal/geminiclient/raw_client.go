package geminiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go-agent/internal/config"
	"go-agent/internal/tools"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// RawClient manages a chat session using raw HTTP requests to bypass SDK limitations regarding thought_signatures.
type RawClient struct {
	APIKey    string
	Model     string
	History   []map[string]interface{}
	Tools     []map[string]interface{}
	Debug     bool
	Settings  *config.Settings
	ConfirmFn func(name string, args map[string]interface{}) bool
}

// NewRawClient creates a new raw HTTP client for Gemini.
func NewRawClient(debug bool) *RawClient {
	settings, err := config.LoadSettings()
	if err != nil {
		log.Printf("Warning: Failed to load settings: %v. Using defaults.", err)
		def := config.DefaultSettings()
		settings = &def
	}

	return &RawClient{
		APIKey:   os.Getenv("GEMINI_API_KEY"),
		Model:    "gemini-3-flash-preview", // Hardcoded for now as per requirement
		Debug:    debug,
		Settings: settings,
		History:  []map[string]interface{}{},
		ConfirmFn: func(name string, args map[string]interface{}) bool {
			// Default implementation if none provided
			argsJSON, _ := json.Marshal(args)
			fmt.Printf("\n\033[1;33m⚠️  O Agente deseja executar a ferramenta:\033[0m \033[1;36m%s\033[0m\n", name)
			fmt.Printf("   Argumentos: %s\n", string(argsJSON))
			fmt.Printf("   \033[1;35mAutorizar execução? [y/N]: \033[0m")

			var response string
			fmt.Scanln(&response)
			response = strings.ToLower(strings.TrimSpace(response))
			return response == "y" || response == "yes"
		},
		Tools: []map[string]interface{}{
			{
				"function_declarations": []map[string]interface{}{
					{
						"name":        "list_files",
						"description": "List files and directories in a given path.",
						"parameters": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"path": map[string]interface{}{
									"type":        "STRING",
									"description": "The relative path to list. Defaults to '.'",
								},
							},
						},
					},
					{
						"name":        "read_file",
						"description": "Read the content of a specific file.",
						"parameters": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"path": map[string]interface{}{
									"type":        "STRING",
									"description": "The relative path of the file to read.",
								},
							},
							"required": []string{"path"},
						},
					},
					{
						"name":        "search_files",
						"description": "Search for a text pattern in the codebase using grep.",
						"parameters": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"pattern": map[string]interface{}{
									"type":        "STRING",
									"description": "The string pattern to search for.",
								},
								"path": map[string]interface{}{
									"type":        "STRING",
									"description": "The directory to search in. Defaults to '.'",
								},
							},
							"required": []string{"pattern"},
						},
					},
					{
						"name":        "run_shell",
						"description": "Execute a shell command. Use for git operations, building, running tests, etc. Commands must be whitelisted.",
						"parameters": map[string]interface{}{
							"type": "OBJECT",
							"properties": map[string]interface{}{
								"command": map[string]interface{}{
									"type":        "STRING",
									"description": "The command to execute (e.g., 'git status', 'go test ./...').",
								},
							},
							"required": []string{"command"},
						},
					},
				},
			},
		},
	}
}

// ClearHistory resets the chat history.
func (c *RawClient) ClearHistory() {
	c.History = []map[string]interface{}{}
}

// SendMessageRaw sends a message and handles tool loops using raw JSON.
func (c *RawClient) SendMessageRaw(ctx context.Context, msg string) (string, error) {
	// Add user message to history
	c.History = append(c.History, map[string]interface{}{
		"role": "user",
		"parts": []map[string]interface{}{
			{"text": msg},
		},
	})

	for {
		resp, err := c.generateContent(ctx)
		if err != nil {
			return "", err
		}

		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("no candidates returned")
		}

		// Check for content
		candidate := resp.Candidates[0]
		if candidate.Content == nil {
			return "", fmt.Errorf("empty content in response")
		}

		// Important: Add the RAW content back to history to preserve everything (thoughts, etc)
		c.History = append(c.History, candidate.Content)

		// Check parts for tool calls
		var functionCalls []map[string]interface{}
		var textParts []string

		parts, ok := candidate.Content["parts"].([]interface{})
		if !ok {
			// Try without interface slice cast if it came differently
			return "", fmt.Errorf("invalid parts format in response")
		}

		for _, p := range parts {
			partMap, ok := p.(map[string]interface{})
			if !ok {
				continue
			}

			if txt, ok := partMap["text"].(string); ok {
				textParts = append(textParts, txt)
			}

			if fc, ok := partMap["functionCall"].(map[string]interface{}); ok {
				functionCalls = append(functionCalls, fc)
			}
		}

		// If no function calls, we are done
		if len(functionCalls) == 0 {
			return strings.Join(textParts, "\n"), nil
		}

		// Handle function calls
		var functionResponses []map[string]interface{}
		for _, fc := range functionCalls {
			name, _ := fc["name"].(string)
			args, _ := fc["args"].(map[string]interface{})

			if c.Debug {
				log.Printf("[DEBUG] Raw Tool Call: %s(%v)", name, args)
			}

			var result map[string]interface{}
			var toolErr error

			if c.ConfirmFn(name, args) {
				fmt.Printf("   🛠️  Executando %s...\n", name)
				// Execute tool
				switch name {
				case "list_files":
					result, toolErr = tools.ToolListFiles(ctx, args)
				case "read_file":
					result, toolErr = tools.ToolReadFile(ctx, args)
				case "search_files":
					result, toolErr = tools.ToolSearchFiles(ctx, args)
				case "run_shell":
					result, toolErr = tools.ToolRunShell(ctx, args, c.Settings.ShellWhitelist)
				default:
					result = map[string]interface{}{"error": "Unknown function"}
				}
			} else {
				fmt.Printf("   🚫 Execução negada pelo usuário.\n")
				result = map[string]interface{}{"error": "User denied execution of this tool."}
			}

			if toolErr != nil {
				result = map[string]interface{}{"error": toolErr.Error()}
			}

			functionResponses = append(functionResponses, map[string]interface{}{
				"functionResponse": map[string]interface{}{
					"name":     name,
					"response": result,
				},
			})
		}

		// Add function responses to history
		c.History = append(c.History, map[string]interface{}{
			"role":  "function", 
			"parts": functionResponses,
		})
		
		// Continue loop to send results back to model
	}
}

type rawGenerateResponse struct {
	Candidates []struct {
		Content map[string]interface{} `json:"content"`
	} `json:"candidates"`
}

func (c *RawClient) generateContent(ctx context.Context) (*rawGenerateResponse, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Model, c.APIKey)

	payload := map[string]interface{}{
		"contents": c.History,
		"tools":    c.Tools,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var genResp rawGenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return nil, err
	}

	return &genResp, nil
}
