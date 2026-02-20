// Package main runs a simple CLI agent that uses the Anthropic API with tools
// (e.g. read file). It reads user input from stdin and streams agent replies to stdout.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"agentExample/tools"

	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()
	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	agentTools := []tools.ToolDefinition{
		tools.ReadFileDefinition, tools.ListFilesDefinition, tools.EditFileDefinition,
		tools.CreateFileDefinition, tools.RemoveFileDefinition, tools.SearchFileDefinition,
		tools.GrepInFileDefinition, tools.GrepInFilesDefinition, tools.RunCommandDefinition,
		tools.GetWorkingDirDefinition, tools.MoveFileDefinition, tools.CopyFileDefinition,
		tools.FileInfoDefinition, tools.ListFilesRecursiveDefinition, tools.ReadFileLinesDefinition,
		tools.CreateDirectoryDefinition, tools.RemoveDirectoryDefinition,
	}
	agent := NewAgent(&client, getUserMessage, agentTools)
	err := agent.Run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

// Agent holds the API client, user input source, and available tools for a chat run.
type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []tools.ToolDefinition
}

// NewAgent builds an Agent with the given client, message reader, and tool set.
func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), agentTools []tools.ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          agentTools,
	}
}

const maxToolRounds = 10

// Run runs the interactive loop: read user message, call the model (with tool use),
// print the final text reply, repeat until stdin is closed.
func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}
	fmt.Println("Chat with the agent. Type 'ctrl+c' to exit.")

	for {
		fmt.Print("\033[94mYou\033[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}
		userInput = strings.TrimSpace(userInput)
		if userInput == "" {
			fmt.Fprintln(os.Stderr, "Please enter a message.")
			continue
		}

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		conversation = append(conversation, userMessage)
		message, err := a.runInterface(ctx, &conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())
		for _, block := range message.Content {
			switch block.Type {
			case "text":
				fmt.Printf("\033[93mAgent\033[0m: %s\n", block.Text)
			}
		}
	}
	return nil
}

// runInterface sends the conversation to the API and handles tool-use rounds
// until the model returns a non-tool response or maxToolRounds is reached.
func (a *Agent) runInterface(ctx context.Context, conversation *[]anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := make([]anthropic.ToolUnionParam, 0, len(a.tools))
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages:  *conversation,
		Tools:     anthropicTools,
	})
	if err != nil {
		return nil, err
	}

	for round := 0; round < maxToolRounds && message.StopReason == anthropic.StopReasonToolUse; round++ {
		*conversation = append(*conversation, message.ToParam())

		var toolResultBlocks []anthropic.ContentBlockParamUnion
		for _, block := range message.Content {
			toolUse, ok := block.AsAny().(anthropic.ToolUseBlock)
			if !ok {
				continue
			}
			// Print green "tool: name(input)" line for each tool activation
			fmt.Printf("\033[32mtool: %s(%s)\033[0m\n", toolUse.Name, string(toolUse.Input))
			var result string
			var isError bool
			if fn := a.findTool(toolUse.Name); fn != nil {
				result, err = fn.Function(toolUse.Input)
				if err != nil {
					result = err.Error()
					isError = true
				}
			} else {
				result = fmt.Sprintf("unknown tool: %s", toolUse.Name)
				isError = true
			}
			toolResultBlocks = append(toolResultBlocks, anthropic.NewToolResultBlock(toolUse.ID, result, isError))
		}
		if len(toolResultBlocks) == 0 {
			break
		}

		toolResultMessage := anthropic.NewUserMessage(toolResultBlocks...)
		*conversation = append(*conversation, toolResultMessage)

		message, err = a.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     anthropic.ModelClaudeSonnet4_6,
			MaxTokens: 1024,
			Messages:  *conversation,
			Tools:     anthropicTools,
		})
		if err != nil {
			return nil, err
		}
	}

	return message, nil
}

func (a *Agent) findTool(name string) *tools.ToolDefinition {
	for i := range a.tools {
		if a.tools[i].Name == name {
			return &a.tools[i]
		}
	}
	return nil
}
