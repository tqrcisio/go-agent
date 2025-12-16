package report

import (
	"fmt"
	"go-agent/internal/geminiclient"
	"os"
	"sort"
	"strings"
	"time"
)

// GenerateMarkdown creates a markdown report from the findings.
func GenerateMarkdown(findings []geminiclient.Finding, repoURL string) (string, error) {
	if len(findings) == 0 {
		return "No findings to report.", nil
	}

	// Group findings by severity
	findingsBySeverity := make(map[string][]geminiclient.Finding)
	for _, f := range findings {
		severityUpper := strings.ToUpper(f.Severity) // Convert to uppercase
		findingsBySeverity[severityUpper] = append(findingsBySeverity[severityUpper], f)
	}

	var md strings.Builder

	// Header
	md.WriteString(fmt.Sprintf("# BugScan Analysis Report for %s\n\n", repoURL))
	md.WriteString(fmt.Sprintf("**Generated on:** %s\n\n", time.Now().Format(time.RFC1123)))
	md.WriteString("---\n\n")

	// Write sections for each severity, ordered
	severities := []string{"HIGH", "MEDIUM", "LOW"}
	for _, severity := range severities {
		if issues, found := findingsBySeverity[severity]; found {
			// Sort issues by file and line for consistent output
			sort.Slice(issues, func(i, j int) bool {
				if issues[i].File != issues[j].File {
					return issues[i].File < issues[j].File
				}
				return issues[i].Line < issues[j].Line
			})

			md.WriteString(getSeverityHeader(severity))
			for _, issue := range issues {
				md.WriteString(fmt.Sprintf("- `%s:%d` – %s\n", issue.File, issue.Line, issue.Description))
			}
			md.WriteString("\n")
		}
	}

	reportContent := md.String()
	reportFileName := "report.md"
	err := os.WriteFile(reportFileName, []byte(reportContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write report to %s: %w", reportFileName, err)
	}

	return fmt.Sprintf("Report successfully generated: %s", reportFileName), nil
}

func getSeverityHeader(severity string) string {
	switch severity {
	case "HIGH":
		return "## 🔴 High Severity\n"
	case "MEDIUM":
		return "## 🟡 Medium Severity\n"
	case "LOW":
		return "## 🔵 Low Severity\n"
	default:
		return fmt.Sprintf("## %s Severity\n", strings.Title(strings.ToLower(severity)))
	}
}
