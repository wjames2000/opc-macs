// ===== Configuration =====
const DEFAULT_API_URL = 'http://localhost:8080/api/v1';

async function getConfig() {
  const { apiUrl, apiKey, defaultAgent } = await chrome.storage.sync.get({
    apiUrl: DEFAULT_API_URL,
    apiKey: '',
    defaultAgent: 'copywriter',
  });
  return { apiUrl, apiKey, defaultAgent };
}

// ===== API Client =====
async function callAgent(text, agent) {
  const config = await getConfig();
  const url = `${config.apiUrl}/agents/execute`;

  try {
    const res = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(config.apiKey ? { 'Authorization': `Bearer ${config.apiKey}` } : {}),
      },
      body: JSON.stringify({ agent: agent || config.defaultAgent, input: text }),
    });

    if (!res.ok) {
      const err = await res.text();
      throw new Error(`API Error: ${res.status} - ${err}`);
    }

    return await res.json();
  } catch (err) {
    throw new Error(`无法连接到 OPC-Agent 服务 (${url}): ${err.message}`);
  }
}

// ===== Context Menu =====
chrome.runtime.onInstalled.addListener(() => {
  chrome.contextMenus.create({
    id: 'opc-agent-process',
    title: '交给 Agent 处理',
    contexts: ['selection'],
  });

  chrome.contextMenus.create({
    id: 'opc-agent-copywriter',
    title: '✍️ 生成文案',
    contexts: ['selection'],
    parentId: 'opc-agent-process',
  });

  chrome.contextMenus.create({
    id: 'opc-agent-analysis',
    title: '📊 竞品分析',
    contexts: ['selection'],
    parentId: 'opc-agent-process',
  });

  chrome.contextMenus.create({
    id: 'opc-agent-summary',
    title: '📋 会议纪要',
    contexts: ['selection'],
    parentId: 'opc-agent-process',
  });
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  const agentMap = {
    'opc-agent-copywriter': 'copywriter',
    'opc-agent-analysis': 'competitive_analysis',
    'opc-agent-summary': 'meeting_minutes',
  };

  const agent = agentMap[info.menuItemId] || 'copywriter';
  const text = info.selectionText;

  if (!text) return;

  try {
    const result = await callAgent(text, agent);
    const output = typeof result.result === 'string' ? result.result : JSON.stringify(result.result, null, 2);

    // Store result for popup to show
    await chrome.storage.local.set({
      lastResult: { agent, input: text, output, time: Date.now() },
    });

    // Show notification
    chrome.notifications.create({
      type: 'basic',
      iconUrl: 'icons/icon48.png',
      title: `OPC-Agent: ${agent}`,
      message: `✅ 处理完成 (${result.tokens?.input_tokens || 0} tokens)`,
    });
  } catch (err) {
    chrome.notifications.create({
      type: 'basic',
      iconUrl: 'icons/icon48.png',
      title: 'OPC-Agent 错误',
      message: err.message,
    });
  }
});

// ===== Commands (Alt+Shift+O) =====
chrome.commands.onCommand.addListener(async (command) => {
  if (command === 'open-floating-window') {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
    if (tab?.id) {
      chrome.tabs.sendMessage(tab.id, { action: 'toggle-floating-window' });
    }
  }
});

// ===== Message Handler =====
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.action === 'call-agent') {
    callAgent(request.text, request.agent)
      .then(result => sendResponse({ success: true, data: result }))
      .catch(err => sendResponse({ success: false, error: err.message }));
    return true; // Keep message channel open
  }

  if (request.action === 'get-last-result') {
    chrome.storage.local.get('lastResult').then(data => {
      sendResponse(data.lastResult || null);
    });
    return true;
  }
});
