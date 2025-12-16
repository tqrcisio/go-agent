# BugScan Agents

This directory contains the individual agents that form the multi-agent architecture of BugScan. Each agent is responsible for a specific task within the code analysis pipeline, working sequentially and communicating via a shared `state.State` object.

## Agents Overview

- **RepoCloneAgent**: Handles the cloning or updating of the target GitHub repository into a local cache.
- **FileStructureAgent**: Scans the cloned repository and constructs a string representation of its file and directory structure.
- **IdentifyProjectAgent**: Utilizes Google's Gemini AI (Agent 1 - The Architect) to analyze the file structure, identify the primary programming language, relevant file extensions, and formulate a specific analysis goal/persona for subsequent agents.
- **ScanFilesAgent**: Based on the `ProjectInfo` from the `IdentifyProjectAgent`, this agent dynamically scans the repository to locate and collect all relevant source code files.
- **ChunkingAgent**: Takes the identified source files and splits their content into smaller, manageable `CodeChunk`s. This ensures that large files can be processed effectively by the AI model without exceeding context window limits.
- **AnalyzeChunksAgent**: Employs Google's Gemini AI (Agent 2 - The Reviewer) to perform deep code analysis on each `CodeChunk`. It works concurrently using a worker pool, identifying potential bugs, security vulnerabilities, and code quality issues.
- **DeduplicateFindingsAgent**: Processes the raw findings from the `AnalyzeChunksAgent` to remove any duplicate issues, ensuring a clean and concise report.
- **PrioritizeFindingsAgent**: Sorts the unique findings based on their severity (High, Medium, Low) and other criteria, presenting the most critical issues first.
- **ReportAgent**: Generates the final, comprehensive `report.md` file in Markdown format, summarizing all findings categorized by severity, file, and line number.
