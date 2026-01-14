/**
 * Navigation Tools
 */

export class NavigationTools {
  /**
   * Navigate to a URL
   */
  static async navigate(params) {
    const { url, tabId } = params;

    if (!url) {
      throw new Error('URL is required');
    }

    // Validate URL
    try {
      new URL(url);
    } catch {
      throw new Error(`Invalid URL: ${url}`);
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

    // Navigate
    await chrome.tabs.update(targetTabId, { url });

    // Wait for navigation to complete
    return new Promise((resolve) => {
      const listener = (updatedTabId, changeInfo) => {
        if (updatedTabId === targetTabId && changeInfo.status === 'complete') {
          chrome.tabs.onUpdated.removeListener(listener);
          resolve({
            success: true,
            tabId: targetTabId,
            url
          });
        }
      };
      chrome.tabs.onUpdated.addListener(listener);
      
      // Timeout after 30 seconds
      setTimeout(() => {
        chrome.tabs.onUpdated.removeListener(listener);
        resolve({
          success: true,
          tabId: targetTabId,
          url,
          warning: 'Navigation timeout - page may still be loading'
        });
      }, 30000);
    });
  }
}
