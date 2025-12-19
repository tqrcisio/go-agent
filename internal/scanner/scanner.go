package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

// Scan walks the given root directory and returns a slice of paths to files matching the provided extensions.
func Scan(rootDir string, extensions []string, sourceDirs []string, excludePatterns []string) ([]string, error) {
	var matchedFiles []string
	ignoredDirs := map[string]bool{
		".git":         true,
		"vendor":       true,
		"node_modules": true,
		"dist":         true,
		"build":        true,
		".next":        true,
		".nuxt":        true,
	}

	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[ext] = true
	}

	// Determine where to start walking
	walkRoots := []string{rootDir}
	if len(sourceDirs) > 0 {
		var validRoots []string
		for _, sd := range sourceDirs {
			fullPath := filepath.Join(rootDir, sd)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				validRoots = append(validRoots, fullPath)
			}
		}
		if len(validRoots) > 0 {
			walkRoots = validRoots
		}
	}

	for _, startPath := range walkRoots {
		err := filepath.Walk(startPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, _ := filepath.Rel(rootDir, path)

			// Skip ignored directories
			if info.IsDir() {
				if ignoredDirs[info.Name()] {
					return filepath.SkipDir
				}
				return nil
			}

			// Check custom exclude patterns (simple glob match)
			for _, pattern := range excludePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					return nil
				}
				// Also check relative path for patterns like "**/test/**"
				if strings.Contains(relPath, pattern) {
					return nil
				}
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
	}

	return matchedFiles, nil
}
