import fs from 'fs-extra';
import path from 'path';
// import os from 'os';
import {v4 as uuidv4} from 'uuid';
import {Message} from './engine.js';

export interface SessionData {
	id: string;
	created: number;
	updated: number;
	messages: Message[];
	workingDirectory: string;
}

export class SessionManager {
	private sessionsDir: string;

	constructor() {
		// Use local directory for sessions to avoid permission issues in sandbox
		// this.sessionsDir = path.join(os.homedir(), '.floyd-cli', 'sessions');
		this.sessionsDir = path.join(process.cwd(), '.floyd', 'sessions');
		fs.ensureDirSync(this.sessionsDir);
	}

	async createSession(cwd: string): Promise<SessionData> {
		const id = uuidv4();
		const session: SessionData = {
			id,
			created: Date.now(),
			updated: Date.now(),
			messages: [],
			workingDirectory: cwd,
		};
		await this.saveSession(session);
		return session;
	}

	async saveSession(session: SessionData): Promise<void> {
		session.updated = Date.now();
		const filePath = path.join(this.sessionsDir, `${session.id}.json`);
		await fs.writeJson(filePath, session, {spaces: 2});
	}

	async loadSession(id: string): Promise<SessionData | null> {
		const filePath = path.join(this.sessionsDir, `${id}.json`);
		if (!(await fs.pathExists(filePath))) return null;
		return await fs.readJson(filePath);
	}

	async listSessions(): Promise<SessionData[]> {
		if (!(await fs.pathExists(this.sessionsDir))) return [];
		const files = await fs.readdir(this.sessionsDir);
		const sessions: SessionData[] = [];
		for (const file of files) {
			if (file.endsWith('.json')) {
				try {
					const session = await fs.readJson(path.join(this.sessionsDir, file));
					sessions.push(session);
				} catch (e) {
					// Ignore corrupted files
				}
			}
		}
		return sessions.sort((a, b) => b.updated - a.updated);
	}

	async getLastSession(): Promise<SessionData | null> {
		const sessions = await this.listSessions();
		return sessions.length > 0 ? sessions[0] || null : null;
	}
}
