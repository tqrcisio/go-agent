package agents

import (
	"context"
	"fmt"
	"go-agent/internal/state"
	"os/exec"
)

type GitAnalysisAgent struct{}

func NewGitAnalysisAgent() *GitAnalysisAgent {
	return &GitAnalysisAgent{}
}

func (a *GitAnalysisAgent) Name() string {
	return "GitAnalysisAgent"
}

func (a *GitAnalysisAgent) Run(ctx context.Context, s *state.State) error {
	// We want to get a sense of recent activity and churn.
	// git log --stat -n 10 gives us the last 10 commits with changed files stats.
	cmd := exec.CommandContext(ctx, "git", "-C", s.RepoPath, "log", "--stat", "-n", "10")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If git command fails (e.g. not a git repo?), we shouldn't fail the whole pipeline,
		// just log it and proceed with empty history.
		s.GitHistory = fmt.Sprintf("Could not retrieve git history: %v", err)
		return nil 
	}

	s.GitHistory = string(output)
	return nil
}
