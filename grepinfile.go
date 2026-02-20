// Package main provides the grepInFile tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GrepInFileDefinition is the tool that searches for a pattern inside a single file.
var GrepInFileDefinition = ToolDefinition{
	Name:        "grepInFile",
	Description: "Search for a string pattern inside a single file; return matching lines with line numbers. Use when you need to find where something appears in a file.",
	InputSchema: GrepInFileInputSchema,
	Function:    GrepInFile,
}

// GrepInFileInput is the JSON shape for the grepInFile tool.
type GrepInFileInput struct {
	Path       string `json:"path" jsonschema_description:"The relative path of the file to search."`
	Pattern    string `json:"pattern" jsonschema_description:"The string to search for (substring match)."`
	MaxMatches int    `json:"maxMatches" jsonschema_description:"Optional cap on number of matches returned; 0 or omit means 50."`
}

// GrepInFileInputSchema is the Anthropic tool input schema for grepInFile.
var GrepInFileInputSchema = GenerateSchema[GrepInFileInput]()

const defaultGrepInFileMax = 50

// GrepInFile implements the grepInFile tool: reads file, returns lines containing pattern with line numbers.
func GrepInFile(input json.RawMessage) (string, error) {
	var grepInFileInput GrepInFileInput
	if err := json.Unmarshal(input, &grepInFileInput); err != nil {
		return "", fmt.Errorf("grepInFile input: %w", err)
	}
	path := grepInFileInput.Path
	pattern := grepInFileInput.Pattern
	maxMatches := grepInFileInput.MaxMatches
	if maxMatches <= 0 {
		maxMatches = defaultGrepInFileMax
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(content), "\n")
	var matches []string
	for i, line := range lines {
		if len(matches) >= maxMatches {
			break
		}
		if strings.Contains(line, pattern) {
			matches = append(matches, fmt.Sprintf("%d: %s", i+1, line))
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("No matches for %q in %s", pattern, path), nil
	}
	return strings.Join(matches, "\n"), nil
}
