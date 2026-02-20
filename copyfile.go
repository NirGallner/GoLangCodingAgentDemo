// Package main provides the copyFile tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CopyFileDefinition is the tool that copies a file to another path.
var CopyFileDefinition = ToolDefinition{
	Name:        "copyFile",
	Description: "Copy a file to another path. Overwrites destination if it exists. Use to duplicate a file or create a backup before editing.",
	InputSchema: CopyFileInputSchema,
	Function:    CopyFile,
}

// CopyFileInput is the JSON shape for the copyFile tool.
type CopyFileInput struct {
	FromPath string `json:"fromPath" jsonschema_description:"The path of the file to copy."`
	ToPath   string `json:"toPath" jsonschema_description:"The destination path."`
}

// CopyFileInputSchema is the Anthropic tool input schema for copyFile.
var CopyFileInputSchema = GenerateSchema[CopyFileInput]()

// CopyFile implements the copyFile tool: reads source, ensures parent dir of destination, writes with same mode.
func CopyFile(input json.RawMessage) (string, error) {
	var copyFileInput CopyFileInput
	if err := json.Unmarshal(input, &copyFileInput); err != nil {
		return "", fmt.Errorf("copyFile input: %w", err)
	}
	fromPath := filepath.Clean(copyFileInput.FromPath)
	toPath := filepath.Clean(copyFileInput.ToPath)
	content, err := os.ReadFile(fromPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("copyFile: source not found: %s", fromPath)
		}
		return "", err
	}
	info, err := os.Stat(fromPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("copyFile: source is a directory: %s", fromPath)
	}
	mode := info.Mode().Perm()
	toDir := filepath.Dir(toPath)
	if toDir != "." {
		if err := os.MkdirAll(toDir, 0755); err != nil {
			return "", fmt.Errorf("copyFile: mkdir %s: %w", toDir, err)
		}
	}
	if err := os.WriteFile(toPath, content, mode); err != nil {
		return "", err
	}
	return fmt.Sprintf("Copied %s to %s", fromPath, toPath), nil
}
