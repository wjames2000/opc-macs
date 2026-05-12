const DEFAULT_API_URL = 'http://localhost:8080/api/v1';

async function getConfig() {
  return chrome.storage.sync.get({
    apiUrl: DEFAULT_API_URL,
    apiKey: '',
    defaultAgent: 'copywriter',
  });
}

async function callAgent(text, agent) {
  const config = await getConfig();
  const url = `${config.apiUrl}/agents/execute`;

  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(config.apiKey ? { 'Authorization': `Bearer ${config.apiKey}` } : {}),
    },
    body: JSON.stringify({ agent, input: text }),
  });

  if (!res.ok) {
    const err = await res.text();
    throw new Error(`API Error (${res.status}): ${err.slice(0, 100)}`);
  }

  return res.json();
}

// ===== UI Elements =====
const inputEl = document.getElementById('inputText');
const agentEl = document.getElementById('agentSelect');
const submitBtn = document.getElementById('submitBtn');
const statusEl = document.getElementById('statusMsg');
const resultEl = document.getElementById('resultBox');
const errorEl = document.getElementById('errorBox');
const tokenInfo = document.getElementById('tokenInfo');

// Load defaults
getConfig().then(config => {
  agentEl.value = config.defaultAgent;
});

// Submit
submitBtn.addEventListener('click', async () => {
  const text = inputEl.value.trim();
  if (!text) return;

  submitBtn.disabled = true;
  submitBtn.textContent = '⏳ 处理中...';
  resultEl.classList.remove('show');
  errorEl.classList.remove('show');
  statusEl.classList.add('show');
  statusEl.textContent = '⏳ 正在调用 Agent...';

  try {
    const data = await callAgent(text, agentEl.value);
    const output = typeof data.result === 'string'
      ? data.result
      : JSON.stringify(data.result, null, 2);

    resultEl.textContent = output;
    resultEl.classList.add('show');
    statusEl.textContent = `✅ 完成`;
    tokenInfo.textContent = `Token: ${data.tokens?.input_tokens || 0} in / ${data.tokens?.output_tokens || 0} out`;

  } catch (err) {
    errorEl.textContent = `❌ ${err.message}`;
    errorEl.classList.add('show');
    statusEl.textContent = '';
  }

  submitBtn.disabled = false;
  submitBtn.textContent = '➤ 执行';
});

// Enter to submit
inputEl.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    submitBtn.click();
  }
});

// Open options
document.getElementById('openOptions').addEventListener('click', (e) => {
  e.preventDefault();
  chrome.runtime.openOptionsPage();
});
