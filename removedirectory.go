// Package main provides the removeDirectory tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// RemoveDirectoryDefinition is the tool that removes a directory.
var RemoveDirectoryDefinition = ToolDefinition{
	Name:        "removeDirectory",
	Description: "Remove a directory. If recursive is true, remove its contents too; otherwise the directory must be empty.",
	InputSchema: RemoveDirectoryInputSchema,
	Function:    RemoveDirectory,
}

// RemoveDirectoryInput is the JSON shape for the removeDirectory tool.
type RemoveDirectoryInput struct {
	Path      string `json:"path" jsonschema_description:"The path of the directory to remove."`
	Recursive bool   `json:"recursive" jsonschema_description:"If true, remove directory and all contents; if false, directory must be empty."`
}

// RemoveDirectoryInputSchema is the Anthropic tool input schema for removeDirectory.
var RemoveDirectoryInputSchema = GenerateSchema[RemoveDirectoryInput]()

// RemoveDirectory implements the removeDirectory tool: Remove or RemoveAll based on recursive.
func RemoveDirectory(input json.RawMessage) (string, error) {
	var removeDirectoryInput RemoveDirectoryInput
	if err := json.Unmarshal(input, &removeDirectoryInput); err != nil {
		return "", fmt.Errorf("removeDirectory input: %w", err)
	}
	path := filepath.Clean(removeDirectoryInput.Path)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("removeDirectory: path not found: %s", path)
		}
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("removeDirectory: path is not a directory: %s", path)
	}
	if removeDirectoryInput.Recursive {
		if err := os.RemoveAll(path); err != nil {
			return "", err
		}
		return fmt.Sprintf("Removed directory and contents: %s", path), nil
	}
	if err := os.Remove(path); err != nil {
		return "", fmt.Errorf("removeDirectory: %w (directory may not be empty)", err)
	}
	return fmt.Sprintf("Removed directory %s", path), nil
}
