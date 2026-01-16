/**
 * MCP Protocol Message Handling
 * Implements JSON-RPC 2.0 for Model Context Protocol
 */

export class MCPMessageHandler {
  constructor() {
    this.requestId = 0;
    this.pendingRequests = new Map();
  }

  /**
   * Create a JSON-RPC request message
   */
  createRequest(method, params = {}) {
    const id = ++this.requestId;
    return {
      jsonrpc: '2.0',
      id,
      method,
      params
    };
  }

  /**
   * Create a JSON-RPC response message
   */
  createResponse(id, result = null, error = null) {
    const response = {
      jsonrpc: '2.0',
      id
    };

    if (error) {
      response.error = {
        code: error.code || -32603,
        message: error.message || 'Internal error',
        data: error.data
      };
    } else {
      response.result = result;
    }

    return response;
  }

  /**
   * Create a JSON-RPC notification (no response expected)
   */
  createNotification(method, params = {}) {
    return {
      jsonrpc: '2.0',
      method,
      params
    };
  }

  /**
   * Parse incoming message
   */
  parseMessage(message) {
    try {
      const parsed = typeof message === 'string' ? JSON.parse(message) : message;
      
      if (!parsed.jsonrpc || parsed.jsonrpc !== '2.0') {
        throw new Error('Invalid JSON-RPC version');
      }

      return parsed;
    } catch (error) {
      throw new Error(`Failed to parse MCP message: ${error.message}`);
    }
  }

  /**
   * Validate message structure
   */
  validateMessage(message) {
    if (!message.jsonrpc || message.jsonrpc !== '2.0') {
      return { valid: false, error: 'Invalid jsonrpc version' };
    }

    if (message.method && !message.id) {
      // Notification
      return { valid: true, isNotification: true };
    }

    if (message.method && message.id) {
      // Request
      if (!message.method || typeof message.method !== 'string') {
        return { valid: false, error: 'Invalid method' };
      }
      return { valid: true, isRequest: true };
    }

    if (message.id && (message.result !== undefined || message.error !== undefined)) {
      // Response
      return { valid: true, isResponse: true };
    }

    return { valid: false, error: 'Invalid message structure' };
  }
}
