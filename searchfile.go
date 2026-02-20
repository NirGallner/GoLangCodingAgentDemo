// Package main provides the searchFile tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SearchFileDefinition is the tool that searches for a file by name and returns its path if found.
var SearchFileDefinition = ToolDefinition{
	Name:        "searchFile",
	Description: "Search for a file by name under a given directory. Returns the relative path(s) of any matching file(s), or a message if not found. Use this when you need to locate a file but only know its name.",
	InputSchema: SearchFileInputSchema,
	Function:    SearchFile,
}

// SearchFileInput is the JSON shape for the searchFile tool.
type SearchFileInput struct {
	FileName string `json:"fileName" jsonschema_description:"The name of the file to search for (e.g. main.go or README.md)."`
	RootPath string `json:"rootPath" jsonschema_description:"The directory to search in, relative to the working directory. Default is the current directory (.)."`
}

// SearchFileInputSchema is the Anthropic tool input schema for searchFile.
var SearchFileInputSchema = GenerateSchema[SearchFileInput]()

// SearchFile implements the searchFile tool: walks the directory tree from rootPath and returns paths of files whose name matches fileName.
func SearchFile(input json.RawMessage) (string, error) {
	var searchFileInput SearchFileInput
	if err := json.Unmarshal(input, &searchFileInput); err != nil {
		return "", fmt.Errorf("searchFile input: %w", err)
	}
	fileName := strings.TrimSpace(searchFileInput.FileName)
	if fileName == "" {
		return "", fmt.Errorf("fileName is required")
	}
	rootPath := searchFileInput.RootPath
	if rootPath == "" {
		rootPath = "."
	}
	rootPath = filepath.Clean(rootPath)

	info, err := os.Stat(rootPath)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("rootPath must be a directory: %s", rootPath)
	}

	var matches []string
	err = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == fileName {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No file named %q found under %s", fileName, rootPath), nil
	}
	return strings.Join(matches, "\n"), nil
}
