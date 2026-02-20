# agentExample

A CLI agent that uses the Anthropic API with tools (read/edit files, run commands, grep, etc.). Includes a VS Code extension for chatting with the agent inside the editor.

## Requirements

- Go 1.21+
- [Anthropic API key](https://console.anthropic.com/) (set `ANTHROPIC_API_KEY` or use extension settings)

## CLI

Build and run the agent from stdin/stdout:

```bash
go build -o agentExample .
./agentExample
```

Type messages and press Enter. Use `/clear` or `/reset` to clear conversation context.

## VS Code extension

The `extension/` folder contains a VS Code extension that opens a chat panel powered by the same agent.

- **Setup**: See [extension/README.md](extension/README.md) for API key and binary setup.
- **Develop**: From `extension/`: `npm install`, `npm run compile`, then run via F5 in VS Code.

## Project layout

- `main.go` — CLI entrypoint and agent loop
- `tools/` — Tool definitions (file ops, grep, run command, etc.)
- `extension/` — VS Code extension (TypeScript) for the chat UI

## License

Use as you like.
