import * as child_process from 'child_process';
import * as readline from 'readline';

const PROMPT_LINE = 'You: ';
const AGENT_PREFIX = 'Agent: ';
const TOOL_PREFIX = 'tool: ';

function stripAnsi(s: string): string {
	return s.replace(/\x1b\[[0-9;]*m/g, '');
}

export interface AgentTurnMessage {
	text: string;
}

export interface AgentToolCall {
	name: string;
	input?: string;
}

export interface AgentTurnResult {
	messages: AgentTurnMessage[];
	toolCalls: AgentToolCall[];
}

export class AgentProcess {
	private process: child_process.ChildProcess | null = null;
	private disposed = false;

	constructor(
		private readonly binPath: string,
		private readonly cwd: string,
		private readonly apiKey: string
	) {}

	isDisposed(): boolean {
		return this.disposed;
	}

	sendMessage(userMessage: string): Promise<AgentTurnResult> {
		return new Promise((resolve, reject) => {
			if (this.disposed || !this.process || !this.process.stdin) {
				reject(new Error('Agent process not running'));
				return;
			}
			const messages: AgentTurnMessage[] = [];
			const toolCalls: AgentToolCall[] = [];
			const rl = readline.createInterface({
				input: this.process.stdout!,
				crlfDelay: Infinity
			});

			const onLine = (line: string) => {
				const plain = stripAnsi(line);
				if (plain === PROMPT_LINE || plain.startsWith(PROMPT_LINE)) {
					const isInitialPrompt = messages.length === 0 && toolCalls.length === 0;
					if (isInitialPrompt) {
						return;
					}
					rl.removeListener('line', onLine);
					rl.close();
					resolve({ messages, toolCalls });
					return;
				}
				if (plain.startsWith(TOOL_PREFIX)) {
					const rest = plain.slice(TOOL_PREFIX.length);
					const open = rest.indexOf('(');
					const name = open >= 0 ? rest.slice(0, open).trim() : rest.trim();
					let input: string | undefined;
					if (open >= 0) {
						input = rest.slice(open + 1);
						if (input.endsWith(')')) {
							input = input.slice(0, -1);
						}
					}
					toolCalls.push({ name, input });
					return;
				}
				if (plain.startsWith(AGENT_PREFIX)) {
					messages.push({ text: plain.slice(AGENT_PREFIX.length) });
					return;
				}
				// Continuation of agent text (Go prints "Agent: first line" then "line2", "line3", ...)
				if (messages.length > 0) {
					const last = messages[messages.length - 1];
					last.text += '\n' + plain;
				}
			};

			rl.on('line', onLine);
			rl.on('close', () => {
				if (messages.length === 0 && toolCalls.length === 0) {
					reject(new Error('Agent process ended unexpectedly'));
				}
			});

			this.process.once('error', (err) => {
				rl.removeListener('line', onLine);
				reject(err);
			});
			this.process.stdin.write(userMessage + '\n', (err) => {
				if (err) {
					rl.removeListener('line', onLine);
					reject(err);
				}
			});
		});
	}

	start(): void {
		if (this.process || this.disposed) {
			return;
		}
		this.process = child_process.spawn(this.binPath, [], {
			cwd: this.cwd,
			env: { ...process.env, ANTHROPIC_API_KEY: this.apiKey },
			stdio: ['pipe', 'pipe', 'pipe']
		});
		this.process.on('error', (err) => {
			console.error('agentExample process error:', err);
		});
		this.process.on('exit', (code, signal) => {
			this.process = null;
		});
	}

	dispose(): void {
		this.disposed = true;
		if (this.process) {
			this.process.kill();
			this.process = null;
		}
	}
}
