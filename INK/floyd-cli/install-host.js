import fs from 'fs';
import path from 'path';
import os from 'os';
import {fileURLToPath} from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

const HOST_NAME = 'com.floyd.chrome';
const HOST_DESC = 'FLOYD CLI Native Messaging Host';

// Check for Extension ID argument
const extensionId = process.argv[2];
if (!extensionId) {
	console.log('Usage: node install-host.js <chrome-extension-id>');
	console.log('You can find the ID in chrome://extensions/');
	process.exit(1);
}

const EXTENSION_ORIGIN = `chrome-extension://${extensionId}/`;

// Determine path to host script
const hostScript = path.join(__dirname, 'bin', 'chrome-bridge.sh');

// Manifest content
const manifest = {
	name: HOST_NAME,
	description: HOST_DESC,
	path: hostScript,
	type: 'stdio',
	allowed_origins: [EXTENSION_ORIGIN],
};

// Target directory
let targetDir;
if (process.platform === 'darwin') {
	targetDir = path.join(
		os.homedir(),
		'Library/Application Support/Google/Chrome/NativeMessagingHosts',
	);
} else if (process.platform === 'linux') {
	targetDir = path.join(
		os.homedir(),
		'.config/google-chrome/NativeMessagingHosts',
	);
}

if (targetDir) {
	try {
		fs.mkdirSync(targetDir, {recursive: true});
		const manifestPath = path.join(targetDir, `${HOST_NAME}.json`);
		fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2));
		console.log(`✅ Installed Native Messaging Host manifest to:`);
		console.log(`   ${manifestPath}`);
		console.log(`✅ Pointing to bridge script:`);
		console.log(`   ${hostScript}`);
		console.log(`✅ Allowed Extension ID:`);
		console.log(`   ${extensionId}`);
	} catch (err) {
		console.error('Failed to install manifest:', err);
		process.exit(1);
	}
} else {
	console.error('Unsupported platform');
	process.exit(1);
}
