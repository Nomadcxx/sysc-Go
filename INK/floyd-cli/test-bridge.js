import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const bridgePath = path.join(__dirname, 'bin', 'chrome-bridge.js');

console.log('Spawning bridge:', bridgePath);
const bridge = spawn('node', [bridgePath], {
    stdio: ['pipe', 'pipe', 'inherit']
});

// Helper to write native message
function writeMessage(msg) {
    const buffer = Buffer.from(JSON.stringify(msg));
    const header = Buffer.alloc(4);
    header.writeUInt32LE(buffer.length, 0);
    bridge.stdin.write(header);
    bridge.stdin.write(buffer);
}

// Read from bridge
let messageLength = null;
let buffer = Buffer.alloc(0);

bridge.stdout.on('data', (chunk) => {
    console.log('Received data from bridge:', chunk.length, 'bytes');
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
                
                const msg = JSON.parse(msgBuffer.toString());
                console.log('Received message from bridge:', msg);
            } else {
                break;
            }
        }
    }
});

bridge.on('exit', (code) => {
    console.log('Bridge exited with code:', code);
    process.exit(code);
});

// Send a test message after a delay
setTimeout(() => {
    console.log('Sending test message...');
    writeMessage({
        jsonrpc: '2.0',
        method: 'ping',
        id: 1
    });
}, 1000);

// Keep alive for a bit then exit
setTimeout(() => {
    console.log('Test finished');
    bridge.kill();
    process.exit(0);
}, 5000);
