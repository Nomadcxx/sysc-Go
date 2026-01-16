#!/usr/bin/env node
import {WebSocket} from 'ws';
import process from 'process';

const ws = new WebSocket('ws://localhost:3000');
const stdin = process.stdin;
const stdout = process.stdout;

// Helper to send native message
function sendNativeMessage(msg) {
	const buffer = Buffer.from(JSON.stringify(msg));
	const header = Buffer.alloc(4);
	header.writeUInt32LE(buffer.length, 0);
	stdout.write(header);
	stdout.write(buffer);
}

// WS -> Native
ws.on('open', () => {
	// Connected to FLOYD CLI
});

ws.on('message', data => {
	try {
		const msg = JSON.parse(data.toString());
		sendNativeMessage(msg);
	} catch (e) {
		// ignore
	}
});

ws.on('error', err => {
	// Log to stderr so it doesn't break native messaging (which uses stdout)
	console.error('WS Error:', err);
	process.exit(1);
});

ws.on('close', () => {
	process.exit(0);
});

// Native -> WS
let messageLength = null;
let buffer = Buffer.alloc(0);

stdin.on('readable', () => {
	let chunk;
	while (null !== (chunk = stdin.read())) {
		buffer = Buffer.concat([buffer, chunk]);

		while (true) {
			if (messageLength === null) {
				if (buffer.length >= 4) {
					messageLength = buffer.readUInt32LE(0);
					buffer = buffer.slice(4);
				} else {
					break;
				}
			}

			if (messageLength !== null) {
				if (buffer.length >= messageLength) {
					const msgBuffer = buffer.slice(0, messageLength);
					buffer = buffer.slice(messageLength);
					messageLength = null;

					try {
						const msg = JSON.parse(msgBuffer.toString());
						if (ws.readyState === WebSocket.OPEN) {
							ws.send(JSON.stringify(msg));
						}
					} catch (e) {
						// ignore
					}
				} else {
					break;
				}
			}
		}
	}
});
