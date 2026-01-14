/**
 * Tab Management Tools
 */

export class TabTools {
  /**
   * Create a new tab
   */
  static async createTab(params) {
    const { url } = params;

    const tab = await chrome.tabs.create({
      url: url || 'chrome://newtab/',
      active: true
    });

    return {
      success: true,
      tabId: tab.id,
      url: tab.url || url
    };
  }

  /**
   * Get all open tabs
   */
  static async getTabs(params) {
    const tabs = await chrome.tabs.query({});

    return {
      success: true,
      tabs: tabs.map(tab => ({
        id: tab.id,
        url: tab.url,
        title: tab.title,
        active: tab.active,
        windowId: tab.windowId
      }))
    };
  }
}
