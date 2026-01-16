import {Transport} from '@modelcontextprotocol/sdk/shared/transport.js';
import {JSONRPCMessage} from '@modelcontextprotocol/sdk/types.js';
import WebSocket from 'ws';

export class WebSocketConnectionTransport implements Transport {
	private _ws: WebSocket;

	onclose?: () => void;
	onerror?: (error: Error) => void;
	onmessage?: (message: JSONRPCMessage) => void;

	constructor(ws: WebSocket) {
		this._ws = ws;

		this._ws.on('message', data => {
			try {
				const json = JSON.parse(data.toString());
				this.onmessage?.(json);
			} catch (e) {
				this.onerror?.(e as Error);
			}
		});

		this._ws.on('close', () => {
			this.onclose?.();
		});

		this._ws.on('error', err => {
			this.onerror?.(err);
		});
	}

	async start(): Promise<void> {
		if (this._ws.readyState === WebSocket.OPEN) {
			return;
		}
		return new Promise((resolve, reject) => {
			this._ws.once('open', () => resolve());
			this._ws.once('error', reject);
		});
	}

	async send(message: JSONRPCMessage): Promise<void> {
		this._ws.send(JSON.stringify(message));
	}

	async close(): Promise<void> {
		this._ws.close();
	}
}
