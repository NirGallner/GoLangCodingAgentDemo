// Package tools provides the clear_context tool for the agent.
package tools

import (
	"encoding/json"
)

// MakeClearContextDefinition returns a clear_context tool that invokes onClear when called.
// The callback is typically used to reset conversation history (e.g. in the main run loop).
func MakeClearContextDefinition(onClear func()) ToolDefinition {
	return ToolDefinition{
		Name:        "clear_context",
		Description: "Clear the conversation history so the next user message starts a fresh context. Use when the user asks to start over, forget the past, or clear the chat.",
		InputSchema: ClearContextInputSchema,
		Function: func(input json.RawMessage) (string, error) {
			var clearContextInput ClearContextInput
			if err := json.Unmarshal(input, &clearContextInput); err != nil {
				return "", err
			}
			onClear()
			return "Context cleared.", nil
		},
	}
}

// ClearContextInput is the JSON shape for the clear_context tool (no required fields).
type ClearContextInput struct{}

// ClearContextInputSchema is the Anthropic tool input schema for clear_context.
var ClearContextInputSchema = GenerateSchema[ClearContextInput]()
