// Package main provides the runCommand tool for the agent.
package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// RunCommandDefinition is the tool that runs a shell command and returns output and exit code.
var RunCommandDefinition = ToolDefinition{
	Name:        "runCommand",
	Description: "Run a shell command and return stdout, stderr, and exit code. Use for builds, tests, linters, or any shell command. Working directory is optional.",
	InputSchema: RunCommandInputSchema,
	Function:    RunCommand,
}

// RunCommandInput is the JSON shape for the runCommand tool.
type RunCommandInput struct {
	Command    string `json:"command" jsonschema_description:"The shell command to run (e.g. go build, go test, ./go-lint)."`
	WorkingDir string `json:"workingDir" jsonschema_description:"Optional working directory for the command; default is current directory."`
}

// RunCommandInputSchema is the Anthropic tool input schema for runCommand.
var RunCommandInputSchema = GenerateSchema[RunCommandInput]()

// RunCommand implements the runCommand tool: runs the command via sh -c and returns exit code, stdout, stderr.
func RunCommand(input json.RawMessage) (string, error) {
	var runCommandInput RunCommandInput
	if err := json.Unmarshal(input, &runCommandInput); err != nil {
		return "", fmt.Errorf("runCommand input: %w", err)
	}
	command := strings.TrimSpace(runCommandInput.Command)
	if command == "" {
		return "", fmt.Errorf("runCommand: command is required")
	}
	cmd := exec.Command("sh", "-c", command)
	if runCommandInput.WorkingDir != "" {
		cmd.Dir = runCommandInput.WorkingDir
	}
	stdout, err := cmd.Output()
	stdoutStr := string(stdout)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderrStr := string(exitErr.Stderr)
			return fmt.Sprintf("exitCode: %d\nstdout:\n%sstderr:\n%s", exitErr.ExitCode(), stdoutStr, stderrStr), nil
		}
		return "", err
	}
	return fmt.Sprintf("exitCode: 0\nstdout:\n%s", stdoutStr), nil
}
