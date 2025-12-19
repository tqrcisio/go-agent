package tools

import (
	"os"
	"path/filepath"
)

// ProjectFilesCache stores a list of relative file paths for autocomplete.
type ProjectFilesCache struct {
	Files []string
}

// NewProjectFilesCache scans the directory and builds a cache of files.
func NewProjectFilesCache(rootDir string) (*ProjectFilesCache, error) {
	var files []string
	ignoredDirs := map[string]bool{
		".git":         true,
		"vendor":       true,
		"node_modules": true,
		"dist":         true,
		"build":        true,
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip ignored directories
		if info.IsDir() {
			if ignoredDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		if relPath == "." {
			return nil
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ProjectFilesCache{Files: files}, nil
}
