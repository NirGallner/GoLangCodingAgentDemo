// Package tools provides the readFile tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadFileDefinition is the tool that reads file contents by relative path.
var ReadFileDefinition = ToolDefinition{
	Name:        "readFile",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

// ReadFileInput is the JSON shape for the readFile tool.
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

// ReadFileInputSchema is the Anthropic tool input schema for readFile.
var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

// ReadFile implements the readFile tool: reads the file at the given path and returns its contents or an error.
func ReadFile(input json.RawMessage) (string, error) {
	var readFileInput ReadFileInput
	if err := json.Unmarshal(input, &readFileInput); err != nil {
		return "", fmt.Errorf("readFile input: %w", err)
	}
	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
