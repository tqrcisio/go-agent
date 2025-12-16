package scanner

import (
	"os"
	"path/filepath"
)

// Scan walks the given root directory and returns a slice of paths to files matching the provided extensions.
// It ignores specified directories like .git and vendor.
func Scan(rootDir string, extensions []string) ([]string, error) {
	var matchedFiles []string
	ignoredDirs := map[string]bool{
		".git":         true,
		"vendor":       true,
		"node_modules": true,
		"dist":         true,
		"build":        true,
	}

	// Create a map for faster lookup of extensions
	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the directory should be ignored
		if info.IsDir() {
			if ignoredDirs[info.Name()] {
				return filepath.SkipDir // Skip this directory
			}
			return nil
		}

		// Check if it's a matching file
		ext := filepath.Ext(info.Name())
		if extMap[ext] {
			matchedFiles = append(matchedFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return matchedFiles, nil
}
