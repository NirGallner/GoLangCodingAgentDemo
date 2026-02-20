// Package main provides the grepInFiles tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GrepInFilesDefinition is the tool that searches for a pattern in files under a directory.
var GrepInFilesDefinition = ToolDefinition{
	Name:        "grepInFiles",
	Description: "Search for a string pattern in files under a directory; return file path and matching lines. Optionally filter by glob (e.g. *.go).",
	InputSchema: GrepInFilesInputSchema,
	Function:    GrepInFiles,
}

// GrepInFilesInput is the JSON shape for the grepInFiles tool.
type GrepInFilesInput struct {
	RootPath   string `json:"rootPath" jsonschema_description:"Directory to search in; default is current directory (.)."`
	Pattern    string `json:"pattern" jsonschema_description:"The string to search for (substring match)."`
	Glob       string `json:"glob" jsonschema_description:"Optional glob to filter files (e.g. *.go); empty means all files."`
	MaxResults int    `json:"maxResults" jsonschema_description:"Optional cap on total match count; 0 or omit means 100."`
}

// GrepInFilesInputSchema is the Anthropic tool input schema for grepInFiles.
var GrepInFilesInputSchema = GenerateSchema[GrepInFilesInput]()

const defaultGrepInFilesMax = 100

// GrepInFiles implements the grepInFiles tool: walks directory, searches each file, returns path:lineNum: line.
func GrepInFiles(input json.RawMessage) (string, error) {
	var grepInFilesInput GrepInFilesInput
	if err := json.Unmarshal(input, &grepInFilesInput); err != nil {
		return "", fmt.Errorf("grepInFiles input: %w", err)
	}
	rootPath := grepInFilesInput.RootPath
	if rootPath == "" {
		rootPath = "."
	}
	rootPath = filepath.Clean(rootPath)
	pattern := grepInFilesInput.Pattern
	glob := strings.TrimSpace(grepInFilesInput.Glob)
	maxResults := grepInFilesInput.MaxResults
	if maxResults <= 0 {
		maxResults = defaultGrepInFilesMax
	}
	info, err := os.Stat(rootPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("grepInFiles: rootPath must be a directory: %s", rootPath)
	}
	var results []string
	err = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if glob != "" {
			matched, err := filepath.Match(glob, filepath.Base(path))
			if err != nil || !matched {
				return nil
			}
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if len(results) >= maxResults {
				return filepath.SkipAll
			}
			if strings.Contains(line, pattern) {
				results = append(results, fmt.Sprintf("%s:%d: %s", path, i+1, line))
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return fmt.Sprintf("No matches for %q under %s", pattern, rootPath), nil
	}
	return strings.Join(results, "\n"), nil
}
