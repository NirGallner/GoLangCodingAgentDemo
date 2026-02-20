package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// Keep imports used until they are referenced in code.
var (
	_ = bufio.Reader{}
	_ = os.Stdin
	_ = context.Background
	_ = (*anthropic.Client)(nil)
)

func main() {
	fmt.Println("Hello from agentExample")
}
