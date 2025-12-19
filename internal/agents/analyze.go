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
	var mu sync.Mutex // Protect shared state updates
	chunkChan := make(chan chunker.CodeChunk, len(s.Chunks))
	findingChan := make(chan []geminiclient.Finding, len(s.Chunks))
	
	// Start workers
	numWorkers := 10 // Could be configurable
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case chunk, ok := <-chunkChan:
					if !ok {
						return
					}
					// processa chunk
					resp, err := a.client.AnalyzeChunk(ctx, chunk, *s.ProjectInfo)
					if err != nil {
						mu.Lock()
						s.Errors = append(s.Errors,
							fmt.Errorf("analyze failed for %s:%d-%d: %w",
								chunk.FilePath, chunk.StartLine, chunk.EndLine, err))
						mu.Unlock()
						continue
					}
					
					mu.Lock()
					s.TotalPromptTokens += resp.PromptTokens
					s.TotalCandidatesTokens += resp.CandidatesTokens
					mu.Unlock()

					if resp != nil && len(resp.Issues) > 0 {
						findingChan <- resp.Issues
					}
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

	collectFindings(s, findingChan)

	return nil
}

func collectFindings(s *state.State, ch <-chan []geminiclient.Finding) {
	for findings := range ch {
		s.Findings = append(s.Findings, findings...)
	}
}
