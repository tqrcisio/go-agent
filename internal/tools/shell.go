package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ToolRunShell executes a shell command if it is whitelisted.
// Arguments: "command" (string).
func ToolRunShell(ctx context.Context, args map[string]interface{}, whitelist []string) (map[string]interface{}, error) {
	fullCommand, ok := args["command"].(string)
	if !ok || fullCommand == "" {
		return map[string]interface{}{ "error": "Missing 'command' argument" }, nil
	}

	// Parse command to get binary
	parts := strings.Fields(fullCommand)
	if len(parts) == 0 {
		return map[string]interface{}{ "error": "Empty command" }, nil
	}
	bin := parts[0]

	// Check whitelist
	allowed := false
	for _, w := range whitelist {
		if w == bin {
			allowed = true
			break
		}
	}

	if !allowed {
		return map[string]interface{}{
			"error": fmt.Sprintf("Command '%s' is not in the whitelist. Allowed: %v", bin, whitelist),
		},
		nil
	}

	// Execute
	// Using "bash -c" allows pipes and redirects, but is riskier. 
	// Given the requirement for "Tool Shell", usually models expect a real shell.
	// However, whitelist check is on the *full string* binary? No, we checked the first token.
	// If we use `bash -c`, the actual binary is `bash`.
	// If we execute directly `exec.Command(bin, args...)`, pipes won't work.
	// Let's stick to direct execution for safety first, unless the user explicitly wants shell features.
	// If the user allows "ls", running "ls | grep x" via direct execution won't work easily without "bash".
	
	// Compromise: We execute the binary directly with arguments.
	// This prevents "ls; rm -rf /" injection if we parsed correctly.
	cmdArgs := parts[1:]
	cmd := exec.CommandContext(ctx, bin, cmdArgs...)
	
	output, err := cmd.CombinedOutput() 
	
	// Handle exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1 // Unknown error
		}
	}

	result := string(output)
	// Truncate output
	if len(result) > 2000 {
		result = result[:2000] + "\n... (output truncated)"
	}

	return map[string]interface{}{
		"command":   fullCommand,
		"output":    result,
		"exit_code": exitCode,
		"error":     err != nil, // Boolean flag for convenience
	},
	nil
}
