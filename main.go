package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

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

	agent := NewAgent(&client, getUserMessage)
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool)) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with the agent. Type 'ctrl+c' to exit.")

	for {
		fmt.Print("\033[94mYou\033[0m: ")
		userInput, ok := a.getUserMessage()
		if !ok {
			break
		}

		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
		conversation = append(conversation, userMessage)
		message, err := a.runInterface(ctx, conversation)
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

func (a *Agent) runInterface(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		Messages:  conversation,
	})
	if err != nil {
		return nil, err
	}
	return message, nil
}
