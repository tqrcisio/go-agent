package agents

import (
	"context"
	"fmt"
	"go-agent/internal/chunker"
	"go-agent/internal/geminiclient"
	"go-agent/internal/state"
	"sync"
)

type AnalyzeChunksAgent struct {
	client *geminiclient.GeminiClient
}

func NewAnalyzeChunksAgent(client *geminiclient.GeminiClient) *AnalyzeChunksAgent {
	return &AnalyzeChunksAgent{client: client}
}

func (a *AnalyzeChunksAgent) Name() string {
	return "AnalyzeChunksAgent"
}

func (a *AnalyzeChunksAgent) Run(ctx context.Context, s *state.State) error {
	if s.ProjectInfo == nil {
		return fmt.Errorf("project info is missing")
	}

	var wg sync.WaitGroup
	chunkChan := make(chan chunker.CodeChunk, len(s.Chunks))
	findingChan := make(chan []geminiclient.Finding, len(s.Chunks))
	
	// Start workers
	numWorkers := 10 // Could be configurable
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				// Check for cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				resp, err := a.client.AnalyzeChunk(ctx, chunk, *s.ProjectInfo)
				if err != nil {
					// In a real agent, we might want to collect these errors
					continue
				}
				if resp != nil && len(resp.Issues) > 0 {
					findingChan <- resp.Issues
				}
			}
		}()
	}

	// Send chunks to workers
	for _, chunk := range s.Chunks {
		chunkChan <- chunk
	}
	close(chunkChan)

	// Wait for workers to finish and collect findings
	go func() {
		wg.Wait()
		close(findingChan)
	}()

	for findings := range findingChan {
		s.Findings = append(s.Findings, findings...)
	}

	return nil
}
