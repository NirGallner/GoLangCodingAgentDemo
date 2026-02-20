// Package tools provides the fileInfo tool for the agent.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileInfoDefinition is the tool that returns metadata for a path.
var FileInfoDefinition = ToolDefinition{
	Name:        "fileInfo",
	Description: "Return metadata for a path: size, modification time, whether it is a directory, and permissions. Use this to check if a path exists, is a file or directory, or how large it is before reading.",
	InputSchema: FileInfoInputSchema,
	Function:    FileInfo,
}

// FileInfoInput is the JSON shape for the fileInfo tool.
type FileInfoInput struct {
	Path string `json:"path" jsonschema_description:"The relative path to stat."`
}

// FileInfoInputSchema is the Anthropic tool input schema for fileInfo.
var FileInfoInputSchema = GenerateSchema[FileInfoInput]()

// FileInfo implements the fileInfo tool: returns size, mod time, isDir, and mode.
func FileInfo(input json.RawMessage) (string, error) {
	var fileInfoInput FileInfoInput
	if err := json.Unmarshal(input, &fileInfoInput); err != nil {
		return "", fmt.Errorf("fileInfo input: %w", err)
	}
	info, err := os.Stat(fileInfoInput.Path)
	if err != nil {
		return "", err
	}
	mode := info.Mode()
	return fmt.Sprintf("path: %s\nsize: %d\nmodTime: %s\nisDir: %t\nmode: %s", fileInfoInput.Path, info.Size(), info.ModTime().Format("2006-01-02 15:04:05"), info.IsDir(), mode.String()), nil
}
