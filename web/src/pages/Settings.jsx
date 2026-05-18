import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';

const DEFAULT_PROVIDERS = [
  { value: 'openai', label: 'OpenAI', url: 'https://api.openai.com' },
  { value: 'deepseek', label: 'DeepSeek', url: 'https://api.deepseek.com' },
  { value: 'qwen', label: '通义千问 (Qwen)', url: 'https://dashscope.aliyuncs.com/compatible-mode' },
  { value: 'kimi', label: 'Kimi (Moonshot)', url: 'https://api.moonshot.cn' },
  { value: 'claude', label: 'Claude (Anthropic)', url: 'https://api.anthropic.com' },
  { value: 'gemini', label: 'Gemini (Google)', url: 'https://generativelanguage.googleapis.com' },
];

const DEFAULT_MODELS = [
  { id: 1, name: 'deepseek-v4-flash', provider: 'DeepSeek', url: 'api.deepseek.com', default: true, color: 'bg-blue-50', icon: 'D' },
  { id: 2, name: 'gemini-2.0-flash', provider: 'Google', url: 'generativelanguage.googleapis.com', default: false, color: 'bg-yellow-50', icon: 'G' },
  { id: 3, name: 'gpt-4o-mini', provider: 'OpenAI', url: 'api.openai.com', default: false, color: 'bg-green-50', icon: 'O' },
];

export default function Settings() {
  const [models, setModels] = useState(DEFAULT_MODELS);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: '', provider: '', url: '', key: '', temp: '0.3', maxTokens: '4096' });

  useEffect(() => {
    api.listAgents?.().then(agents => {
      if (agents?.length) {
        const modelList = agents.map((a, i) => ({
          id: i + 1,
          name: a.name,
          provider: a.provider || 'Unknown',
          url: a.model_url || '',
          default: i === 0,
          color: ['bg-blue-50', 'bg-yellow-50', 'bg-green-50', 'bg-purple-50'][i % 4],
          icon: a.name.charAt(0).toUpperCase(),
        }));
        if (modelList.length > 0) setModels(modelList);
      }
    }).catch(() => {});
  }, []);

  const addModel = () => {
    const p = DEFAULT_PROVIDERS.find(p => p.value === form.provider);
    setModels([...models, { id: Date.now(), name: form.name, provider: p?.label || form.provider, url: form.url, default: false, color: 'bg-purple-50', icon: form.name.charAt(0).toUpperCase() }]);
    setShowForm(false);
    setForm({ name: '', provider: '', url: '', key: '', temp: '0.3', maxTokens: '4096' });
  };

  const removeModel = (id) => setModels(models.filter(m => m.id !== id));

  return (
    <div>
      <h2 className="text-lg font-semibold mb-1">⚙️ 系统设置</h2>
      <p className="text-sm text-gray-500 mb-6">管理多个模型连接、配置系统参数和存储后端。</p>

      <div className="bg-white rounded-xl border p-5 mb-6">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold">🤖 可用模型</h3>
          <button className="bg-blue-800 text-white px-3 py-1.5 rounded-lg text-xs" onClick={() => setShowForm(!showForm)}>+ 添加模型</button>
        </div>
        <div className="space-y-2">
          {models.map(m => (
            <div key={m.id} className="flex items-center gap-3 border rounded-lg p-3 hover:border-cyan-500 transition-colors">
              <div className={`w-9 h-9 rounded-lg ${m.color} flex items-center justify-center text-sm font-bold`}>{m.icon}</div>
              <div className="flex-1">
                <div className="text-sm font-medium">{m.name} {m.default && <span className="text-xs text-cyan-600 ml-1">默认</span>}</div>
                <div className="text-xs text-gray-400">{m.provider} · {m.url}</div>
              </div>
              {!m.default && <button className="text-xs text-red-500 hover:text-red-700" onClick={() => removeModel(m.id)}>移除</button>}
            </div>
          ))}
        </div>

        {showForm && (
          <div className="mt-4 border-t pt-4">
            <div className="grid grid-cols-2 gap-3 mb-3">
              <input className="border rounded-lg px-3 py-2 text-sm" placeholder="模型名称" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} />
              <select className="border rounded-lg px-3 py-2 text-sm" value={form.provider} onChange={e => setForm({ ...form, provider: e.target.value, url: DEFAULT_PROVIDERS.find(p => p.value === e.target.value)?.url || '' })}>
                <option value="">选择提供商</option>
                {DEFAULT_PROVIDERS.map(p => <option key={p.value} value={p.value}>{p.label}</option>)}
              </select>
              <input className="border rounded-lg px-3 py-2 text-sm" placeholder="API URL" value={form.url} onChange={e => setForm({ ...form, url: e.target.value })} />
              <input className="border rounded-lg px-3 py-2 text-sm" type="password" placeholder="API Key" value={form.key} onChange={e => setForm({ ...form, key: e.target.value })} />
              <input className="border rounded-lg px-3 py-2 text-sm" placeholder="温度 (0-1)" value={form.temp} onChange={e => setForm({ ...form, temp: e.target.value })} />
              <input className="border rounded-lg px-3 py-2 text-sm" placeholder="最大 Token" value={form.maxTokens} onChange={e => setForm({ ...form, maxTokens: e.target.value })} />
            </div>
            <button className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm" onClick={addModel}>确认添加</button>
          </div>
        )}
      </div>
    </div>
  );
}
