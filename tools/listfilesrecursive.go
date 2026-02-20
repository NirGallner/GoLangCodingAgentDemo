// Package tools provides the listFilesRecursive tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ListFilesRecursiveDefinition is the tool that lists files under a directory recursively.
var ListFilesRecursiveDefinition = ToolDefinition{
	Name:        "listFilesRecursive",
	Description: "List all files and directories under a directory recursively, optionally limited by max depth. Use to see the full tree. Entries use trailing / for directories.",
	InputSchema: ListFilesRecursiveInputSchema,
	Function:    ListFilesRecursive,
}

// ListFilesRecursiveInput is the JSON shape for the listFilesRecursive tool.
type ListFilesRecursiveInput struct {
	RootPath string `json:"rootPath" jsonschema_description:"Directory to list; default is current directory (.)."`
	MaxDepth int    `json:"maxDepth" jsonschema_description:"Optional maximum depth (0 or omit = unlimited). Depth 1 is immediate children only."`
}

// ListFilesRecursiveInputSchema is the Anthropic tool input schema for listFilesRecursive.
var ListFilesRecursiveInputSchema = GenerateSchema[ListFilesRecursiveInput]()

// ListFilesRecursive implements the listFilesRecursive tool: WalkDir and collect paths; apply maxDepth if set.
func ListFilesRecursive(input json.RawMessage) (string, error) {
	var listFilesRecursiveInput ListFilesRecursiveInput
	if err := json.Unmarshal(input, &listFilesRecursiveInput); err != nil {
		return "", fmt.Errorf("listFilesRecursive input: %w", err)
	}
	rootPath := listFilesRecursiveInput.RootPath
	if rootPath == "" {
		rootPath = "."
	}
	rootPath = filepath.Clean(rootPath)
	maxDepth := listFilesRecursiveInput.MaxDepth
	info, err := os.Stat(rootPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("listFilesRecursive: rootPath must be a directory: %s", rootPath)
	}
	var entries []string
	err = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil
		}
		if rel == "." {
			return nil
		}
		if maxDepth > 0 {
			depth := strings.Count(rel, string(os.PathSeparator))
			if d.IsDir() {
				if depth+1 > maxDepth {
					return filepath.SkipDir
				}
			} else if depth >= maxDepth {
				return nil
			}
		}
		if d.IsDir() {
			entries = append(entries, path+"/")
		} else {
			entries = append(entries, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(entries)
	return strings.Join(entries, "\n"), nil
}
