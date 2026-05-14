import React, { useState, useEffect } from 'react';

const providers = [
  { value: 'deepseek', label: 'DeepSeek', url: 'https://api.deepseek.com' },
  { value: 'openai', label: 'OpenAI', url: 'https://api.openai.com' },
  { value: 'qwen', label: '通义千问 (Qwen)', url: 'https://dashscope.aliyuncs.com/compatible-mode' },
  { value: 'kimi', label: 'Kimi (Moonshot)', url: 'https://api.moonshot.cn' },
  { value: 'claude', label: 'Claude (Anthropic)', url: 'https://api.anthropic.com' },
  { value: 'gemini', label: 'Gemini (Google)', url: 'https://generativelanguage.googleapis.com' },
];

const defaultModels = [
  { name: 'deepseek-v4-flash', provider: 'DeepSeek', url: 'api.deepseek.com', default: true },
  { name: 'gemini-2.0-flash', provider: 'Google', url: 'generativelanguage.googleapis.com', default: false },
  { name: 'gpt-4o-mini', provider: 'OpenAI', url: 'api.openai.com', default: false },
];

export default function Settings({ onClose }) {
  const [settings, setSettings] = useState({
    api_base_url: '', api_key: '', provider: 'deepseek',
    model_name: 'deepseek-v4-flash', default_agent: 'copywriter',
  });
  const [saving, setSaving] = useState(false);
  const [models, setModels] = useState(defaultModels);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: '', provider: '', url: '', key: '', temp: '0.3', maxTokens: '4096' });

  useEffect(() => {
    if (window.go?.main?.App?.GetSettings) {
      window.go.main.App.GetSettings().then(s => {
        if (s) setSettings(s);
      }).catch(console.warn);
    }
  }, []);

  const saveSettings = async () => {
    setSaving(true);
    if (window.go?.main?.App?.SaveSettings) {
      const ok = await window.go.main.App.SaveSettings(settings);
      if (ok && onClose) onClose();
    }
    setSaving(false);
  };

  const addModel = () => {
    const p = providers.find(p => p.value === form.provider);
    setModels([...models, {
      name: form.name,
      provider: p?.label || form.provider,
      url: form.url,
      default: false,
      id: Date.now(),
    }]);
    setShowForm(false);
    setForm({ name: '', provider: '', url: '', key: '', temp: '0.3', maxTokens: '4096' });
  };

  const removeModel = (idx) => {
    setModels(models.filter((_, i) => i !== idx));
  };

  return (
    <div>
      <h2 className="text-lg font-semibold mb-1">⚙️ 系统设置</h2>
      <p className="text-sm text-gray-500 mb-6">管理模型连接、默认 Agent 等系统参数</p>

      {/* Active Model Config */}
      <div className="bg-white rounded-xl border p-5 mb-6">
        <h3 className="font-semibold mb-4">🔌 当前连接</h3>
        <div className="space-y-3">
          <div>
            <label className="text-xs text-gray-500 block mb-1">Provider</label>
            <select className="w-full border rounded-lg px-3 py-2 text-sm outline-none"
              value={settings.provider}
              onChange={e => {
                const p = providers.find(p => p.value === e.target.value);
                setSettings({ ...settings, provider: e.target.value, api_base_url: p?.url || '' });
              }}>
              {providers.map(p => <option key={p.value} value={p.value}>{p.label}</option>)}
            </select>
          </div>
          <div>
            <label className="text-xs text-gray-500 block mb-1">模型名称</label>
            <input className="w-full border rounded-lg px-3 py-2 text-sm outline-none"
              value={settings.model_name} onChange={e => setSettings({ ...settings, model_name: e.target.value })} />
          </div>
          <div>
            <label className="text-xs text-gray-500 block mb-1">API 地址</label>
            <input className="w-full border rounded-lg px-3 py-2 text-sm outline-none" placeholder="留空自动使用默认地址"
              value={settings.api_base_url} onChange={e => setSettings({ ...settings, api_base_url: e.target.value })} />
          </div>
          <div>
            <label className="text-xs text-gray-500 block mb-1">API Key</label>
            <input className="w-full border rounded-lg px-3 py-2 text-sm outline-none" type="password"
              value={settings.api_key} onChange={e => setSettings({ ...settings, api_key: e.target.value })} />
          </div>
          <div>
            <label className="text-xs text-gray-500 block mb-1">默认 Agent</label>
            <select className="w-full border rounded-lg px-3 py-2 text-sm outline-none"
              value={settings.default_agent} onChange={e => setSettings({ ...settings, default_agent: e.target.value })}>
              <option value="copywriter">✍️ 文案生成</option>
              <option value="email_sorter">📧 邮件分类</option>
              <option value="xhs_poster">📕 小红书</option>
              <option value="competitive_analysis">📊 竞品分析</option>
              <option value="meeting_minutes">📋 会议纪要</option>
            </select>
          </div>
          <div className="flex gap-2 pt-2">
            <button className="bg-[#1E3A5F] text-white px-5 py-2 rounded-lg text-sm" onClick={saveSettings} disabled={saving}>
              {saving ? '保存中...' : '💾 保存'}
            </button>
            {onClose && <button className="border px-5 py-2 rounded-lg text-sm" onClick={onClose}>取消</button>}
          </div>
        </div>
      </div>

      {/* Model Library */}
      <div className="bg-white rounded-xl border p-5">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold">📚 模型库</h3>
          <button className="bg-[#1E3A5F] text-white px-3 py-1.5 rounded-lg text-xs" onClick={() => setShowForm(!showForm)}>
            + 添加模型
          </button>
        </div>

        <div className="space-y-2 mb-4">
          {models.map((m, i) => (
            <div key={i} className="flex items-center gap-3 border rounded-lg p-3 hover:border-cyan-500 transition-colors">
              <div className="w-9 h-9 rounded-lg bg-blue-50 flex items-center justify-center text-sm font-bold">
                {m.name.charAt(0).toUpperCase()}
              </div>
              <div className="flex-1">
                <div className="text-sm font-medium">{m.name} {m.default && <span className="text-xs text-cyan-600 ml-1">默认</span>}</div>
                <div className="text-xs text-gray-400">{m.provider} · {m.url}</div>
              </div>
              <span className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">● 可用</span>
              {!m.default && (
                <button className="text-xs text-red-400 hover:text-red-600" onClick={() => removeModel(i)}>🗑️</button>
              )}
            </div>
          ))}
        </div>

        {showForm && (
          <div className="p-4 bg-gray-50 rounded-lg space-y-3">
            <h4 className="text-sm font-medium">添加模型</h4>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-gray-600 block mb-1">Provider</label>
                <select className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.provider}
                  onChange={e => {
                    const p = providers.find(p => p.value === e.target.value);
                    setForm({ ...form, provider: e.target.value, url: p?.url || '' });
                  }}>
                  <option value="">— 选择 —</option>
                  {providers.map(p => <option key={p.value} value={p.value}>{p.label}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-gray-600 block mb-1">模型名称</label>
                <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.name}
                  onChange={e => setForm({ ...form, name: e.target.value })} placeholder="deepseek-chat" />
              </div>
            </div>
            <div>
              <label className="text-xs text-gray-600 block mb-1">API 地址</label>
              <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.url}
                onChange={e => setForm({ ...form, url: e.target.value })} />
            </div>
            <div className="flex gap-2 pt-1">
              <button className="bg-[#1E3A5F] text-white px-4 py-1.5 rounded-lg text-xs" onClick={addModel}>✅ 确认添加</button>
              <button className="border px-4 py-1.5 rounded-lg text-xs" onClick={() => setShowForm(false)}>取消</button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
