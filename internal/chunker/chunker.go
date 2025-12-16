package chunker

import (
	"bufio"
	"os"
	"strings"
)

const maxLinesPerChunk = 800

// CodeChunk represents a chunk of code from a file.
type CodeChunk struct {
	FilePath  string
	StartLine int
	EndLine   int
	Content   string
}

// ChunkFile reads a file and splits its content into multiple CodeChunks.
func ChunkFile(filePath string) ([]CodeChunk, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chunks []CodeChunk
	var currentChunk strings.Builder
	var lines []string
	
	scanner := bufio.NewScanner(file)
	startLine := 1
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		lineCount++

		if lineCount >= maxLinesPerChunk {
			chunks = append(chunks, CodeChunk{
				FilePath:  filePath,
				StartLine: startLine,
				EndLine:   startLine + len(lines) - 1,
				Content:   strings.Join(lines, "\n"),
			})
			// Reset for the next chunk
			lines = []string{}
			startLine += lineCount
			lineCount = 0
		}
	}

	// Add the last remaining chunk if any
	if len(lines) > 0 {
		chunks = append(chunks, CodeChunk{
			FilePath:  filePath,
			StartLine: startLine,
			EndLine:   startLine + len(lines) - 1,
			Content:   strings.Join(lines, "\n"),
		})
	}


	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return chunks, nil
}
