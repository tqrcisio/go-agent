package geminiclient

import (
	"context"
	"github.com/google/generative-ai-go/genai"
)

// ChatSession wraps the RawClient for compatibility.
type ChatSession struct {
	RawClient *RawClient
	Client    *GeminiClient
}

// StartChat initializes a new chat session with tools configured using RawClient.
func (c *GeminiClient) StartChat() *ChatSession {
	// Initialize RawClient
	raw := NewRawClient(c.Debug)
	return &ChatSession{
		RawClient: raw,
		Client:    c,
	}
}

// SendMessage sends a message to the chat and handles tool calls loop via RawClient.
func (s *ChatSession) SendMessage(ctx context.Context, msg string) (string, error) {
	return s.RawClient.SendMessageRaw(ctx, msg)
}

// History returns the chat history.
// Note: RawClient uses a different history format (JSON maps), so this returns nil or empty for now
// to preserve interface compatibility without complex conversion.
func (s *ChatSession) History() []*genai.Content {
	return []*genai.Content{}
}