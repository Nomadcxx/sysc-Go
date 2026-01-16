import React, {useState, useEffect, useRef} from 'react';
import {Box, Text, useInput, useApp} from 'ink';
import TextInput from 'ink-text-input';
import Spinner from 'ink-spinner';
import {AgentEngine} from './agent/engine.js';
import {MCPClientManager} from './mcp/client.js';
import {SessionManager} from './agent/session.js';
import {ConfigLoader} from './utils/config.js';
import {PermissionManager} from './agent/permissions.js';
import {floydTheme, floydRoles} from './theme.js';
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

const floydBlocksLines = [
	'████  ██   ███   ██   ██  █████',
	'█     ██  █   █  ██   ██  █    ',
	'████  ██  █████  ██   ██  ████ ',
	'█     ██  █   █   ██ ██   █    ',
	'████  ██  █   █    ███    █████',
];

const floydBlockColors = [
	'17',
	'18',
	'19',
	'20',
	'21',
];

const floydGradientText = 'FLOYD';

const floydGradientColors = [
	'#FF60FF',
	'#D457FF',
	'#B85CFF',
	'#9054FF',
	'#6B50FF',
];

export default function App({name = 'User', chrome = false}: AppProps) {
	const [input, setInput] = useState('');
	const [messages, setMessages] = useState<Message[]>([]);
	const [isThinking, setIsThinking] = useState(false);
	const [status, setStatus] = useState('Initializing...');
	const [showHelp, setShowHelp] = useState(false);
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

	useInput((inputKey, key) => {
		if (key.escape) {
			if (showHelp) {
				setShowHelp(false);
			} else {
				exit();
			}
		}

		if (inputKey === '?') {
			setShowHelp(value => !value);
		}
	});

	if (showHelp) {
		return (
			<Box
				flexDirection="column"
				padding={1}
				width="100%"
				height="100%"
				borderStyle="round"
				borderColor={floydTheme.colors.borderFocus}
			>
				<Text bold color={floydRoles.headerTitle}>
					FLOYD Help
				</Text>
				<Box marginTop={1} flexDirection="column">
					<Text color={floydTheme.colors.fgBase}>
						?: toggle this help view
					</Text>
					<Text color={floydTheme.colors.fgBase}>
						Enter: send message
					</Text>
					<Text color={floydTheme.colors.fgBase}>
						Esc: close help or exit
					</Text>
				</Box>
				<Box marginTop={1}>
					<Text color={floydRoles.hint} dimColor>
						Floyd uses a CharmTone-inspired theme with floating frames and blocks.
					</Text>
				</Box>
			</Box>
		);
	}

	return (
		<Box flexDirection="column" padding={1} width="100%" height="100%">
			<Box flexDirection="column" marginBottom={1}>
				{floydBlocksLines.map((line, rowIndex) => (
					<Box key={rowIndex}>
						{Array.from(line).map((ch, colIndex) => {
							const colorIndex =
								colIndex < 5
									? floydBlockColors[0]
									: colIndex < 11
										? floydBlockColors[1]
										: colIndex < 17
											? floydBlockColors[2]
											: colIndex < 23
												? floydBlockColors[3]
												: floydBlockColors[4];

							const prefix = `\u001B[48;5;${colorIndex}m`;
							const suffix = '\u001B[0m';

							return (
								<Text key={colIndex}>
									{ch === ' ' ? ' ' : `${prefix} ${suffix}`}
								</Text>
							);
						})}
					</Box>
				))}
			</Box>
			<Box
				borderStyle="round"
				borderColor={floydTheme.colors.borderFocus}
				paddingX={1}
				marginBottom={1}
			>
				<Box width="100%" justifyContent="space-between">
					<Box>
						{Array.from(floydGradientText).map((ch, index) => (
							<Text
								key={index}
								bold
								color={floydGradientColors[index] ?? floydRoles.headerTitle}
							>
								{ch}
							</Text>
						))}
						<Text color={floydRoles.headerStatus}> CLI</Text>
					</Box>
					<Text color={floydRoles.headerStatus}>
						{process.cwd()} • {status}
					</Text>
				</Box>
			</Box>

			<Box flexDirection="column" flexGrow={1} overflowY="hidden">
				{messages.map((msg, index) => (
					<Box
						key={index}
						flexDirection="column"
						marginBottom={1}
						borderStyle="single"
						borderColor={
							msg.role === 'assistant'
								? floydTheme.colors.border
								: floydTheme.colors.borderFocus
						}
						paddingX={1}
						paddingY={0}
					>
						<Box>
							<Text
								bold
								color={
									msg.role === 'user'
										? floydRoles.userLabel
										: msg.role === 'assistant'
											? floydRoles.assistantLabel
											: msg.role === 'system'
												? floydRoles.systemLabel
												: floydRoles.toolLabel
								}
							>
								{msg.role === 'user' ? name : 'Floyd'}:
							</Text>
						</Box>
						<Box marginTop={0}>
							<Text color={floydTheme.colors.fgBase}>
								{typeof msg.content === 'string'
									? msg.content
									: JSON.stringify(msg.content)}
							</Text>
						</Box>
					</Box>
				))}
			</Box>

			{isThinking && (
				<Box marginBottom={1}>
					<Text color={floydRoles.thinking}>
						<Spinner type="dots" /> Thinking...
					</Text>
				</Box>
			)}

			<Box
				borderStyle="single"
				borderColor={floydTheme.colors.border}
				paddingX={1}
			>
				<Text color={floydRoles.inputPrompt}>❯ </Text>
				<TextInput
					value={input}
					onChange={setInput}
					onSubmit={handleSubmit}
					placeholder={isThinking ? 'Please wait...' : 'Type a message...'}
				/>
			</Box>

			<Box marginTop={1}>
				<Text color={floydRoles.hint} dimColor>
					Press Esc to exit
				</Text>
			</Box>
		</Box>
	);
}
