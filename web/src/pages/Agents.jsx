import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';

const COLORS = ['border-l-blue-500', 'border-l-yellow-500', 'border-l-green-500', 'border-l-purple-500', 'border-l-red-500', 'border-l-cyan-500', 'border-l-pink-500', 'border-l-indigo-500', 'border-l-teal-500', 'border-l-orange-500'];

export default function Agents() {
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.listAgents().then(data => {
      if (Array.isArray(data)) setAgents(data);
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="p-8 text-gray-500">加载中...</div>;

  return (
    <div>
      <div className="flex justify-between items-center mb-5">
        <p className="text-sm text-gray-500">管理已安装的 Agent，配置运行参数</p>
        <button className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm">+ 添加 Agent</button>
      </div>
      <div className="grid md:grid-cols-2 gap-4">
        {agents.map((agent, i) => (
          <div key={agent.name || i} className={`bg-white rounded-xl border border-l-4 ${COLORS[i % COLORS.length]} p-5`}>
            <div className="flex items-center gap-3 mb-3">
              <span className="w-2.5 h-2.5 rounded-full bg-green-500" />
              <div>
                <h3 className="font-semibold">{agent.name}</h3>
                <p className="text-sm text-gray-500">{agent.summary}</p>
              </div>
              <span className="ml-auto text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">● 在线</span>
            </div>
            <div className="flex gap-4 text-xs text-gray-500 mb-3">
              <span>🤖 {agent.model_name || 'default'}</span>
              <span>{agent.requires_hitl ? '⚡ 需 HITL' : '✅ 自动'}</span>
            </div>
            <div className="flex gap-2">
              <button className="text-xs border px-3 py-1.5 rounded hover:bg-gray-50">⚙️ 配置</button>
              <button className="text-xs border px-3 py-1.5 rounded hover:bg-gray-50">📊 详情</button>
              <button className="text-xs border border-red-200 text-red-600 px-3 py-1.5 rounded hover:bg-red-50">🔌 禁用</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
