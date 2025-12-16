package state

import (
	"go-agent/internal/chunker"
	"go-agent/internal/geminiclient"
)

// State holds the shared context for the analysis pipeline.
type State struct {
	RepoURL       string
	RepoPath      string
	FileStructure string

	ProjectInfo *geminiclient.ProjectInfo

	SourceFiles []string
	Chunks      []chunker.CodeChunk

	Findings []geminiclient.Finding
	Errors   []error
}
