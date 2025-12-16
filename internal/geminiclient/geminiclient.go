package geminiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"go-agent/internal/chunker"
	"os"
	"time"
	"log"

	"cloud.google.com/go/vertexai/apiv1"
	"cloud.google.com/go/vertexai/apiv1/vertexaipb"
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

// AnalyzeChunk sends a CodeChunk to the Gemini API for analysis.
// For now, it returns a mocked response.
func AnalyzeChunk(chunk chunker.CodeChunk) (*GeminiResponse, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// Mocked response for development
	// In the next step, we will replace this with a real API call.
	if os.Getenv("MOCK_API") == "true" {
		fmt.Printf("Analyzing chunk %s:%d-%d (mocked)...\n", chunk.FilePath, chunk.StartLine, chunk.EndLine)
		time.Sleep(100 * time.Millisecond) // Simulate network latency
		mockResponse := &GeminiResponse{
			Issues: []Finding{
				{
					Severity:    "HIGH",
					Description: "Potential nil pointer dereference detected.",
					File:        chunk.FilePath,
					Line:        chunk.StartLine + 5, // Example line
				},
			},
		}
		return mockResponse, nil
	}

	return callGeminiAPI(chunk)
}


func callGeminiAPI(chunk chunker.CodeChunk) (*GeminiResponse, error) {
	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")
	location := "us-central1" // Or your desired location
	modelName := "gemini-1.5-pro-preview-0409"   // Or your desired model

	client, err := vertexai.NewPredictionClient(ctx)
	if err != nil {
		log.Printf("Error creating Vertex AI client: %v", err)
		return nil, fmt.Errorf("error creating vertex ai client: %w", err)
	}
	defer client.Close()

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

	req := &vertexaipb.GenerateContentRequest{
		Model: fmt.Sprintf("projects/%s/locations/%s/publishers/google/models/%s", projectID, location, modelName),
		Contents: []*vertexaipb.Content{
			{
				Role: "user",
				Parts: []*vertexaipb.Part{
					{
						Data: &vertexaipb.Part_Text{
							Text: prompt,
						},
					},
				},
			},
		},
	}

	resp, err := client.GenerateContent(ctx, req)
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
	// The response is expected to be a text part containing JSON
	if textPart, ok := part.GetData().(*vertexaipb.Part_Text); ok {
		// Clean the response to extract pure JSON
		jsonStr := strings.Trim(textPart.Text, "```json\n")
		
		err = json.Unmarshal([]byte(jsonStr), &geminiResp)
		if err != nil {
			log.Printf("Error unmarshalling JSON response: %v. Response string: %s", err, jsonStr)
			return nil, fmt.Errorf("error unmarshalling json response: %w", err)
		}
	} else {
		log.Println("Unexpected response format from Gemini API")
		return nil, fmt.Errorf("unexpected response format from gemini api")
	}

	return &geminiResp, nil
}
