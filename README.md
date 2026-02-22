# agentExample

A CLI agent that uses the Anthropic API with tools (read/edit files, run commands, grep, search the web, etc.). It follows the same pattern as [**How to Build an Agent**](https://ampcode.com/notes/how-to-build-an-agent) (Amp): an LLM, a loop, tool definitions, and execution of tool results back into the conversation. This project extends that idea with more tools and includes a VS Code extension for chatting with the agent inside the editor.

---

## Origin

This project is based on the article **[How to Build an Agent](https://ampcode.com/notes/how-to-build-an-agent)** (Thorsten Ball, Amp). The core loop is: send messages and tool definitions to the model; when the model requests a tool, execute it and append the result to the conversation; repeat until the model responds with text. All tools are implemented as name + description + JSON input schema + handler function.

---

## Requirements

- **Go 1.21+**
- **Anthropic API key** — [Create one](https://console.anthropic.com/), then set `ANTHROPIC_API_KEY` in your environment (or configure it in the VS Code extension settings).

---

## Tools

The agent has access to these tools (defined in `tools/`). The model chooses when to call them.

| Tool | Purpose |
|------|--------|
| `readFile` | Read contents of a file by relative path. |
| `readFileLines` | Read a range of lines (1-based) from a file; useful for large files. |
| `listFiles` | List files and directories at a given path. |
| `listFilesRecursive` | List all files/dirs under a path recursively; optional max depth. |
| `edit_file` | Edit a file by replacing one string with another (all occurrences). |
| `create_file` | Create a new file with given content; creates parent dirs if needed. |
| `remove_file` | Delete a file at the given path. |
| `searchFile` | Find file(s) by name under a directory. |
| `grepInFile` | Search for a string in a single file; returns matching lines with line numbers. |
| `grepInFiles` | Search for a pattern in files under a directory; optional glob filter (e.g. `*.go`). |
| `runCommand` | Run a shell command; returns stdout, stderr, and exit code; optional working directory. |
| `getWorkingDir` | Return the current working directory path. |
| `moveFile` | Move or rename a file to a new path. |
| `copyFile` | Copy a file to another path. |
| `fileInfo` | Return metadata for a path: size, mtime, is-dir, permissions. |
| `createDirectory` | Create a directory (and parents); like `mkdir -p`. |
| `removeDirectory` | Remove a directory; optional recursive. |
| `searchInternet` | Search the internet; returns titles, URLs, and snippets (no API key required). |
| `fetchHtml` | Fetch the HTML or text body of a URL. |
| `fetchFile` | Download a file from a URL; optional save path (otherwise returns body or summary). |
| `clear_context` | Clear conversation history so the next message starts fresh (internal/special). |

---

## Getting started

### CLI

1. Set your API key:
   ```bash
   export ANTHROPIC_API_KEY="your-key-here"
   ```
2. Build and run from the project root:
   ```bash
   go build -o agentExample .
   ./agentExample
   ```
3. Type messages and press Enter. The agent can read files, edit them, run commands, search the web, etc., using the tools above. Use `/clear` or `/reset` to clear conversation context.

### VS Code extension

The `extension/` folder contains a VS Code extension that opens a chat panel powered by the same agent.

1. **API key**: Set `agentExample.apiKey` in VS Code settings, or set `ANTHROPIC_API_KEY` in your environment.
2. **Binary**: From the repo root, run `npm run build:agent` inside `extension/` to build the Go agent into `extension/bin/`, or set `agentExample.agentPath` to the path of your `agentExample` binary.
3. **Use**: Run the “agentExample Chat” command from the Command Palette (Ctrl/Cmd+Shift+P), or open the “agentExample Chat” view in the sidebar.

For more details, see [extension/README.md](extension/README.md).

---

## Project layout

- **`main.go`** — CLI entrypoint and agent loop (conversation, tool use detection, tool execution, streaming).
- **`tools/`** — Tool definitions: each file provides a `ToolDefinition` (name, description, input schema, handler) for one or more tools.
- **`extension/`** — VS Code extension (TypeScript) for the chat UI; spawns the Go binary and communicates via stdin/stdout.

---

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
