/**
 * Interaction Tools
 * Click, type, and interact with page elements
 */

export class InteractionTools {
  /**
   * Click element at coordinates or by selector
   */
  static async click(params) {
    const { x, y, selector, tabId } = params;

    // Get target tab
    let targetTabId = tabId;
    if (!targetTabId) {
      const [activeTab] = await chrome.tabs.query({ active: true, currentWindow: true });
      if (!activeTab) {
        throw new Error('No active tab found');
      }
      targetTabId = activeTab.id;
    }

    // Attach debugger
    try {
      await chrome.debugger.attach({ tabId: targetTabId }, '1.0');
    } catch (error) {
      if (!error.message.includes('Another debugger')) {
        throw new Error(`Failed to attach debugger: ${error.message}`);
      }
    }

    try {
      let clickX = x;
      let clickY = y;

      // If selector provided, get element coordinates
      if (selector && (!x || !y)) {
        const results = await chrome.scripting.executeScript({
          target: { tabId: targetTabId },
          func: (sel) => {
            const element = document.querySelector(sel);
            if (!element) return null;
            
            const rect = element.getBoundingClientRect();
            return {
              x: rect.left + rect.width / 2,
              y: rect.top + rect.height / 2
            };
          },
          args: [selector]
        });

        if (!results || !results[0] || !results[0].result) {
          throw new Error(`Element not found: ${selector}`);
        }

        const coords = results[0].result;
        clickX = coords.x;
        clickY = coords.y;
      }

      if (!clickX || !clickY) {
        throw new Error('Either coordinates (x, y) or selector must be provided');
      }

      // Dispatch mouse events
      await chrome.debugger.sendCommand(
        { tabId: targetTabId },
        'Input.dispatchMouseEvent',
        {
          type: 'mousePressed',
          x: clickX,
          y: clickY,
          button: 'left',
          clickCount: 1
        }
      );

      await chrome.debugger.sendCommand(
        { tabId: targetTabId },
        'Input.dispatchMouseEvent',
        {
          type: 'mouseReleased',
          x: clickX,
          y: clickY,
          button: 'left',
          clickCount: 1
        }
      );

      return {
        success: true,
        tabId: targetTabId,
        coordinates: { x: clickX, y: clickY },
        selector: selector || null
      };
    } catch (error) {
      throw new Error(`Failed to click: ${error.message}`);
    }
  }

  /**
   * Type text into focused element
   */
  static async type(params) {
    const { text, tabId } = params;

    if (!text) {
      throw new Error('Text is required');
    }

    // Get target tab
    let targetTabId = tabId;
    if (!targetTabId) {
      const [activeTab] = await chrome.tabs.query({ active: true, currentWindow: true });
      if (!activeTab) {
        throw new Error('No active tab found');
      }
      targetTabId = activeTab.id;
    }

    // Attach debugger
    try {
      await chrome.debugger.attach({ tabId: targetTabId }, '1.0');
    } catch (error) {
      if (!error.message.includes('Another debugger')) {
        throw new Error(`Failed to attach debugger: ${error.message}`);
      }
    }

    try {
      // Insert text using Input.insertText (more reliable than key events)
      await chrome.debugger.sendCommand(
        { tabId: targetTabId },
        'Input.insertText',
        { text }
      );

      return {
        success: true,
        tabId: targetTabId,
        text,
        length: text.length
      };
    } catch (error) {
      throw new Error(`Failed to type: ${error.message}`);
    }
  }
}
