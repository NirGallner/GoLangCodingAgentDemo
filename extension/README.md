# agentExample Chat

VS Code extension that opens a chat panel powered by the agentExample Go agent. The agent can read and edit files in your workspace (including `main.go`).

## Setup

1. **API key**: Set `agentExample.apiKey` in settings, or set the `ANTHROPIC_API_KEY` environment variable.
2. **Binary**: Either run `npm run build:agent` from this folder to build the Go agent into `extension/bin/`, or set `agentExample.agentPath` to the path of your `agentExample` binary.

## Usage

- **Open agentExample Chat**: Run the command from the Command Palette (Ctrl/Cmd+Shift+P), or open the "agentExample Chat" view in the Explorer sidebar.
- **main.go**: Click the "main.go" button to insert the contents of `main.go` into the input so the agent has that context.

## Development

- `npm install` — install dependencies
- `npm run compile` — compile TypeScript
- `npm run build:agent` — build the Go agent into `bin/` (from repo root)
- Run the extension via VS Code: open this folder and press F5 (Run > Start Debugging).
