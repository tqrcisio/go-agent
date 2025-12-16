package agents

import (
	"context"
	"fmt"
	"go-agent/internal/repomanager"
	"go-agent/internal/state"
)

type RepoCloneAgent struct{}

func NewRepoCloneAgent() *RepoCloneAgent {
	return &RepoCloneAgent{}
}

func (a *RepoCloneAgent) Name() string {
	return "RepoCloneAgent"
}

func (a *RepoCloneAgent) Run(ctx context.Context, s *state.State) error {
	if s.RepoURL == "" {
		return fmt.Errorf("repository URL is empty")
	}

	repoPath, err := repomanager.Clone(s.RepoURL)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	s.RepoPath = repoPath
	return nil
}
