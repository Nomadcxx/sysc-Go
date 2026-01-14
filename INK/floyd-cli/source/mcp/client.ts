import {Client} from '@modelcontextprotocol/sdk/client/index.js';
import {StdioClientTransport} from '@modelcontextprotocol/sdk/client/stdio.js';
import {ListToolsResult} from '@modelcontextprotocol/sdk/types.js';
import {WebSocketServer} from 'ws';
import {WebSocketConnectionTransport} from './transport.js';
import {v4 as uuidv4} from 'uuid';

export class MCPClientManager {
	private clients: Map<string, Client>;
	private wss: WebSocketServer | null = null;

	constructor() {
		this.clients = new Map();
	}

	async startServer(port: number = 3000) {
		this.wss = new WebSocketServer({port});
		console.log(`MCP Server listening on port ${port}`);

		this.wss.on('connection', async ws => {
			console.log('New MCP connection from Chrome');
			const transport = new WebSocketConnectionTransport(ws);
			const clientId = `chrome-${uuidv4()}`;

			const client = new Client(
				{
					name: 'floyd-cli',
					version: '0.1.0',
				},
				{
					capabilities: {
						sampling: {},
					},
				},
			);

			try {
				await client.connect(transport);
				this.clients.set(clientId, client);

				ws.on('close', () => {
					console.log(`MCP connection ${clientId} closed`);
					this.clients.delete(clientId);
				});
			} catch (e) {
				console.error('Failed to connect to MCP client', e);
				ws.close();
			}
		});
	}

	async connectStdio(name: string, command: string, args: string[] = []) {
		const transport = new StdioClientTransport({
			command,
			args,
		});

		const client = new Client(
			{
				name: 'floyd-cli',
				version: '0.1.0',
			},
			{
				capabilities: {
					sampling: {},
				},
			},
		);

		await client.connect(transport);
		this.clients.set(name, client);
		return client;
	}

	async listTools(): Promise<any[]> {
		let allTools: any[] = [];
		for (const client of this.clients.values()) {
			try {
				const result = (await client.listTools()) as ListToolsResult;
				allTools = [...allTools, ...result.tools];
			} catch (e) {
				console.error('Failed to list tools', e);
			}
		}
		return allTools;
	}

	async callTool(name: string, args: any): Promise<any> {
		// Find which client has this tool
		// For simplicity, we search all clients. In production, we should map tools to clients.
		for (const client of this.clients.values()) {
			try {
				const tools = (await client.listTools()) as ListToolsResult;
				if (tools.tools.find(t => t.name === name)) {
					return await client.callTool({
						name,
						arguments: args,
					});
				}
			} catch (e) {
				// continue
			}
		}
		throw new Error(`Tool ${name} not found`);
	}
}
