// Package main provides the readFileLines tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ReadFileLinesDefinition is the tool that reads a range of lines from a file.
var ReadFileLinesDefinition = ToolDefinition{
	Name:        "readFileLines",
	Description: "Read a range of lines from a file. Lines are 1-based: startLine 1 is the first line. Use for large files when you only need a portion.",
	InputSchema: ReadFileLinesInputSchema,
	Function:    ReadFileLines,
}

// ReadFileLinesInput is the JSON shape for the readFileLines tool.
type ReadFileLinesInput struct {
	Path      string `json:"path" jsonschema_description:"The relative path of the file."`
	StartLine int    `json:"startLine" jsonschema_description:"First line to include (1-based)."`
	EndLine   int    `json:"endLine" jsonschema_description:"Last line to include (1-based, inclusive)."`
}

// ReadFileLinesInputSchema is the Anthropic tool input schema for readFileLines.
var ReadFileLinesInputSchema = GenerateSchema[ReadFileLinesInput]()

// ReadFileLines implements the readFileLines tool: reads file, returns lines [startLine..endLine] (1-based).
func ReadFileLines(input json.RawMessage) (string, error) {
	var readFileLinesInput ReadFileLinesInput
	if err := json.Unmarshal(input, &readFileLinesInput); err != nil {
		return "", fmt.Errorf("readFileLines input: %w", err)
	}
	path := readFileLinesInput.Path
	start := readFileLinesInput.StartLine
	end := readFileLinesInput.EndLine
	if start < 1 || end < 1 {
		return "", fmt.Errorf("readFileLines: startLine and endLine must be >= 1")
	}
	if end < start {
		return "", fmt.Errorf("readFileLines: endLine must be >= startLine")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	if start > len(lines) {
		return "", fmt.Errorf("readFileLines: startLine %d is beyond file length %d", start, len(lines))
	}
	endIdx := end
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	slice := lines[start-1 : endIdx]
	return strings.Join(slice, "\n"), nil
}
