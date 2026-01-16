import fs from 'fs-extra';
import path from 'path';

export interface Config {
	systemPrompt: string;
	allowedTools: string[];
	mcpServers: Record<string, any>;
}

export class ConfigLoader {
	static async loadProjectConfig(cwd: string = process.cwd()): Promise<Config> {
		const config: Config = {
			systemPrompt: 'You are a helpful AI assistant.',
			allowedTools: [],
			mcpServers: {},
		};

		// Load CLAUDE.md
		const claudeMdPath = path.join(cwd, 'CLAUDE.md');
		if (await fs.pathExists(claudeMdPath)) {
			const content = await fs.readFile(claudeMdPath, 'utf-8');
			config.systemPrompt += `\n\nProject Context (from CLAUDE.md):\n${content}`;
		}

		// Load .claude/settings.json (or .floyd/settings.json)
		const settingsPath = path.join(cwd, '.floyd', 'settings.json');
		if (await fs.pathExists(settingsPath)) {
			try {
				const settings = await fs.readJson(settingsPath);
				if (settings.systemPrompt)
					config.systemPrompt += `\n${settings.systemPrompt}`;
				if (settings.allowedTools) config.allowedTools = settings.allowedTools;
				if (settings.mcpServers) config.mcpServers = settings.mcpServers;
			} catch (e) {
				console.error('Failed to parse settings.json', e);
			}
		}

		return config;
	}
}
