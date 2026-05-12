// ===== Floating Window =====
let floatingWindow = null;
let floatingVisible = false;

function createFloatingWindow() {
  if (floatingWindow) return;

  floatingWindow = document.createElement('div');
  floatingWindow.id = 'opc-agent-floating';
  floatingWindow.innerHTML = `
    <style>
      #opc-agent-floating {
        position: fixed;
        top: 80px;
        right: 20px;
        width: 380px;
        max-height: 520px;
        background: white;
        border: 1px solid #E2E8F0;
        border-radius: 12px;
        box-shadow: 0 8px 32px rgba(0,0,0,0.15);
        z-index: 2147483647;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
        font-size: 13px;
        color: #1E293B;
        display: none;
        flex-direction: column;
        resize: both;
        overflow: hidden;
      }
      #opc-agent-floating.show { display: flex; }
      #opc-agent-floating .header {
        display: flex; align-items: center; justify-content: space-between;
        padding: 12px 16px; border-bottom: 1px solid #E2E8F0;
        background: #1E3A5F; color: white; border-radius: 12px 12px 0 0;
        cursor: move; user-select: none;
      }
      #opc-agent-floating .header h3 { margin: 0; font-size: 14px; font-weight: 600; }
      #opc-agent-floating .header button {
        background: none; border: none; color: white; cursor: pointer; font-size: 18px; padding: 0 4px;
      }
      #opc-agent-floating .body { padding: 12px 16px; flex: 1; overflow-y: auto; }
      #opc-agent-floating textarea {
        width: 100%; min-height: 60px; max-height: 120px;
        border: 1px solid #E2E8F0; border-radius: 8px; padding: 8px 12px;
        font-size: 13px; font-family: inherit; resize: vertical; outline: none;
        box-sizing: border-box;
      }
      #opc-agent-floating textarea:focus { border-color: #0891B2; }
      #opc-agent-floating .agent-select {
        width: 100%; padding: 6px 8px; border: 1px solid #E2E8F0;
        border-radius: 6px; font-size: 12px; margin-bottom: 8px; outline: none;
      }
      #opc-agent-floating .btn {
        display: inline-flex; align-items: center; gap: 4px;
        padding: 7px 16px; border-radius: 6px; font-size: 12px; font-weight: 500;
        cursor: pointer; border: none; margin-top: 8px;
      }
      #opc-agent-floating .btn-primary { background: #1E3A5F; color: white; }
      #opc-agent-floating .btn-primary:hover { background: #2D5A8E; }
      #opc-agent-floating .btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
      #opc-agent-floating .result {
        margin-top: 10px; padding: 10px 12px; background: #F8FAFC;
        border-radius: 8px; border: 1px solid #E2E8F0;
        font-size: 12px; line-height: 1.6; white-space: pre-wrap; word-break: break-word;
        display: none; max-height: 250px; overflow-y: auto;
      }
      #opc-agent-floating .result.show { display: block; }
      #opc-agent-floating .status {
        font-size: 11px; color: #94A3B8; margin-top: 6px; display: none;
      }
      #opc-agent-floating .status.show { display: block; }
      #opc-agent-floating .error {
        margin-top: 8px; padding: 8px 12px; background: #FEF2F2;
        border-radius: 6px; color: #DC2626; font-size: 12px; display: none;
      }
      #opc-agent-floating .error.show { display: block; }
    </style>
    <div class="header" id="opc-float-header">
      <h3>🤖 OPC-Agent</h3>
      <button id="opc-float-close">✕</button>
    </div>
    <div class="body">
      <select class="agent-select" id="opc-float-agent">
        <option value="copywriter">✍️ 文案生成</option>
        <option value="email_sorter">📧 邮件分类</option>
        <option value="xhs_poster">📕 小红书</option>
        <option value="competitive_analysis">📊 竞品分析</option>
        <option value="meeting_minutes">📋 会议纪要</option>
      </select>
      <textarea id="opc-float-input" placeholder="输入任务描述，或选中网页文本后按 Alt+Shift+O..."></textarea>
      <button class="btn btn-primary" id="opc-float-submit">➤ 执行</button>
      <div class="status" id="opc-float-status"></div>
      <div class="result" id="opc-float-result"></div>
      <div class="error" id="opc-float-error"></div>
    </div>
  `;

  document.body.appendChild(floatingWindow);
  bindFloatingEvents();

  // Make draggable
  const header = floatingWindow.querySelector('#opc-float-header');
  let isDragging = false, startX, startY, origX, origY;

  header.addEventListener('mousedown', (e) => {
    isDragging = true;
    const rect = floatingWindow.getBoundingClientRect();
    startX = e.clientX; startY = e.clientY;
    origX = rect.left; origY = rect.top;
    document.addEventListener('mousemove', onDrag);
    document.addEventListener('mouseup', () => {
      isDragging = false;
      document.removeEventListener('mousemove', onDrag);
    });
  });

  function onDrag(e) {
    if (!isDragging) return;
    floatingWindow.style.left = `${origX + e.clientX - startX}px`;
    floatingWindow.style.top = `${origY + e.clientY - startY}px`;
    floatingWindow.style.right = 'auto';
  }
}

function bindFloatingEvents() {
  document.getElementById('opc-float-close').onclick = () => hideFloating();

  document.getElementById('opc-float-submit').onclick = async () => {
    const input = document.getElementById('opc-float-input');
    const agent = document.getElementById('opc-float-agent');
    const result = document.getElementById('opc-float-result');
    const status = document.getElementById('opc-float-status');
    const error = document.getElementById('opc-float-error');
    const btn = document.getElementById('opc-float-submit');

    if (!input.value.trim()) return;

    btn.disabled = true;
    btn.textContent = '⏳ 处理中...';
    result.classList.remove('show');
    result.textContent = '';
    error.classList.remove('show');
    error.textContent = '';
    status.classList.add('show');
    status.textContent = '⏳ 正在调用 Agent...';

    try {
      const response = await chrome.runtime.sendMessage({
        action: 'call-agent',
        text: input.value,
        agent: agent.value,
      });

      if (!response.success) throw new Error(response.error);

      const output = typeof response.data.result === 'string'
        ? response.data.result
        : JSON.stringify(response.data.result, null, 2);

      result.textContent = output;
      result.classList.add('show');
      status.textContent = `✅ 完成 (${response.data.tokens?.input_tokens || 0} tokens)`;
    } catch (err) {
      error.textContent = `❌ ${err.message}`;
      error.classList.add('show');
      status.textContent = '';
    }

    btn.disabled = false;
    btn.textContent = '➤ 执行';
  };

  // Auto-resize textarea
  document.getElementById('opc-float-input').addEventListener('input', function() {
    this.style.height = 'auto';
    this.style.height = Math.min(this.scrollHeight, 120) + 'px';
  });
}

function showFloating(selectedText) {
  if (!floatingWindow) createFloatingWindow();
  floatingWindow.classList.add('show');
  floatingVisible = true;

  const input = document.getElementById('opc-float-input');
  if (selectedText) {
    input.value = selectedText;
    input.style.height = 'auto';
    input.style.height = Math.min(input.scrollHeight, 120) + 'px';
  }
  input.focus();
}

function hideFloating() {
  if (floatingWindow) {
    floatingWindow.classList.remove('show');
    floatingVisible = false;
  }
}

// ===== Message Listener =====
chrome.runtime.onMessage.addListener((request) => {
  if (request.action === 'toggle-floating-window') {
    if (floatingVisible) {
      hideFloating();
    } else {
      const selection = window.getSelection()?.toString() || '';
      showFloating(selection);
    }
  }
});
