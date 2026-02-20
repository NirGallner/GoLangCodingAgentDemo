// Package tools provides the createDirectory tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateDirectoryDefinition is the tool that creates a directory and parent directories.
var CreateDirectoryDefinition = ToolDefinition{
	Name:        "createDirectory",
	Description: "Create a directory at the given path; create parent directories if needed (like mkdir -p). Use when you need to ensure a directory exists.",
	InputSchema: CreateDirectoryInputSchema,
	Function:    CreateDirectory,
}

// CreateDirectoryInput is the JSON shape for the createDirectory tool.
type CreateDirectoryInput struct {
	Path string `json:"path" jsonschema_description:"The path of the directory to create."`
}

// CreateDirectoryInputSchema is the Anthropic tool input schema for createDirectory.
var CreateDirectoryInputSchema = GenerateSchema[CreateDirectoryInput]()

// CreateDirectory implements the createDirectory tool: MkdirAll(path, 0755).
func CreateDirectory(input json.RawMessage) (string, error) {
	var createDirectoryInput CreateDirectoryInput
	if err := json.Unmarshal(input, &createDirectoryInput); err != nil {
		return "", fmt.Errorf("createDirectory input: %w", err)
	}
	path := filepath.Clean(createDirectoryInput.Path)
	if path == "" || path == "." {
		return "", fmt.Errorf("createDirectory: path is required")
	}
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("createDirectory: path exists and is not a directory: %s", path)
		}
		return fmt.Sprintf("Directory already exists: %s", path), nil
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("createDirectory: %w", err)
	}
	return fmt.Sprintf("Created directory %s", path), nil
}
