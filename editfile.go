// Package main provides the editFile tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// EditFileDefinition is the tool that edits an existing file by replacing strings.
var EditFileDefinition = ToolDefinition{
	Name:        "edit_file",
	Description: "Edit an existing file by replacing one string with another. Use this when you need to change specific text within a file. Pass the file path (relative to the working directory), the exact string to find (old_string), and the string to replace it with (new_string). All occurrences of old_string in the file are replaced. Returns the number of replacements made or an error.",
	InputSchema: EditFileInputSchema,
	Function:    EditFile,
}

// EditFileInput is the JSON shape for the edit_file tool.
type EditFileInput struct {
	Path      string `json:"path" jsonschema_description:"The relative path of the file to edit."`
	OldString string `json:"old_string" jsonschema_description:"The exact string to find and replace in the file."`
	NewString string `json:"new_string" jsonschema_description:"The string to replace old_string with."`
}

// EditFileInputSchema is the Anthropic tool input schema for edit_file.
var EditFileInputSchema = GenerateSchema[EditFileInput]()

// EditFile implements the edit_file tool: reads the file, replaces all occurrences of old_string with new_string, writes back.
func EditFile(input json.RawMessage) (string, error) {
	var editFileInput EditFileInput
	if err := json.Unmarshal(input, &editFileInput); err != nil {
		return "", fmt.Errorf("edit_file input: %w", err)
	}
	path := editFileInput.Path
	oldStr := editFileInput.OldString
	newStr := editFileInput.NewString

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	s := string(content)
	if oldStr == "" {
		return "", fmt.Errorf("edit_file: old_string must not be empty")
	}
	newContent := strings.ReplaceAll(s, oldStr, newStr)
	count := strings.Count(s, oldStr)
	if count == 0 {
		return "", fmt.Errorf("edit_file: old_string not found in file")
	}
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("Replaced %d occurrence(s) of the given string in %s", count, path), nil
}
