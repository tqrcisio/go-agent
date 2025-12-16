package agents

import (
	"context"
	"fmt"
	"go-agent/internal/geminiclient"
	"go-agent/internal/state"
)

type IdentifyProjectAgent struct {
	client *geminiclient.GeminiClient
}

func NewIdentifyProjectAgent(client *geminiclient.GeminiClient) *IdentifyProjectAgent {
	return &IdentifyProjectAgent{client: client}
}

func (a *IdentifyProjectAgent) Name() string {
	return "IdentifyProjectAgent"
}

func (a *IdentifyProjectAgent) Run(ctx context.Context, s *state.State) error {
	projectCtx, err := a.client.IdentifyProject(ctx, s.FileStructure, s.GitHistory)
	if err != nil {
		return fmt.Errorf("failed to identify project: %w", err)
	}

	if projectCtx == nil {
		return fmt.Errorf("project identification returned nil info")
	}
	if projectCtx.Language == "" {
		return fmt.Errorf("identified project language is empty")
	}
	if len(projectCtx.FileExtensions) == 0 {
		return fmt.Errorf("identified project file extensions are empty")
	}

	s.ProjectInfo = projectCtx
	return nil
}
