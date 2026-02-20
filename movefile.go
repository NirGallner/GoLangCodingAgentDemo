// Package main provides the moveFile tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MoveFileDefinition is the tool that moves or renames a file.
var MoveFileDefinition = ToolDefinition{
	Name:        "moveFile",
	Description: "Move or rename a file to a new path. Use when refactoring or reorganizing files. Overwrites destination if it exists and is a file.",
	InputSchema: MoveFileInputSchema,
	Function:    MoveFile,
}

// MoveFileInput is the JSON shape for the moveFile tool.
type MoveFileInput struct {
	FromPath string `json:"fromPath" jsonschema_description:"The current path of the file."`
	ToPath   string `json:"toPath" jsonschema_description:"The destination path."`
}

// MoveFileInputSchema is the Anthropic tool input schema for moveFile.
var MoveFileInputSchema = GenerateSchema[MoveFileInput]()

// MoveFile implements the moveFile tool: renames/moves the file; copies then removes if cross-filesystem.
func MoveFile(input json.RawMessage) (string, error) {
	var moveFileInput MoveFileInput
	if err := json.Unmarshal(input, &moveFileInput); err != nil {
		return "", fmt.Errorf("moveFile input: %w", err)
	}
	fromPath := filepath.Clean(moveFileInput.FromPath)
	toPath := filepath.Clean(moveFileInput.ToPath)
	info, err := os.Stat(fromPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("moveFile: source not found: %s", fromPath)
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("moveFile: source is a directory, not a file: %s", fromPath)
	}
	err = os.Rename(fromPath, toPath)
	if err == nil {
		return fmt.Sprintf("Moved %s to %s", fromPath, toPath), nil
	}
	// Cross-filesystem: copy then remove
	content, err := os.ReadFile(fromPath)
	if err != nil {
		return "", fmt.Errorf("moveFile: read: %w", err)
	}
	toDir := filepath.Dir(toPath)
	if toDir != "." {
		if err := os.MkdirAll(toDir, 0755); err != nil {
			return "", fmt.Errorf("moveFile: mkdir %s: %w", toDir, err)
		}
	}
	if err := os.WriteFile(toPath, content, info.Mode().Perm()); err != nil {
		return "", fmt.Errorf("moveFile: write: %w", err)
	}
	if err := os.Remove(fromPath); err != nil {
		return "", fmt.Errorf("moveFile: remove source after copy: %w", err)
	}
	return fmt.Sprintf("Moved %s to %s (cross-filesystem)", fromPath, toPath), nil
}
