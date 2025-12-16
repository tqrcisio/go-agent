# BugScan (Go Agent)

BugScan is a powerful command-line tool written in Go that leverages Google's Gemini AI to analyze GitHub repositories for potential bugs, vulnerabilities, and code quality issues. It automates the process of cloning a repository, scanning for Go files, chunking the code, and submitting it to the Gemini 2.5 Pro model for deep analysis.

## Features

- **Automated Repository Cloning**: seamless cloning of GitHub repositories to a temporary cache.
- **Smart Code Scanning**: recursively scans for `.go` files while ignoring `.git` and `vendor` directories.
- **Intelligent Chunking**: splits large files into manageable chunks to ensure optimal processing by the AI model.
- **AI-Powered Analysis**: utilizes the **Gemini 2.5 Pro** model to detect:
  - Potential bugs
  - Race conditions
  - Nil pointer risks
  - Incorrect error handling
- **Concurrent Processing**: employs a worker pool pattern to analyze code chunks in parallel for high performance.
- **Comprehensive Reporting**: generates a detailed Markdown report (`report.md`) categorized by severity (High, Medium, Low).

## Project Structure

```
.
├── cmd/
│   └── go-agent/
│       └── main.go           # Application entry point (CLI definition)
├── internal/
│   ├── chunker/
│   │   └── chunker.go        # Logic for splitting files into code chunks
│   ├── geminiclient/
│   │   └── geminiclient.go   # Client for interacting with Google's Gemini API
│   ├── repomanager/
│   │   └── repomanager.go    # Handles git cloning and updates
│   ├── report/
│   │   └── report.go         # Generates Markdown reports from findings
│   ├── scanner/
│   │   └── scanner.go        # Scans directories for Go files
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

### Example

```bash
go run ./cmd/go-agent analyze https://github.com/gin-gonic/gin
```

### Output

The tool will:
1.  Clone the specified repository.
2.  Scan and chunk the code.
3.  Analyze chunks in parallel using Gemini.
4.  Print progress to the console.
5.  Generate a `report.md` file in the current directory containing the analysis results.

## Technical Details

- **Concurrency**: The tool uses a worker pool (default 10 workers) to send requests to Gemini concurrently, significantly speeding up the analysis of large codebases.
- **Context Awareness**: While currently analyzing file chunks individually, the prompt is engineered to ask for specific Go-related issues.
- **Cache**: Repositories are cloned to `~/.bugscan/repos` to avoid re-downloading if they have already been analyzed (doing a `git pull` instead).
