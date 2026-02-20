// Package main provides the createFile tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CreateFileDefinition is the tool that creates a new file with the given content.
var CreateFileDefinition = ToolDefinition{
	Name:        "create_file",
	Description: "Create a new file at the given path with the given content. Use this when the user wants to create a new file. Pass the relative path and the full file content. Creates parent directories if needed. If the file already exists, it is overwritten.",
	InputSchema: CreateFileInputSchema,
	Function:    CreateFile,
}

// CreateFileInput is the JSON shape for the create_file tool.
type CreateFileInput struct {
	Path    string `json:"path" jsonschema_description:"The relative path of the file to create."`
	Content string `json:"content" jsonschema_description:"The full content to write to the file."`
}

// CreateFileInputSchema is the Anthropic tool input schema for create_file.
var CreateFileInputSchema = GenerateSchema[CreateFileInput]()

// CreateFile implements the create_file tool: creates the file (and parent dirs if needed) with the given content.
func CreateFile(input json.RawMessage) (string, error) {
	var createFileInput CreateFileInput
	if err := json.Unmarshal(input, &createFileInput); err != nil {
		return "", fmt.Errorf("create_file input: %w", err)
	}
	path := filepath.Clean(createFileInput.Path)
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("create_file: mkdir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(createFileInput.Content), 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("Created file %s", path), nil
}
