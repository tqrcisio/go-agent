package tui

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func processFilesContext(ctx context.Context, input string) string {
	re := regexp.MustCompile(`@([^
]+)`)
	matches := re.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return input
	}

	var sb strings.Builder
	sb.WriteString(input)
	sb.WriteString("\n\n---\nContexto de arquivos selecionados:\n")

	processed := make(map[string]bool)
	for _, match := range matches {
		if ctx.Err() != nil {
			break
		}

		fileName := match[1]
		if processed[fileName] {
			continue
		}

		// 1. Basic security and existence checks
		info, err := os.Stat(fileName)
		if err != nil {
			// For TUI, maybe we should return a warning message? 
			// For now, just inline it like the original code.
			sb.WriteString(fmt.Sprintf("\n(⚠️ File not found: %s)\n", fileName))
			continue
		}

		if info.IsDir() {
			sb.WriteString(fmt.Sprintf("\n(⚠️ Skipping directory: %s)\n", fileName))
			continue
		}

		// 2. Size limit check
		if info.Size() > 100*1024 { // 100KB limit
			sb.WriteString(fmt.Sprintf("\n(⚠️ File too large >100KB: %s)\n", fileName))
			continue
		}

		// 3. Read and Binary check
		content, err := os.ReadFile(fileName)
		if err != nil {
			sb.WriteString(fmt.Sprintf("\n(⚠️ Error reading file: %s)\n", fileName))
			continue
		}

		if isBinary(content) {
			sb.WriteString(fmt.Sprintf("\n(⚠️ Binary file detected: %s)\n", fileName))
			continue
		}

		sb.WriteString(fmt.Sprintf("\nArquivo: %s\n```\n%s\n```\n", fileName, string(content)))
		processed[fileName] = true
	}

	return sb.String()
}

func isBinary(content []byte) bool {
	limit := 1024
	if len(content) < limit {
		limit = len(content)
	}
	for i := 0; i < limit; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}
