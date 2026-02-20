import * as vscode from 'vscode';
import * as path from 'path';
import * as fs from 'fs';

import { AgentProcess } from './agentProcess';
import { getOrCreateChatPanel } from './chatViewProvider';

let agentProcess: AgentProcess | null = null;

export function activate(context: vscode.ExtensionContext) {
	context.subscriptions.push(
		vscode.commands.registerCommand('agentexample.openChat', () => {
			getOrCreateChatPanel(context);
		})
	);

	context.subscriptions.push({
		dispose: () => {
			if (agentProcess) {
				agentProcess.dispose();
				agentProcess = null;
			}
		}
	});
}

export function deactivate() {
	if (agentProcess) {
		agentProcess.dispose();
		agentProcess = null;
	}
}

export function getOrCreateAgentProcess(context: vscode.ExtensionContext, workspaceRoot: string | undefined): AgentProcess | null {
	if (!workspaceRoot) {
		return null;
	}
	if (!agentProcess || agentProcess.isDisposed()) {
		const config = vscode.workspace.getConfiguration('agentExample');
		const apiKey = config.get<string>('apiKey') || process.env.ANTHROPIC_API_KEY;
		const customPath = config.get<string>('agentPath');
		const binPath = customPath || getBundledBinPath(context);
		if (!binPath) {
			vscode.window.showErrorMessage('agentExample: No binary found. Set agentExample.agentPath or run the build script to bundle the binary.');
			return null;
		}
		if (!apiKey) {
			vscode.window.showErrorMessage('agentExample: Set agentExample.apiKey or ANTHROPIC_API_KEY.');
			return null;
		}
		agentProcess = new AgentProcess(binPath, workspaceRoot, apiKey);
	}
	return agentProcess;
}

function getBundledBinPath(context: vscode.ExtensionContext): string | null {
	const binDir = path.join(context.extensionPath, 'bin');
	const platform = process.platform;
	const arch = process.arch;
	const name = platform === 'win32' ? 'agentExample.exe' : 'agentExample';
	const candidate = path.join(binDir, name);
	if (fs.existsSync(candidate)) {
		return candidate;
	}
	return path.join(binDir, 'agentExample');
}
