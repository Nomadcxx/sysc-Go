import Anthropic from '@anthropic-ai/sdk';
import {MCPClientManager} from '../mcp/client.js';
import {SessionManager, SessionData} from './session.js';
import {PermissionManager} from './permissions.js';
import {Config} from '../utils/config.js';

export type Message = {
	role: 'user' | 'assistant' | 'system' | 'tool';
	content: string | any[];
	tool_call_id?: string;
	name?: string;
};

export class AgentEngine {
	private anthropic: Anthropic;
	private mcpManager: MCPClientManager;
	private sessionManager: SessionManager;
	private permissionManager: PermissionManager;
	private config: Config;
	private currentSession: SessionData | null = null;
	public history: Message[] = [];

	constructor(
		apiKey: string,
		mcpManager: MCPClientManager,
		sessionManager: SessionManager,
		permissionManager: PermissionManager,
		config: Config,
	) {
		this.anthropic = new Anthropic({
			apiKey: apiKey,
			baseURL: 'https://api.z.ai/api/anthropic', // GLM-4.7 via z.ai proxy
		});
		this.mcpManager = mcpManager;
		this.sessionManager = sessionManager;
		this.permissionManager = permissionManager;
		this.config = config;
	}

	async initSession(cwd: string) {
		// Try to resume last session or create new one
		// For MVP, we just create a new one for now, or load if ID provided
		// TODO: logic to load specific session
		this.currentSession = await this.sessionManager.createSession(cwd);

		// Set system prompt
		this.history = [{role: 'system', content: this.config.systemPrompt}];
	}

	async *sendMessage(content: string): AsyncGenerator<string, void, unknown> {
		this.history.push({role: 'user', content});
		if (this.currentSession) {
			this.currentSession.messages = this.history;
			await this.sessionManager.saveSession(this.currentSession);
		}

		let currentTurnDone = false;
		const MAX_TURNS = 10;
		let turns = 0;

		while (!currentTurnDone && turns < MAX_TURNS) {
			turns++;
			const tools = await this.mcpManager.listTools();

			// Transform MCP tools to Anthropic tools format
			const anthropicTools = tools.map(tool => ({
				name: tool.name,
				description: tool.description,
				input_schema: tool.inputSchema,
			}));

			// Separate system prompt from history for Anthropic API
			const systemMessage = this.history.find(m => m.role === 'system');
			const conversationHistory = this.history.filter(m => m.role !== 'system');

			const stream = await this.anthropic.messages.create({
				model: 'claude-opus-4', // Maps to GLM-4.7 via z.ai proxy
				messages: conversationHistory as any[],
				system: systemMessage?.content as string || this.config.systemPrompt,
				tools: anthropicTools.length > 0 ? anthropicTools : undefined,
				stream: true,
				max_tokens: 4096,
			});

			let assistantContent = '';
			let toolCalls: any[] = [];
			let currentBlock: any = null;

			for await (const chunk of stream) {
				if (chunk.type === 'content_block_start') {
					if (chunk.content_block?.type === 'tool_use') {
						currentBlock = {
							id: chunk.content_block.id,
							name: chunk.content_block.name,
							input: '',
						};
					}
				}

				if (chunk.type === 'content_block_delta') {
					if (chunk.delta?.type === 'text_delta') {
						assistantContent += chunk.delta.text;
						yield chunk.delta.text;
					} else if (chunk.delta?.type === 'input_json_delta' && currentBlock) {
						currentBlock.input += chunk.delta.partial_json;
					}
				}

				if (chunk.type === 'content_block_stop' && currentBlock?.name) {
					toolCalls.push({
						id: currentBlock.id,
						name: currentBlock.name,
						input: JSON.parse(currentBlock.input || '{}'),
					});
					currentBlock = null;
				}
			}

			// Add assistant message to history
			const assistantMessage: any = {
				role: 'assistant',
				content: assistantContent,
			};

			if (toolCalls.length > 0) {
				// Build content array with tool_use blocks
				assistantMessage.content = [
					{type: 'text', text: assistantContent},
					...toolCalls.map(tc => ({
						type: 'tool_use',
						id: tc.id,
						name: tc.name,
						input: tc.input,
					})),
				];
			}

			this.history.push(assistantMessage);
			if (this.currentSession) {
				this.currentSession.messages = this.history;
				await this.sessionManager.saveSession(this.currentSession);
			}

			// Process tool calls
			if (toolCalls.length > 0) {
				for (const tc of toolCalls) {
					yield `\n[Requesting tool: ${tc.name}]\n`;

					// Permission Check
					const permission = this.permissionManager.checkPermission(tc.name);
					if (permission === 'deny') {
						yield `\n[Permission denied for tool: ${tc.name}]\n`;
						this.history.push({
							role: 'user',
							content: [
								{
									type: 'tool_result',
									tool_use_id: tc.id,
									content: 'Error: Permission denied by user configuration.',
								},
							],
						});
						continue;
					}

					try {
						const result = await this.mcpManager.callTool(tc.name, tc.input);

						this.history.push({
							role: 'user',
							content: [
								{
									type: 'tool_result',
									tool_use_id: tc.id,
									content: JSON.stringify(result),
								},
							],
						});
					} catch (error: any) {
						this.history.push({
							role: 'user',
							content: [
								{
									type: 'tool_result',
									tool_use_id: tc.id,
									content: `Error: ${error.message}`,
								},
							],
						});
					}
				}
			} else {
				currentTurnDone = true;
			}
		}
	}
}
