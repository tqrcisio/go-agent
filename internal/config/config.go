package config

import (
	"encoding/json"
	"os"
)

// Settings defines the global application configuration.
type Settings struct {
	ShellWhitelist []string `json:"shell_whitelist"`
}

// DefaultSettings returns the default configuration.
func DefaultSettings() Settings {
	return Settings{
		ShellWhitelist: []string{
			"ls", "cat", "grep", "find", 
			"git", "go", "echo", "pwd", "mkdir",
		},
	}
}

// LoadSettings loads settings from settings.json or creates it with defaults.
func LoadSettings() (*Settings, error) {
	configPath := "settings.json" // Root directory
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default
		settings := DefaultSettings()
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return nil, err
		}
		return &settings, nil
	}

	// Read existing
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}
