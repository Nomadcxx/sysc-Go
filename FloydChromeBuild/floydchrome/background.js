/**
 * Background Service Worker
 * Main entry point for FloydChrome extension
 */

import { MCPServer } from './mcp/server.js';
import { FloydAgent } from './agent/floyd.js';

let mcpServer = null;
let floydAgent = null;

// Initialize on extension startup
chrome.runtime.onStartup.addListener(initialize);
chrome.runtime.onInstalled.addListener(initialize);

async function initialize() {
  console.log('[FloydChrome] Initializing...');

  // Initialize MCP Server for FLOYD CLI connection
  mcpServer = new MCPServer();
  
  // Initialize FLOYD Agent stub (will be wired up later)
  floydAgent = new FloydAgent();

  // Attempt to connect to FLOYD CLI
  const connected = await mcpServer.connect();
  
  if (connected) {
    console.log('[FloydChrome] MCP Server connected to FLOYD CLI');
  } else {
    console.log('[FloydChrome] MCP Server waiting for FLOYD CLI connection');
  }

  // Set up side panel action
  chrome.action.onClicked.addListener((tab) => {
    chrome.sidePanel.open({ windowId: tab.windowId });
  });
}

// Handle messages from content scripts or side panel
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === 'get_mcp_status') {
    sendResponse({
      connected: mcpServer?.isConnected || false
    });
    return true;
  }

  if (message.type === 'get_tool_metadata') {
    if (mcpServer) {
      const metadata = mcpServer.toolExecutor.getAllToolMetadata();
      sendResponse({ metadata });
      return true;
    }
    sendResponse({ metadata: {} });
    return true;
  }

  return false;
});

// Keep service worker alive
chrome.runtime.onConnect.addListener((port) => {
  port.onDisconnect.addListener(() => {
    // Connection closed
  });
});

// Export for testing/debugging
if (typeof globalThis !== 'undefined') {
  globalThis.floydChrome = {
    mcpServer: () => mcpServer,
    floydAgent: () => floydAgent
  };
}
