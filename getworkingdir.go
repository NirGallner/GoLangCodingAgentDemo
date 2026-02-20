// Package main provides the getWorkingDir tool for the agent.
package main

import (
	"encoding/json"
	"os"
)

// GetWorkingDirDefinition is the tool that returns the current working directory.
var GetWorkingDirDefinition = ToolDefinition{
	Name:        "getWorkingDir",
	Description: "Return the current working directory path. Use this when you need to know where the process is running or to reason about relative paths.",
	InputSchema: GetWorkingDirInputSchema,
	Function:    GetWorkingDir,
}

// GetWorkingDirInput is the JSON shape for the getWorkingDir tool (no required fields).
type GetWorkingDirInput struct{}

// GetWorkingDirInputSchema is the Anthropic tool input schema for getWorkingDir.
var GetWorkingDirInputSchema = GenerateSchema[GetWorkingDirInput]()

// GetWorkingDir implements the getWorkingDir tool: returns the current working directory.
func GetWorkingDir(input json.RawMessage) (string, error) {
	var getWorkingDirInput GetWorkingDirInput
	if err := json.Unmarshal(input, &getWorkingDirInput); err != nil {
		return "", err
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd, nil
}
