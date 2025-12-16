package agents

import (
	"context"
	"fmt"
	"go-agent/internal/report"
	"go-agent/internal/state"
)

type ReportAgent struct{}

func NewReportAgent() *ReportAgent {
	return &ReportAgent{}
}

func (a *ReportAgent) Name() string {
	return "ReportAgent"
}

func (a *ReportAgent) Run(ctx context.Context, s *state.State) error {
	reportMsg, err := report.GenerateMarkdown(s.Findings, s.RepoURL)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}
	fmt.Println(reportMsg)
	return nil
}
