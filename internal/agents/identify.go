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
	projectCtx, err := a.client.IdentifyProject(ctx, s.FileStructure)
	if err != nil {
		return fmt.Errorf("failed to identify project: %w", err)
	}
	s.ProjectInfo = projectCtx
	return nil
}
