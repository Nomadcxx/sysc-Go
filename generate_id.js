
const crypto = require('crypto');

// Generate an RSA key pair
const { publicKey, privateKey } = crypto.generateKeyPairSync('rsa', {
  modulusLength: 2048,
  publicKeyEncoding: {
    type: 'spki',
    format: 'pem'
  },
  privateKeyEncoding: {
    type: 'pkcs8',
    format: 'pem'
  }
});

// Remove PEM headers/footers and newlines to get the base64 string
const publicKeyString = publicKey
  .replace('-----BEGIN PUBLIC KEY-----', '')
  .replace('-----END PUBLIC KEY-----', '')
  .replace(/[\r\n]/g, '');

// Calculate Extension ID from public key
// 1. Decode base64 to buffer
const pubKeyBuf = Buffer.from(publicKeyString, 'base64');
// 2. SHA256 hash
const hash = crypto.createHash('sha256').update(pubKeyBuf).digest('hex');
// 3. First 32 chars, mapped a-p
const id = hash.slice(0, 32).split('').map(char => {
  // map 0-9a-f to a-p
  const code = parseInt(char, 16);
  return String.fromCharCode(97 + code);
}).join('');

console.log('KEY:' + publicKeyString);
console.log('ID:' + id);
