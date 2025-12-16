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
	"sync"

	"github.com/joho/godotenv"
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
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
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

		ctx := context.Background()
		client, err := geminiclient.New(ctx)
		if err != nil {
			log.Fatalf("Error creating Gemini client: %v", err)
		}

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
					resp, err := client.AnalyzeChunk(ctx, chunk)
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

		fmt.Println(reportMsg)
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
