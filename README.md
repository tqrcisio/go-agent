# BugScan (Multi-Agent Code Analyzer)

BugScan is a powerful, multi-language command-line tool written in Go that leverages Google's Gemini AI to analyze GitHub repositories for potential bugs, vulnerabilities, and code quality issues. It employs a **Multi-Agent Architecture** orchestrated by a central pipeline to first understand the project's context and then perform a deep, tailored analysis.

## Features

- **Multi-Agent Architecture**: A modular pipeline of specialized agents:
  - **RepoCloneAgent**: Clones or updates the target repository.
  - **FileStructureAgent**: Maps the project's file structure.
  - **IdentifyProjectAgent**: Uses AI to identify the language and define an analysis strategy.
  - **ScanFilesAgent**: Locates relevant source code files based on the identified strategy.
  - **ChunkingAgent**: Splits files into manageable chunks for AI processing.
  - **AnalyzeChunksAgent**: Performs deep, concurrent code analysis using Gemini AI.
  - **DeduplicateFindingsAgent**: Removes duplicate issues from the findings.
  - **PrioritizeFindingsAgent**: Sorts findings by severity (High > Medium > Low).
  - **ReportAgent**: Generates the final Markdown report.
- **Multi-Language Support**: Automatically detects and analyzes projects in various languages (Go, Python, JavaScript, TypeScript, etc.).
- **Automated Repository Cloning**: Seamless cloning of GitHub repositories to a temporary cache.
- **Smart Code Scanning**: Dynamically targets relevant file extensions while ignoring noise.
- **Concurrent Processing**: Employs a worker pool pattern to analyze code chunks in parallel.
- **Comprehensive Reporting**: Generates a detailed Markdown report (`report.md`).

## Project Structure

```
.
├── cmd/
│   └── go-agent/
│       └── main.go           # Application entry point
├── internal/
│   ├── agents/               # Individual agent implementations
│   │   ├── analyze.go
│   │   ├── chunk.go
│   │   ├── clone.go
│   │   ├── deduplicate.go
│   │   ├── identify.go
│   │   ├── prioritize.go
│   │   ├── report.go
│   │   ├── scan.go
│   │   └── structure.go
│   ├── orchestrator/
│   │   └── orchestrator.go   # Manages agent execution flow
│   ├── state/
│   │   └── state.go          # Shared state passed between agents
│   ├── chunker/
│   │   └── chunker.go
│   ├── geminiclient/
│   │   └── geminiclient.go
│   ├── repomanager/
│   │   └── repomanager.go
│   ├── report/
│   │   └── report.go
│   └── scanner/
│       └── scanner.go
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
└── README.md                 # Project documentation
```

## Recent Improvements

Based on a recent architectural review, several key improvements have been implemented to enhance the robustness and reliability of the agent pipeline:

- **Robust Error Handling**: The `AnalyzeChunksAgent` now captures and logs all errors, ensuring that analysis failures are visible and the final report's reliability can be assessed.
- **Responsive Cancellation**: Worker goroutines now respond immediately to context cancellation, preventing them from getting stuck in a blocked state.
- **Resilient JSON Parsing**: The logic for parsing JSON from the Gemini API has been hardened to handle variations in markdown formatting.
- **Large File Support**: The file chunker now uses a larger buffer to handle files with very long lines without silent failures.
- **Correct Repo Naming**: The repository cloning logic now correctly trims the `.git` suffix from folder names.
- **Input Validation**: The project identification step now validates the response from the LLM to ensure essential data is present.
- **Code Cleanup**: Removed unused code to improve maintainability.

## Prerequisites

- **Go**: Ensure you have Go installed (version 1.21 or later recommended).
- **Gemini API Key**: You need a valid API key from Google AI Studio.

## Setup & Configuration

1.  **Clone the project** (if you haven't already):
    ```bash
    git clone <this-repo-url>
    cd go-agent
    ```

2.  **Install dependencies**:
    ```bash
    go mod download
    ```

3.  **Configure Environment Variables**:
    Create a `.env` file in the root directory and add your Gemini API key:
    ```env
    GEMINI_API_KEY=your_actual_api_key_here
    ```

4.  **Build the project** (optional):
    If you prefer to have a compiled binary, you can build the project:
    ```bash
    go build -o go-agent ./cmd/go-agent
    ```
    This will create an executable named `go-agent` in the project root.

## Usage

You can run the application using `go run` or by using the compiled binary if you built it.

**Using `go run`:**
```bash
go run ./cmd/go-agent analyze https://github.com/username/repo-name
```

**Using the compiled binary:**
```bash
./go-agent analyze https://github.com/username/repo-name
```

### Examples

**Analyze a Python project:**
```bash
go run ./cmd/go-agent analyze https://github.com/tqrcisio/document-parser-api
```

### Output Example

The tool provides real-time feedback on the progress of each agent:

```text
2025/12/16 13:01:00 [RepoCloneAgent] Starting...
Cloning repository document-parser-api...
2025/12/16 13:01:01 [RepoCloneAgent] Finished in 946.27ms
2025/12/16 13:01:01 [IdentifyProjectAgent] Starting...
2025/12/16 13:01:09 [IdentifyProjectAgent] Finished in 8.54s
...
2025/12/16 13:01:09 [AnalyzeChunksAgent] Starting...
2025/12/16 13:01:48 [AnalyzeChunksAgent] Finished in 39.44s
...
Report successfully generated: report.md
```

## Technical Details

- **Orchestration**: A central `Orchestrator` runs a sequence of `Agents`. Each agent performs a specific task and updates a shared `State` object.
- **Concurrency**: The `AnalyzeChunksAgent` uses a worker pool (default 10 workers) to send requests to Gemini concurrently.
- **Context Awareness**: The `IdentifyProjectAgent` builds a "Project Context" object containing the language, target file extensions, and a specific "persona" for the reviewer to adopt.
- **Cache**: Repositories are cloned to `~/.bugscan/repos` to avoid re-downloading.
