package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ToolListFiles lists files in a directory.
// Arguments: "path" (string) - relative path from project root.
func ToolListFiles(args map[string]interface{}) (map[string]interface{}, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		path = "."
	}

	// Basic security check: prevent escaping root with ".."
	if strings.Contains(path, "..") {
		return map[string]interface{}{"error": "Access denied: '..' is not allowed"}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Failed to read dir: %v", err)}, nil
	}

	var files []interface{}
	var dirs []interface{}

	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name()+"/")
		} else {
			files = append(files, e.Name())
		}
	}

	return map[string]interface{}{
		"path":        path,
		"directories": dirs,
		"files":       files,
	}, nil
}

// ToolReadFile reads the content of a file.
// Arguments: "path" (string).
func ToolReadFile(args map[string]interface{}) (map[string]interface{}, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return map[string]interface{}{"error": "Missing 'path' argument"}, nil
	}

	if strings.Contains(path, "..") {
		return map[string]interface{}{"error": "Access denied: '..' is not allowed"}, nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return map[string]interface{}{"error": fmt.Sprintf("Failed to read file: %v", err)}, nil
	}

	// Truncate if too large to save tokens (limit to ~2000 lines approx 100kb for now)
	const maxSize = 100 * 1024
	truncated := false
	if len(content) > maxSize {
		content = content[:maxSize]
		truncated = true
	}

	return map[string]interface{}{
		"path":      path,
		"content":   string(content),
		"truncated": truncated,
	}, nil
}

// ToolSearchFiles searches for a text pattern using grep.
// Arguments: "pattern" (string), "path" (string - optional, default ".").
func ToolSearchFiles(args map[string]interface{}) (map[string]interface{}, error) {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return map[string]interface{}{"error": "Missing 'pattern' argument"}, nil
	}

	path, ok := args["path"].(string)
	if !ok || path == "" {
		path = "."
	}

    if strings.Contains(path, "..") {
		return map[string]interface{}{"error": "Access denied: '..' is not allowed"}, nil
	}

	// Using grep -rn (recursive, line number)
	// Limiting to 50 matches to avoid token explosion
	cmd := exec.Command("grep", "-rn", "--max-count=50", "--exclude-dir=.git", pattern, path)
	output, err := cmd.CombinedOutput()

	// grep returns exit code 1 if no matches found, which is not an "error" for us
	if err != nil && cmd.ProcessState.ExitCode() != 1 {
		return map[string]interface{}{"error": fmt.Sprintf("Grep failed: %v", err)}, nil
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 50 {
		lines = lines[:50]
		lines = append(lines, "... (matches truncated)")
	}

	return map[string]interface{}{
		"matches": strings.Join(lines, "\n"),
	}, nil
}