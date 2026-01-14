/**
 * Side Panel Script
 * Standalone agent mode UI
 */

// Tab switching
document.querySelectorAll('.tab-button').forEach(button => {
  button.addEventListener('click', () => {
    const tabName = button.dataset.tab;
    
    // Update buttons
    document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
    button.classList.add('active');
    
    // Update content
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    document.getElementById(`${tabName}Tab`).classList.add('active');
    
    // Load tab content
    if (tabName === 'tools') {
      loadTools();
    } else if (tabName === 'logs') {
      loadLogs();
    }
  });
});

// Status check
async function updateStatus() {
  try {
    const response = await chrome.runtime.sendMessage({ type: 'get_mcp_status' });
    const statusIndicator = document.getElementById('statusIndicator');
    const statusText = document.getElementById('statusText');
    
    if (response.connected) {
      statusIndicator.classList.add('connected');
      statusIndicator.classList.remove('disconnected');
      statusText.textContent = 'Connected to FLOYD CLI';
    } else {
      statusIndicator.classList.add('disconnected');
      statusIndicator.classList.remove('connected');
      statusText.textContent = 'Disconnected - Waiting for FLOYD CLI';
    }
  } catch (error) {
    console.error('Failed to get status:', error);
  }
}

// Load tools
async function loadTools() {
  try {
    const response = await chrome.runtime.sendMessage({ type: 'get_tool_metadata' });
    const toolsList = document.getElementById('toolsList');
    
    if (!response.metadata || Object.keys(response.metadata).length === 0) {
      toolsList.innerHTML = '<p>No tools available</p>';
      return;
    }
    
    toolsList.innerHTML = Object.values(response.metadata).map(tool => `
      <div class="tool-item">
        <h3>${tool.name}</h3>
        <p>${tool.description}</p>
      </div>
    `).join('');
  } catch (error) {
    console.error('Failed to load tools:', error);
    document.getElementById('toolsList').innerHTML = '<p>Error loading tools</p>';
  }
}

// Load logs
async function loadLogs() {
  try {
    const { actionLog = [] } = await chrome.storage.local.get('actionLog');
    const logsContainer = document.getElementById('logsContainer');
    
    if (actionLog.length === 0) {
      logsContainer.innerHTML = '<p>No logs yet</p>';
      return;
    }
    
    logsContainer.innerHTML = actionLog.slice(-50).reverse().map(log => {
      const date = new Date(log.timestamp);
      const status = log.result?.success ? 'success' : 'error';
      return `
        <div class="log-entry ${status}">
          <span class="log-timestamp">${date.toLocaleTimeString()}</span>
          <span>${log.method}</span>
        </div>
      `;
    }).join('');
  } catch (error) {
    console.error('Failed to load logs:', error);
  }
}

// Execute task
document.getElementById('executeButton').addEventListener('click', async () => {
  const input = document.getElementById('taskInput');
  const outputArea = document.getElementById('outputArea');
  const button = document.getElementById('executeButton');
  
  const task = input.value.trim();
  if (!task) return;
  
  button.disabled = true;
  outputArea.innerHTML = '<p>Processing task...</p>';
  
  try {
    // STUB: This will call the actual FLOYD agent when wired up
    outputArea.innerHTML = `
      <p><strong>Task:</strong> ${task}</p>
      <p><strong>Status:</strong> FLOYD agent stub - implementation pending</p>
      <p>When the FLOYD agent is wired up, this will execute browser automation tasks.</p>
    `;
  } catch (error) {
    outputArea.innerHTML = `<p class="error">Error: ${error.message}</p>`;
  } finally {
    button.disabled = false;
  }
});

// Initialize
updateStatus();
setInterval(updateStatus, 5000); // Update status every 5 seconds
