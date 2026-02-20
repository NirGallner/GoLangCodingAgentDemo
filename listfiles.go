// Package main provides the listFiles tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListFilesDefinition is the tool that lists files in a given directory.
var ListFilesDefinition = ToolDefinition{
	Name:        "listFiles",
	Description: "List all files and directories at the given path. Use this when you want to see what files exist in a directory. Pass a directory path (relative to the working directory).",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

// ListFilesInput is the JSON shape for the listFiles tool.
type ListFilesInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a directory in the working directory."`
}

// ListFilesInputSchema is the Anthropic tool input schema for listFiles.
var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

// ListFiles implements the listFiles tool: lists entries in the given directory and returns their names or an error.
func ListFiles(input json.RawMessage) (string, error) {
	var listFilesInput ListFilesInput
	if err := json.Unmarshal(input, &listFilesInput); err != nil {
		return "", fmt.Errorf("listFiles input: %w", err)
	}
	path := filepath.Clean(listFilesInput.Path)
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name = name + "/"
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, "\n"), nil
}
