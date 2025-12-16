# BugScan (Multi-Agent Code Analyzer)

BugScan is a powerful, multi-language command-line tool written in Go that leverages Google's Gemini AI to analyze GitHub repositories for potential bugs, vulnerabilities, and code quality issues. It employs a **Multi-Agent Architecture** to first understand the project's context and then perform a deep, tailored analysis.

## Features

- **Multi-Agent Architecture**:
  - **Agent 1 (The Architect)**: Analyzes the file structure to identify the programming language (e.g., Python, Go, JavaScript) and determines the best analysis strategy.
  - **Agent 2 (The Reviewer)**: Performs deep code analysis based on the context provided by the Architect.
- **Multi-Language Support**: Automatically detects and analyzes projects in various languages (Go, Python, JavaScript, TypeScript, etc.).
- **Automated Repository Cloning**: Seamless cloning of GitHub repositories to a temporary cache.
- **Smart Code Scanning**: Dynamically targets relevant file extensions while ignoring noise (e.g., `.git`, `node_modules`, `vendor`).
- **Intelligent Chunking**: Splits large files into manageable chunks to ensure optimal processing by the AI model.
- **Concurrent Processing**: Employs a worker pool pattern to analyze code chunks in parallel for high performance.
- **Comprehensive Reporting**: Generates a detailed Markdown report (`report.md`) categorized by severity (High, Medium, Low).

## Project Structure

```
.
├── cmd/
│   └── go-agent/
│       └── main.go           # Application entry point & Agent Orchestrator
├── internal/
│   ├── chunker/
│   │   └── chunker.go        # Logic for splitting files into code chunks
│   ├── geminiclient/
│   │   └── geminiclient.go   # Agent implementations (Architect & Reviewer)
│   ├── repomanager/
│   │   └── repomanager.go    # Handles git cloning and updates
│   ├── report/
│   │   └── report.go         # Generates Markdown reports from findings
│   ├── scanner/
│   │   └── scanner.go        # Dynamic file scanner
│   └── tree/
│       └── tree.go           # (Legacy) Binary search tree implementation
├── .env                      # Environment variables configuration
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
└── README.md                 # Project documentation
```

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

## Usage

To analyze a GitHub repository, run the `analyze` command with the repository URL:

```bash
go run ./cmd/go-agent analyze https://github.com/username/repo-name
```

### Examples

**Analyze a Go project:**
```bash
go run ./cmd/go-agent analyze https://github.com/gin-gonic/gin
```

**Analyze a Python project:**
```bash
go run ./cmd/go-agent analyze https://github.com/tqrcisio/document-parser-api
```

### Output

The tool will:
1.  **Clone** the specified repository.
2.  **Agent 1** scans the file structure to identify the language (e.g., "Python") and define the analysis goal (e.g., "Check for security vulnerabilities in file processing").
3.  **Scan** for relevant files (e.g., `.py`) based on Agent 1's output.
4.  **Chunk** the code for processing.
5.  **Agent 2** analyzes chunks in parallel using the tailored prompt from Agent 1.
6.  **Generate** a `report.md` file in the current directory containing the findings.

## Technical Details

- **Concurrency**: The tool uses a worker pool (default 10 workers) to send requests to Gemini concurrently.
- **Context Awareness**: Agent 1 builds a "Project Context" object containing the language, target file extensions, and a specific "persona" for the reviewer (Agent 2) to adopt.
- **Cache**: Repositories are cloned to `~/.bugscan/repos` to avoid re-downloading if they have already been analyzed (doing a `git pull` instead).