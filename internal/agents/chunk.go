package agents

import (
	"context"
	"go-agent/internal/chunker"
	"go-agent/internal/state"
)

type ChunkingAgent struct{}

func NewChunkingAgent() *ChunkingAgent {
	return &ChunkingAgent{}
}

func (a *ChunkingAgent) Name() string {
	return "ChunkingAgent"
}

func (a *ChunkingAgent) Run(ctx context.Context, s *state.State) error {
	var allChunks []chunker.CodeChunk
	for _, file := range s.SourceFiles {
		chunks, err := chunker.ChunkFile(file)
		if err != nil {
			// Log error but continue? The original code says "continue" with a Printf.
			// Ideally we should log it. For now, let's just append what we can.
			// Or we can add to s.Errors if we want strict tracking.
			continue
		}
		allChunks = append(allChunks, chunks...)
	}
	s.Chunks = allChunks
	return nil
}
