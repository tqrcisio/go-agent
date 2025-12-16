package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// Scan walks the given root directory and returns a slice of paths to .go files.
// It ignores specified directories like .git and vendor.
func Scan(rootDir string) ([]string, error) {
	var goFiles []string
	ignoredDirs := map[string]bool{
		".git":   true,
		"vendor": true,
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the directory should be ignored
		if info.IsDir() && ignoredDirs[info.Name()] {
			return filepath.SkipDir // Skip this directory
		}

		// Check if it's a Go file
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			goFiles = append(goFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return goFiles, nil
}
