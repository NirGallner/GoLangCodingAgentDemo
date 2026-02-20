// Package tools provides the remove_file tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

// RemoveFileDefinition is the tool that removes a file.
var RemoveFileDefinition = ToolDefinition{
	Name:        "remove_file",
	Description: "Remove (delete) a file at the given path. Use this when the user wants to delete a file. Pass the relative path of the file. Does not remove directories.",
	InputSchema: RemoveFileInputSchema,
	Function:    RemoveFile,
}

// RemoveFileInput is the JSON shape for the remove_file tool.
type RemoveFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of the file to remove."`
}

// RemoveFileInputSchema is the Anthropic tool input schema for remove_file.
var RemoveFileInputSchema = GenerateSchema[RemoveFileInput]()

// RemoveFile implements the remove_file tool: deletes the file at the given path.
func RemoveFile(input json.RawMessage) (string, error) {
	var removeFileInput RemoveFileInput
	if err := json.Unmarshal(input, &removeFileInput); err != nil {
		return "", fmt.Errorf("remove_file input: %w", err)
	}
	path := removeFileInput.Path
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("remove_file: file not found: %s", path)
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("remove_file: path is a directory, not a file: %s", path)
	}
	if err := os.Remove(path); err != nil {
		return "", err
	}
	return fmt.Sprintf("Removed file %s", path), nil
}
