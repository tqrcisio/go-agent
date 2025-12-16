package geminiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"go-agent/internal/chunker"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Finding represents a single issue found by the Gemini analysis.
type Finding struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

// GeminiResponse is the expected structure of the JSON response from Gemini.
type GeminiResponse struct {
	Issues []Finding `json:"issues"`
}

// GeminiClient is a client for the Gemini API.
type GeminiClient struct {
	model *genai.GenerativeModel
}

// New creates a new GeminiClient.
func New(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-pro")
	return &GeminiClient{model: model}, nil
}

// AnalyzeChunk sends a CodeChunk to the Gemini API for analysis.
func (c *GeminiClient) AnalyzeChunk(ctx context.Context, chunk chunker.CodeChunk) (*GeminiResponse, error) {
	prompt := fmt.Sprintf(`
You are a senior Go engineer.
Analyze the following Go code from file '%s' and identify:
- potential bugs
- race conditions
- nil pointer risks
- incorrect error handling
Return findings in a raw JSON format, without any markdown formatting. The JSON object should have a single key "issues" which is an array of objects. Each object should have "severity", "description", "file", and "line".

Here is the code:
---
%s
---
`, chunk.FilePath, chunk.Content)

	resp, err := c.model.GenerateContent(ctx,
		genai.Text(prompt),
	)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return nil, fmt.Errorf("error generating content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No content received from Gemini API")
		return &GeminiResponse{Issues: []Finding{}}, nil // Return empty response
	}

	part := resp.Candidates[0].Content.Parts[0]

	var geminiResp GeminiResponse
	if textPart, ok := part.(genai.Text); ok {
		// The response is expected to be a text part containing JSON
		err = json.Unmarshal([]byte(textPart), &geminiResp)
		if err != nil {
			log.Printf("Error unmarshalling JSON response: %v. Response string: %s", err, textPart)
			return nil, fmt.Errorf("error unmarshalling json response: %w", err)
		}
	} else {
		log.Println("Unexpected response format from Gemini API")
		return nil, fmt.Errorf("unexpected response format from gemini api")
	}

	return &geminiResp, nil
}
