package geminiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"go-agent/internal/chunker"
	"log"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// ProjectInfo represents the identified context of the project.
type ProjectInfo struct {
	Language       string   `json:"language"`
	FileExtensions []string `json:"file_extensions"`
	AnalysisGoal   string   `json:"analysis_goal"`

	PromptTokens     int `json:"-"`
	CandidatesTokens int `json:"-"`
}

// Finding represents a single issue found by the Gemini analysis.
type Finding struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
}

// GeminiResponse is the expected structure of the JSON response from Gemini.
type GeminiResponse struct {
	Issues           []Finding `json:"issues"`
	PromptTokens     int       `json:"-"`
	CandidatesTokens int       `json:"-"`
}

// GeminiClient is a client for the Gemini API.
type GeminiClient struct {
	model *genai.GenerativeModel
	Debug bool
}

// New creates a new GeminiClient.
func New(ctx context.Context, debug bool) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating genai client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-pro")
	return &GeminiClient{model: model, Debug: debug}, nil
}

// IdentifyProject analyzes the file list to determine the technology stack.
func (c *GeminiClient) IdentifyProject(ctx context.Context, fileStructure string, gitHistory string) (*ProjectInfo, error) {
	prompt := fmt.Sprintf(`
You are a generic Software Architect.
Analyze the following file structure and recent git history of a software project and identify:
1. The primary programming language.
2. The file extensions that contain the source code (e.g., .go, .js, .py, .ts).
3. A specific "persona" or goal for a code reviewer analyzing this project (e.g., "You are a senior React engineer looking for state management issues", "You are a Python expert looking for typing issues").

Return the result in raw JSON format with keys: "language", "file_extensions" (array of strings), and "analysis_goal".

File Structure:
---
%s
---

Recent Git History:
---
%s
---
`, fileStructure, gitHistory)

	if c.Debug {
		log.Printf("[DEBUG] IdentifyProject Prompt:\n%s\n", prompt)
	}

	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("error identifying project: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content received from Gemini API during identification")
	}

	part := resp.Candidates[0].Content.Parts[0]
	var info ProjectInfo

	if textPart, ok := part.(genai.Text); ok {
		if c.Debug {
			log.Printf("[DEBUG] IdentifyProject Response:\n%s\n", string(textPart))
		}
		jsonString, err := extractJSON(string(textPart))
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(jsonString), &info)
		if err != nil {
			log.Printf("Error unmarshalling identification response: %v. Response: %s", err, jsonString)
			return nil, err
		}
	}

	if resp.UsageMetadata != nil {
		info.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		info.CandidatesTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	return &info, nil
}

// AnalyzeChunk sends a CodeChunk to the Gemini API for analysis using the identified context.
func (c *GeminiClient) AnalyzeChunk(ctx context.Context, chunk chunker.CodeChunk, contextInfo ProjectInfo) (*GeminiResponse, error) {
	prompt := fmt.Sprintf(`
%s
Analyze the following %s code from file '%s' and identify:
- potential bugs
- security vulnerabilities
- idiomatic issues specific to the language
- performance bottlenecks

Return findings in a raw JSON format, without any markdown formatting. The JSON object should have a single key "issues" which is an array of objects. Each object should have "severity" (HIGH, MEDIUM, LOW), "description", "file", and "line".

Here is the code:
---
%s
---
`, contextInfo.AnalysisGoal, contextInfo.Language, chunk.FilePath, chunk.Content)

	if c.Debug {
		log.Printf("[DEBUG] AnalyzeChunk Prompt for %s:\n%s\n", chunk.FilePath, prompt)
	}

	resp, err := c.model.GenerateContent(ctx,
		genai.Text(prompt),
	)
	if err != nil {
		log.Printf("Error generating content: %v", err)
		return nil, fmt.Errorf("error generating content: %w", err)
	}

	geminiResp := GeminiResponse{}
	if resp.UsageMetadata != nil {
		geminiResp.PromptTokens = int(resp.UsageMetadata.PromptTokenCount)
		geminiResp.CandidatesTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		log.Println("No content received from Gemini API")
		return &geminiResp, nil // Return with token usage but no issues
	}

	part := resp.Candidates[0].Content.Parts[0]

	if textPart, ok := part.(genai.Text); ok {
		if c.Debug {
			log.Printf("[DEBUG] AnalyzeChunk Response for %s:\n%s\n", chunk.FilePath, string(textPart))
		}
		jsonString, err := extractJSON(string(textPart))
		if err != nil {
			// It's possible for the model to return no issues, which is not an error.
			// We'll log it for debugging but return an empty response.
			log.Printf("Could not extract JSON from chunk analysis, maybe no issues found: %v. Raw response: %s", err, string(textPart))
			return &geminiResp, nil
		}

		err = json.Unmarshal([]byte(jsonString), &geminiResp)
		if err != nil {
			log.Printf("Error unmarshalling analysis response: %v. Cleaned response: %s", err, jsonString)
			return nil, err
		}
	}

	return &geminiResp, nil
}

// extractJSON robustly extracts a JSON object from a string that might contain markdown.
func extractJSON(s string) (string, error) {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start == -1 || end == -1 || end <= start {
		return "", fmt.Errorf("no valid JSON object found in the string")
	}

	return s[start : end+1], nil
}
