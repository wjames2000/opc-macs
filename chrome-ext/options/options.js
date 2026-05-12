const defaults = {
  apiUrl: 'http://localhost:8080/api/v1',
  apiKey: '',
  defaultAgent: 'copywriter',
};

// Load saved settings
chrome.storage.sync.get(defaults).then(config => {
  document.getElementById('apiUrl').value = config.apiUrl;
  document.getElementById('apiKey').value = config.apiKey;
  document.getElementById('defaultAgent').value = config.defaultAgent;
});

// Save settings
document.getElementById('saveBtn').addEventListener('click', () => {
  chrome.storage.sync.set({
    apiUrl: document.getElementById('apiUrl').value.trim(),
    apiKey: document.getElementById('apiKey').value.trim(),
    defaultAgent: document.getElementById('defaultAgent').value,
  }, () => {
    showToast('✅ 设置已保存');
  });
});

function showToast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2000);
}
