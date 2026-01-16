/**
 * MCP Server Implementation
 * Handles communication with FLOYD CLI via Native Messaging
 */

import { MCPMessageHandler } from './messages.js';
import { ToolExecutor } from '../tools/executor.js';
import { SafetyLayer } from '../safety/permissions.js';

export class MCPServer {
  constructor() {
    this.messageHandler = new MCPMessageHandler();
    this.toolExecutor = new ToolExecutor();
    this.safetyLayer = new SafetyLayer();
    this.nativePort = null;
    this.isConnected = false;
    this.attachedTabs = new Map(); // tabId -> debugger attachment state
  }

  /**
   * Initialize Native Messaging connection
   */
  async connect() {
    try {
      const hostName = 'com.floyd.chrome';
      this.nativePort = chrome.runtime.connectNative(hostName);

      this.nativePort.onMessage.addListener((message) => {
        this.handleIncomingMessage(message);
      });

      this.nativePort.onDisconnect.addListener(() => {
        this.handleDisconnect();
      });

      this.isConnected = true;
      console.log('[MCP] Connected to FLOYD CLI');
      
      // Send initialization notification
      this.sendNotification('initialized', {
        protocolVersion: '2024-11-05',
        capabilities: {
          tools: this.toolExecutor.listTools()
        }
      });

      return true;
    } catch (error) {
      console.error('[MCP] Connection failed:', error);
      this.isConnected = false;
      return false;
    }
  }

  /**
   * Handle incoming messages from FLOYD CLI
   */
  async handleIncomingMessage(message) {
    try {
      const parsed = this.messageHandler.parseMessage(message);
      const validation = this.messageHandler.validateMessage(parsed);

      if (!validation.valid) {
        this.sendError(parsed.id, -32600, 'Invalid Request', validation.error);
        return;
      }

      if (validation.isNotification) {
        await this.handleNotification(parsed);
        return;
      }

      if (validation.isRequest) {
        await this.handleRequest(parsed);
        return;
      }

      // Response - handle if needed (for bidirectional communication)
      if (validation.isResponse) {
        this.handleResponse(parsed);
        return;
      }
    } catch (error) {
      console.error('[MCP] Error handling message:', error);
      this.sendError(null, -32700, 'Parse error', error.message);
    }
  }

  /**
   * Handle JSON-RPC request
   */
  async handleRequest(message) {
    const { id, method, params } = message;

    try {
      // Safety check
      const safetyCheck = await this.safetyLayer.checkAction(method, params);
      if (!safetyCheck.allowed) {
        this.sendError(id, -32000, 'Safety check failed', safetyCheck.reason);
        return;
      }

      // Execute tool
      const result = await this.toolExecutor.execute(method, params);
      
      // Log action
      await this.logAction(method, params, result);

      this.sendResponse(id, result);
    } catch (error) {
      console.error(`[MCP] Error executing ${method}:`, error);
      this.sendError(id, -32603, 'Internal error', error.message);
    }
  }

  /**
   * Handle JSON-RPC notification
   */
  async handleNotification(message) {
    const { method, params } = message;

    switch (method) {
      case 'ping':
        this.sendNotification('pong', { timestamp: Date.now() });
        break;
      default:
        console.warn(`[MCP] Unknown notification: ${method}`);
    }
  }

  /**
   * Handle JSON-RPC response
   */
  handleResponse(message) {
    // Handle responses if needed for bidirectional communication
    console.log('[MCP] Received response:', message);
  }

  /**
   * Send JSON-RPC response
   */
  sendResponse(id, result) {
    if (!this.isConnected || !this.nativePort) return;

    const response = this.messageHandler.createResponse(id, result);
    this.sendMessage(response);
  }

  /**
   * Send JSON-RPC error
   */
  sendError(id, code, message, data = null) {
    if (!this.isConnected || !this.nativePort) return;

    const response = this.messageHandler.createResponse(id, null, {
      code,
      message,
      data
    });
    this.sendMessage(response);
  }

  /**
   * Send JSON-RPC notification
   */
  sendNotification(method, params = {}) {
    if (!this.isConnected || !this.nativePort) return;

    const notification = this.messageHandler.createNotification(method, params);
    this.sendMessage(notification);
  }

  /**
   * Send message to native host
   */
  sendMessage(message) {
    if (!this.nativePort) {
      console.error('[MCP] Cannot send message: not connected');
      return;
    }

    try {
      this.nativePort.postMessage(message);
    } catch (error) {
      console.error('[MCP] Failed to send message:', error);
      this.handleDisconnect();
    }
  }

  /**
   * Handle disconnection
   */
  handleDisconnect() {
    this.isConnected = false;
    this.nativePort = null;
    
    // Detach debugger from all tabs
    for (const tabId of this.attachedTabs.keys()) {
      this.detachDebugger(tabId).catch(console.error);
    }
    this.attachedTabs.clear();

    console.log('[MCP] Disconnected from FLOYD CLI');
  }

  /**
   * Attach debugger to tab (required for some tools)
   */
  async attachDebugger(tabId) {
    if (this.attachedTabs.has(tabId)) {
      return true;
    }

    try {
      await chrome.debugger.attach({ tabId }, '1.0');
      this.attachedTabs.set(tabId, true);
      return true;
    } catch (error) {
      console.error(`[MCP] Failed to attach debugger to tab ${tabId}:`, error);
      return false;
    }
  }

  /**
   * Detach debugger from tab
   */
  async detachDebugger(tabId) {
    if (!this.attachedTabs.has(tabId)) {
      return true;
    }

    try {
      await chrome.debugger.detach({ tabId });
      this.attachedTabs.delete(tabId);
      return true;
    } catch (error) {
      console.error(`[MCP] Failed to detach debugger from tab ${tabId}:`, error);
      return false;
    }
  }

  /**
   * Log action for audit trail
   */
  async logAction(method, params, result) {
    const logEntry = {
      timestamp: Date.now(),
      method,
      params,
      result: result ? { success: true } : { success: false }
    };

    // Store in chrome.storage.local (last 1000 actions)
    const { actionLog = [] } = await chrome.storage.local.get('actionLog');
    actionLog.push(logEntry);
    
    // Keep only last 1000 entries
    if (actionLog.length > 1000) {
      actionLog.shift();
    }

    await chrome.storage.local.set({ actionLog });
  }
}
