package repomanager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Clone clones a repository from the given URL into a local cache directory.
// It returns the path to the cloned repository.
func Clone(repoURL string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".bugscan", "repos")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("could not create cache directory: %w", err)
	}

	// Basic way to get a repo name from URL. This can be improved.
	repoName := filepath.Base(repoURL)
	repoPath := filepath.Join(cacheDir, repoName)

	// If the repo already exists, pull the latest changes.
	if _, err := os.Stat(repoPath); err == nil {
		fmt.Printf("Repository %s already exists. Pulling latest changes...\n", repoName)
		cmd := exec.Command("git", "-C", repoPath, "pull")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("could not pull repository: %w", err)
		}
	} else {
		// Otherwise, clone it.
		fmt.Printf("Cloning repository %s...\n", repoName)
		cmd := exec.Command("git", "clone", repoURL, repoPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("could not clone repository: %w", err)
		}
	}

	return repoPath, nil
}
