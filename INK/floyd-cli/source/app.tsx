import React, {useState, useEffect, useRef} from 'react';
import {Box, Text, useInput, useApp} from 'ink';
import TextInput from 'ink-text-input';
import Spinner from 'ink-spinner';
import {AgentEngine} from './agent/engine.js';
import {MCPClientManager} from './mcp/client.js';
import {SessionManager} from './agent/session.js';
import {ConfigLoader} from './utils/config.js';
import {PermissionManager} from './agent/permissions.js';
import dotenv from 'dotenv';

dotenv.config();

type Message = {
	role: 'user' | 'assistant' | 'system' | 'tool';
	content: string | any[];
};

type AppProps = {
	name?: string;
	chrome?: boolean;
};

export default function App({name = 'User', chrome = false}: AppProps) {
	const [input, setInput] = useState('');
	const [messages, setMessages] = useState<Message[]>([]);
	const [isThinking, setIsThinking] = useState(false);
	const [status, setStatus] = useState('Initializing...');
	const engineRef = useRef<AgentEngine | null>(null);
	const {exit} = useApp();

	useEffect(() => {
		const init = async () => {
			try {
				const mcpManager = new MCPClientManager();
				if (chrome) {
					await mcpManager.startServer(3000); // Default port for Chrome bridge
					setStatus('Listening for Chrome on port 3000...');
				}

				const sessionManager = new SessionManager();
				const config = await ConfigLoader.loadProjectConfig();
				const permissionManager = new PermissionManager(config.allowedTools);

				// Connect to MCP servers defined in config
				// TODO: Implement MCP server connection from config

				const apiKey = process.env['GLM_API_KEY'] || 'dummy-key';

				if (process.env['GLM_API_KEY'] === undefined) {
					// We warn but continue with dummy for UI testing if requested
					// or we could block.
				}

				engineRef.current = new AgentEngine(
					apiKey,
					mcpManager,
					sessionManager,
					permissionManager,
					config,
				);
				await engineRef.current.initSession(process.cwd());

				setStatus('Ready');

				// Initial greeting
				setMessages([
					{
						role: 'assistant',
						content:
							'Hello! I am Floyd (GLM-4 Powered). How can I help you today?',
					},
				]);
			} catch (error: any) {
				setStatus(`Initialization failed: ${error.message}`);
			}
		};
		init();
	}, []);

	const handleSubmit = async (value: string) => {
		if (!value.trim() || !engineRef.current) return;
		if (isThinking) return;

		const userMsg: Message = {role: 'user', content: value};
		setMessages(prev => [...prev, userMsg]);
		setInput('');
		setIsThinking(true);
		setStatus('Thinking...');

		try {
			const generator = engineRef.current.sendMessage(value);
			let currentAssistantMessage = '';

			// Add placeholder for assistant message
			setMessages(prev => [...prev, {role: 'assistant', content: ''}]);

			for await (const chunk of generator) {
				currentAssistantMessage += chunk;
				// Update last message
				setMessages(prev => {
					const newMessages = [...prev];
					const lastMsg = newMessages[newMessages.length - 1];
					if (lastMsg && lastMsg.role === 'assistant') {
						lastMsg.content = currentAssistantMessage;
					}
					return newMessages;
				});
			}
		} catch (error: any) {
			setMessages(prev => [
				...prev,
				{role: 'assistant', content: `Error: ${error.message}`},
			]);
		} finally {
			setIsThinking(false);
			setStatus('Ready');
		}
	};

	useInput((_input, key) => {
		if (key.escape) {
			exit();
		}
	});

	return (
		<Box flexDirection="column" padding={1} width="100%" height="100%">
			<Box borderStyle="round" borderColor="cyan" paddingX={1} marginBottom={1}>
				<Text bold color="cyan">
					FLOYD CLI
				</Text>
				<Text> | {status}</Text>
			</Box>

			<Box flexDirection="column" flexGrow={1} overflowY="hidden">
				{messages.map((msg, index) => (
					<Box key={index} flexDirection="column" marginBottom={1}>
						<Text bold color={msg.role === 'user' ? 'green' : 'blue'}>
							{msg.role === 'user' ? name : 'Floyd'}:
						</Text>
						<Text>
							{typeof msg.content === 'string'
								? msg.content
								: JSON.stringify(msg.content)}
						</Text>
					</Box>
				))}
			</Box>

			{isThinking && (
				<Box marginBottom={1}>
					<Text color="yellow">
						<Spinner type="dots" /> Thinking...
					</Text>
				</Box>
			)}

			<Box borderStyle="single" borderColor="gray" paddingX={1}>
				<Text color="green">❯ </Text>
				<TextInput
					value={input}
					onChange={setInput}
					onSubmit={handleSubmit}
					placeholder={isThinking ? 'Please wait...' : 'Type a message...'}
				/>
			</Box>

			<Box marginTop={1}>
				<Text color="gray" dimColor>
					Press Esc to exit
				</Text>
			</Box>
		</Box>
	);
}
