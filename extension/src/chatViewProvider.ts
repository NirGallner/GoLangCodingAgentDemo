import * as vscode from 'vscode';
import * as path from 'path';
import { getOrCreateAgentProcess } from './extension';

let activePanel: vscode.WebviewPanel | null = null;

export function getOrCreateChatPanel(context: vscode.ExtensionContext): vscode.WebviewPanel {
	if (activePanel) {
		activePanel.reveal(vscode.ViewColumn.Beside);
		return activePanel;
	}

	const panel = vscode.window.createWebviewPanel(
		'agentexample.chatView',
		'agentExample Chat',
		vscode.ViewColumn.Beside,
		{ enableScripts: true, retainContextWhenHidden: true }
	);

	activePanel = panel;
	panel.onDidDispose(() => {
		activePanel = null;
	});

	panel.webview.options = {
		enableScripts: true,
		localResourceRoots: [context.extensionUri]
	};
	panel.webview.html = getHtml(panel.webview);

	const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;

	panel.webview.onDidReceiveMessage(async (msg) => {
		switch (msg.type) {
			case 'send':
				await handleSend(panel, context, msg.text);
				break;
			case 'injectMainGo':
				await handleInjectMainGo(panel);
				break;
		}
	});

	if (workspaceRoot) {
		panel.webview.postMessage({ type: 'ready', workspaceRoot });
	} else {
		panel.webview.postMessage({ type: 'error', message: 'No workspace folder open.' });
	}

	return panel;
}

async function handleSend(
	panel: vscode.WebviewPanel,
	context: vscode.ExtensionContext,
	text: string
): Promise<void> {
	const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
	if (!workspaceRoot) {
		panel.webview.postMessage({ type: 'error', message: 'No workspace folder open.' });
		return;
	}
	const agent = getOrCreateAgentProcess(context, workspaceRoot);
	if (!agent) {
		return;
	}
	agent.start();
	panel.webview.postMessage({ type: 'agentThinking' });
	try {
		const result = await agent.sendMessage(text);
		panel.webview.postMessage({
			type: 'agentTurn',
			messages: result.messages,
			toolCalls: result.toolCalls
		});
	} catch (e) {
		const message = e instanceof Error ? e.message : String(e);
		panel.webview.postMessage({ type: 'error', message });
	}
}

async function handleInjectMainGo(panel: vscode.WebviewPanel): Promise<void> {
	const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
	if (!workspaceRoot) {
		return;
	}
	const mainPath = path.join(workspaceRoot, 'main.go');
	try {
		const doc = await vscode.workspace.openTextDocument(mainPath);
		const content = doc.getText();
		const injectText = `Context: the user's main.go:\n\`\`\`go\n${content}\n\`\`\``;
		panel.webview.postMessage({ type: 'injectMainGoContent', text: injectText });
	} catch {
		panel.webview.postMessage({ type: 'error', message: 'main.go not found in workspace root.' });
	}
}

function getHtml(webview: vscode.Webview): string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>agentExample Chat</title>
	<style>
		* { box-sizing: border-box; }
		body { margin: 0; padding: 8px; font-family: var(--vscode-font-family); font-size: var(--vscode-font-size); color: var(--vscode-foreground); background: var(--vscode-editor-background); }
		h2 { margin: 0 0 8px 0; font-size: 1.1em; }
		#messages { flex: 1; overflow-y: auto; min-height: 120px; max-height: 50vh; padding: 8px 0; }
		.msg { margin: 6px 0; padding: 6px 8px; border-radius: 6px; }
		.msg.user { background: var(--vscode-input-background); }
		.msg.agent { background: var(--vscode-editor-inactiveSelectionBackground); white-space: pre-wrap; word-break: break-word; }
		.msg.tool { font-size: 0.9em; color: var(--vscode-descriptionForeground); }
		#inputRow { display: flex; gap: 6px; margin-top: 8px; }
		#input { flex: 1; padding: 6px 8px; border: 1px solid var(--vscode-input-border); background: var(--vscode-input-background); color: var(--vscode-input-foreground); border-radius: 4px; }
		button { padding: 6px 12px; background: var(--vscode-button-background); color: var(--vscode-button-foreground); border: none; border-radius: 4px; cursor: pointer; }
		button:hover { background: var(--vscode-button-hoverBackground); }
		button.secondary { background: var(--vscode-button-secondaryBackground); color: var(--vscode-button-secondaryForeground); }
		#thinking { font-style: italic; color: var(--vscode-descriptionForeground); margin: 4px 0; }
	</style>
</head>
<body>
	<h2>agentExample Chat</h2>
	<div id="messages"></div>
	<div id="thinking" style="display:none;">Agent is thinking…</div>
	<div id="inputRow">
		<input id="input" type="text" placeholder="Type a message…" />
		<button id="send">Send</button>
		<button id="mainGo" class="secondary" title="Insert main.go context">main.go</button>
	</div>
	<script>
		const vscode = acquireVsCodeApi();
		const messagesEl = document.getElementById('messages');
		const inputEl = document.getElementById('input');
		const thinkingEl = document.getElementById('thinking');

		function appendMessage(role, text, isTool) {
			const div = document.createElement('div');
			div.className = 'msg ' + (isTool ? 'tool' : role);
			div.textContent = text;
			messagesEl.appendChild(div);
			messagesEl.scrollTop = messagesEl.scrollHeight;
		}

		window.addEventListener('message', e => {
			const msg = e.data;
			switch (msg.type) {
				case 'ready':
					appendMessage('agent', 'Ready. You can ask about your project or main.go.', false);
					break;
				case 'error':
					appendMessage('agent', 'Error: ' + msg.message, false);
					thinkingEl.style.display = 'none';
					break;
				case 'agentThinking':
					thinkingEl.style.display = 'block';
					break;
				case 'agentTurn':
					thinkingEl.style.display = 'none';
					(msg.toolCalls || []).forEach(t => appendMessage('tool', 'tool: ' + t.name + '(' + (t.input || '') + ')', true));
					(msg.messages || []).forEach(m => appendMessage('agent', m.text || m, false));
					break;
				case 'injectMainGoContent':
					inputEl.value = msg.text;
					inputEl.focus();
					break;
			}
		});

		document.getElementById('send').onclick = () => send();
		inputEl.onkeydown = (e) => { if (e.key === 'Enter') send(); };
		function send() {
			const text = inputEl.value.trim();
			if (!text) return;
			appendMessage('user', text, false);
			inputEl.value = '';
			vscode.postMessage({ type: 'send', text });
		}

		document.getElementById('mainGo').onclick = () => vscode.postMessage({ type: 'injectMainGo' });
	</script>
</body>
</html>`;
}
