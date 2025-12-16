package agents

import (
	"context"
	"fmt"
	"go-agent/internal/scanner"
	"go-agent/internal/state"
)

type ScanFilesAgent struct{}

func NewScanFilesAgent() *ScanFilesAgent {
	return &ScanFilesAgent{}
}

func (a *ScanFilesAgent) Name() string {
	return "ScanFilesAgent"
}

func (a *ScanFilesAgent) Run(ctx context.Context, s *state.State) error {
	if s.ProjectInfo == nil {
		return fmt.Errorf("project info is missing")
	}

	sourceFiles, err := scanner.Scan(s.RepoPath, s.ProjectInfo.FileExtensions)
	if err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}
	s.SourceFiles = sourceFiles
	return nil
}
