package chunker

import (
	"os"
	"regexp"
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

var vueStyleRegex = regexp.MustCompile(`(?s)<style[^>]*>.*?</style>`)

func cleanContent(content string, filePath string) string {
	if strings.HasSuffix(filePath, ".vue") {
		return vueStyleRegex.ReplaceAllString(content, "<style scoped>\n/* Style removed for analysis efficiency */\n</style>")
	}
	return content
}

// ChunkFile reads a file and splits its content into multiple CodeChunks.
func ChunkFile(filePath string) ([]CodeChunk, error) {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(contentBytes)
	content = cleanContent(content, filePath)

	var chunks []CodeChunk
	lines := strings.Split(content, "\n")
	
	totalLines := len(lines)
	if totalLines == 0 {
		return chunks, nil
	}

	for i := 0; i < totalLines; i += maxLinesPerChunk {
		end := i + maxLinesPerChunk
		if end > totalLines {
			end = totalLines
		}
		
		chunkLines := lines[i:end]
		chunks = append(chunks, CodeChunk{
			FilePath:  filePath,
			StartLine: i + 1,
			EndLine:   end,
			Content:   strings.Join(chunkLines, "\n"),
		})
	}

	return chunks, nil
}
