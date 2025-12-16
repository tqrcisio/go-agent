package main

import (
	"fmt"
	"go-agent/internal/chunker"
	"go-agent/internal/geminiclient"
	"go-agent/internal/repomanager"
	"go-agent/internal/scanner"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bugscan",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [repo-url]",
	Short: "Analyze a GitHub repository",
	Long:  `Analyzes a GitHub repository for potential bugs and vulnerabilities.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Set this to true to use mocked API responses
		os.Setenv("MOCK_API", "true")
		os.Setenv("GEMINI_API_KEY", "your-api-key-here") // Set a dummy key

		repoURL := args[0]
		fmt.Println("Analyzing repository:", repoURL)

		repoPath, err := repomanager.Clone(repoURL)
		if err != nil {
			fmt.Printf("Error cloning repository: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Repository cloned to: %s\n", repoPath)

		goFiles, err := scanner.Scan(repoPath)
		if err != nil {
			fmt.Printf("Error scanning for Go files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d Go files.\n", len(goFiles))

		var allChunks []chunker.CodeChunk
		for _, file := range goFiles {
			chunks, err := chunker.ChunkFile(file)
			if err != nil {
				fmt.Printf("Error chunking file %s: %v\n", file, err)
				continue
			}
			allChunks = append(allChunks, chunks...)
		}
		fmt.Printf("Total chunks created: %d. Starting analysis...\n", len(allChunks))

		var allFindings []geminiclient.Finding
		for _, chunk := range allChunks {
			resp, err := geminiclient.AnalyzeChunk(chunk)
			if err != nil {
				fmt.Printf("Error analyzing chunk %s:%d-%d: %v\n", chunk.FilePath, chunk.StartLine, chunk.EndLine, err)
				continue
			}
			if resp != nil && len(resp.Issues) > 0 {
				allFindings = append(allFindings, resp.Issues...)
			}
		}

		fmt.Println("\n--- Analysis Complete ---")
		if len(allFindings) == 0 {
			fmt.Println("No issues found.")
			return
		}

		fmt.Printf("Found %d potential issues:\n", len(allFindings))
		for _, finding := range allFindings {
			fmt.Printf("- [%s] %s:%d: %s\n", finding.Severity, finding.File, finding.Line, finding.Description)
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
