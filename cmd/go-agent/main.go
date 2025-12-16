package main

import (
	"context"
	"fmt"
	"go-agent/internal/chunker"
	"go-agent/internal/geminiclient"
	"go-agent/internal/repomanager"
	"go-agent/internal/report"
	"go-agent/internal/scanner"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bugscan",
	Short: "AI-powered multi-language code analysis tool",
	Long: `BugScan is a tool that uses Gemini AI to analyze code repositories.
It first identifies the project's language and context (Agent 1), and then
performs a deep analysis of the source code (Agent 2).`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [repo-url]",
	Short: "Analyze a GitHub repository",
	Long:  `Analyzes a GitHub repository for potential bugs and vulnerabilities using a multi-agent approach.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := godotenv.Load()
		if err != nil {
			log.Println("Warning: .env file not found, relying on environment variables.")
		}

		repoURL := args[0]
		fmt.Println("Analyzing repository:", repoURL)

		repoPath, err := repomanager.Clone(repoURL)
		if err != nil {
			fmt.Printf("Error cloning repository: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(repoPath)
		fmt.Printf("Repository cloned to: %s\n", repoPath)

		// --- Agent 1: The Architect (Context Identification) ---
		fmt.Println("Agent 1 (Architect) is analyzing the project structure...")
		ctx := context.Background()
		client, err := geminiclient.New(ctx)
		if err != nil {
			log.Fatalf("Error creating Gemini client: %v", err)
		}

		fileStructure, err := getFileStructure(repoPath)
		if err != nil {
			log.Fatalf("Error reading file structure: %v", err)
		}

		projectCtx, err := client.IdentifyProject(ctx, fileStructure)
		if err != nil {
			log.Fatalf("Error identifying project context: %v", err)
		}

		fmt.Printf("Identified Project Context:\n")
		fmt.Printf("  Language: %s\n", projectCtx.Language)
		fmt.Printf("  Target Extensions: %v\n", projectCtx.FileExtensions)
		fmt.Printf("  Analysis Goal: %s\n", projectCtx.AnalysisGoal)

		// --- Dynamic Scanning ---
		sourceFiles, err := scanner.Scan(repoPath, projectCtx.FileExtensions)
		if err != nil {
			fmt.Printf("Error scanning for source files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d source files.\n", len(sourceFiles))

		var allChunks []chunker.CodeChunk
		for _, file := range sourceFiles {
			chunks, err := chunker.ChunkFile(file)
			if err != nil {
				fmt.Printf("Error chunking file %s: %v\n", file, err)
				continue
			}
			allChunks = append(allChunks, chunks...)
		}
		fmt.Printf("Total chunks created: %d. Starting deep analysis...\n", len(allChunks))

		// --- Agent 2: The Reviewer (Deep Analysis) ---
		fmt.Println("Agent 2 (Reviewer) is analyzing code chunks...")
		
		var allFindings []geminiclient.Finding
		var wg sync.WaitGroup
		chunkChan := make(chan chunker.CodeChunk, len(allChunks))
		findingChan := make(chan []geminiclient.Finding, len(allChunks))

		// Start workers
		numWorkers := 10 // Adjust as needed
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for chunk := range chunkChan {
					// Pass the project context to the analysis agent
					resp, err := client.AnalyzeChunk(ctx, chunk, *projectCtx)
					if err != nil {
						fmt.Printf("Error analyzing chunk %s:%d-%d: %v\n", chunk.FilePath, chunk.StartLine, chunk.EndLine, err)
						continue
					}
					if resp != nil && len(resp.Issues) > 0 {
						findingChan <- resp.Issues
					}
				}
			}()
		}

		// Send chunks to workers
		for _, chunk := range allChunks {
			chunkChan <- chunk
		}
		close(chunkChan)

		// Wait for workers to finish and collect findings
		go func() {
			wg.Wait()
			close(findingChan)
		}()

		for findings := range findingChan {
			allFindings = append(allFindings, findings...)
		}

		fmt.Println("\n--- Analysis Complete ---")
		reportMsg, err := report.GenerateMarkdown(allFindings, repoURL)
		if err != nil {
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(reportMsg)
	},
}

// getFileStructure returns a string representation of the file structure (up to a limit).
func getFileStructure(rootDir string) (string, error) {
	var structure strings.Builder
	fileCount := 0
	maxFiles := 100

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileCount >= maxFiles {
			return filepath.SkipDir
		}
		
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir // Skip hidden dirs like .git
			}
			structure.WriteString(fmt.Sprintf("%s/\n", relPath))
		} else {
			structure.WriteString(fmt.Sprintf("%s\n", relPath))
			fileCount++
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return structure.String(), nil
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
