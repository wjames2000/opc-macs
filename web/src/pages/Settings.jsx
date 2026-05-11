import React, { useState } from 'react';

const initialModels = [
  { id: 1, name: 'deepseek-v4-flash', provider: 'DeepSeek', url: 'api.deepseek.com', default: true, color: 'bg-blue-50', icon: 'D' },
  { id: 2, name: 'gemini-2.0-flash', provider: 'Google', url: 'generativelanguage.googleapis.com', default: false, color: 'bg-yellow-50', icon: 'G' },
  { id: 3, name: 'gpt-4o-mini', provider: 'OpenAI', url: 'api.openai.com', default: false, color: 'bg-green-50', icon: 'O' },
];

export default function Settings() {
  const [models, setModels] = useState(initialModels);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ name: '', provider: '', url: '', key: '', temp: '0.3', maxTokens: '4096' });

  const providers = [
    { value: 'openai', label: 'OpenAI', url: 'https://api.openai.com' },
    { value: 'deepseek', label: 'DeepSeek', url: 'https://api.deepseek.com' },
    { value: 'qwen', label: '通义千问 (Qwen)', url: 'https://dashscope.aliyuncs.com/compatible-mode' },
    { value: 'kimi', label: 'Kimi (Moonshot)', url: 'https://api.moonshot.cn' },
    { value: 'claude', label: 'Claude (Anthropic)', url: 'https://api.anthropic.com' },
    { value: 'gemini', label: 'Gemini (Google)', url: 'https://generativelanguage.googleapis.com' },
  ];

  const addModel = () => {
    const p = providers.find(p => p.value === form.provider);
    setModels([...models, { id: Date.now(), name: form.name, provider: p?.label || form.provider, url: form.url, default: false, color: 'bg-purple-50', icon: form.name.charAt(0).toUpperCase() }]);
    setShowForm(false);
    setForm({ name: '', provider: '', url: '', key: '', temp: '0.3', maxTokens: '4096' });
  };

  const removeModel = (id) => setModels(models.filter(m => m.id !== id));

  return (
    <div>
      <h2 className="text-lg font-semibold mb-1">⚙️ 系统设置</h2>
      <p className="text-sm text-gray-500 mb-6">管理多个模型连接、配置系统参数和存储后端。</p>

      {/* Models */}
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
                <div className="text-sm font-medium">{m.name} {m.default && <span className="text-xs bg-green-100 text-green-700 px-1.5 py-0.5 rounded">默认</span>}</div>
                <div className="text-xs text-gray-400">{m.provider} · {m.url}</div>
              </div>
              <span className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">● 可用</span>
              <button className="text-xs text-gray-400 hover:text-gray-700">✏️</button>
              <button className="text-xs text-red-400 hover:text-red-600" onClick={() => removeModel(m.id)}>🗑️</button>
            </div>
          ))}
        </div>

        {showForm && (
          <div className="mt-4 p-4 bg-gray-50 rounded-lg space-y-3">
            <h4 className="text-sm font-medium">添加模型</h4>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-gray-600 block mb-1">Provider</label>
                <select className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.provider}
                  onChange={e => { const p = providers.find(p => p.value === e.target.value); setForm({ ...form, provider: e.target.value, url: p?.url || '' }); }}>
                  <option value="">— 选择 —</option>
                  {providers.map(p => <option key={p.value} value={p.value}>{p.label}</option>)}
                </select>
              </div>
              <div>
                <label className="text-xs text-gray-600 block mb-1">模型名称</label>
                <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.name} onChange={e => setForm({ ...form, name: e.target.value })} placeholder="deepseek-chat" />
              </div>
            </div>
            <div>
              <label className="text-xs text-gray-600 block mb-1">API 地址</label>
              <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.url} onChange={e => setForm({ ...form, url: e.target.value })} />
            </div>
            <div>
              <label className="text-xs text-gray-600 block mb-1">API Key</label>
              <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" type="password" placeholder="sk-..." />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-gray-600 block mb-1">Temperature</label>
                <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.temp} onChange={e => setForm({ ...form, temp: e.target.value })} />
              </div>
              <div>
                <label className="text-xs text-gray-600 block mb-1">Max Tokens</label>
                <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" value={form.maxTokens} onChange={e => setForm({ ...form, maxTokens: e.target.value })} />
              </div>
            </div>
            <div className="flex gap-2">
              <button className="bg-blue-800 text-white px-4 py-1.5 rounded-lg text-xs" onClick={addModel}>💾 保存</button>
              <button className="border px-4 py-1.5 rounded-lg text-xs" onClick={() => setShowForm(false)}>取消</button>
            </div>
          </div>
        )}
      </div>

      {/* Memory */}
      <div className="bg-white rounded-xl border p-5 mb-6">
        <h3 className="font-semibold mb-4">🧠 记忆配置</h3>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-xs text-gray-600 block mb-1">存储后端</label>
            <select className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none"><option>嵌入式文件存储</option><option>PostgreSQL + pgvector</option></select>
          </div>
          <div>
            <label className="text-xs text-gray-600 block mb-1">Top-K 检索</label>
            <input className="w-full border rounded-lg px-3 py-1.5 text-sm outline-none" defaultValue="5" />
          </div>
        </div>
        <button className="mt-4 bg-blue-800 text-white px-4 py-1.5 rounded-lg text-xs">💾 保存</button>
      </div>

      {/* System */}
      <div className="bg-white rounded-xl border p-5">
        <h3 className="font-semibold mb-4">🔧 系统操作</h3>
        <div className="flex gap-2">
          <button className="border px-4 py-2 rounded-lg text-xs">🔄 重载配置</button>
          <button className="border px-4 py-2 rounded-lg text-xs">📥 导出数据</button>
          <button className="border border-red-200 text-red-600 px-4 py-2 rounded-lg text-xs">🗑️ 清除缓存</button>
        </div>
        <p className="text-xs text-gray-400 mt-4">OPC-Agent v0.1.0 · 3 个模型已注册</p>
      </div>
    </div>
  );
}
