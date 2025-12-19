package agents

import (
	"context"
	"fmt"
	"go-agent/internal/state"
	"os"
	"path/filepath"
	"strings"
)

type FileStructureAgent struct{}

func NewFileStructureAgent() *FileStructureAgent {
	return &FileStructureAgent{}
}

func (a *FileStructureAgent) Name() string {
	return "FileStructureAgent"
}

func (a *FileStructureAgent) Run(ctx context.Context, s *state.State) error {
	structure, err := getFileStructure(s.RepoPath)
	if err != nil {
		return fmt.Errorf("failed to get file structure: %w", err)
	}
	s.FileStructure = structure
	return nil
}

// getFileStructure returns a string representation of the file structure (up to a limit).
// Copied and adapted from main.go
func getFileStructure(rootDir string) (string, error) {
	var structure strings.Builder
	fileCount := 0
	maxFiles := 100

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileCount >= maxFiles {
			return filepath.SkipDir
		}
		
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			// Ignore hidden dirs and common dependency/build dirs
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			switch name {
			case "node_modules", "vendor", "dist", "build", "coverage", "public", "assets":
				return filepath.SkipDir
			}
			
			structure.WriteString(fmt.Sprintf("%s/\n", relPath))
		} else {
			structure.WriteString(fmt.Sprintf("%s\n", relPath))
			fileCount++
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return structure.String(), nil
}
