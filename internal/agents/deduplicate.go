package agents

import (
	"context"
	"fmt"
	"go-agent/internal/geminiclient"
	"go-agent/internal/state"
)

type DeduplicateFindingsAgent struct{}

func NewDeduplicateFindingsAgent() *DeduplicateFindingsAgent {
	return &DeduplicateFindingsAgent{}
}

func (a *DeduplicateFindingsAgent) Name() string {
	return "DeduplicateFindingsAgent"
}

func (a *DeduplicateFindingsAgent) Run(ctx context.Context, s *state.State) error {
	seen := make(map[string]bool)
	uniqueFindings := []geminiclient.Finding{}

	for _, finding := range s.Findings {
		// Create a unique key for each finding
		key := fmt.Sprintf("%s|%d|%s|%s", finding.File, finding.Line, finding.Severity, finding.Description)
		
		if !seen[key] {
			seen[key] = true
			uniqueFindings = append(uniqueFindings, finding)
		}
	}

	s.Findings = uniqueFindings
	return nil
}
