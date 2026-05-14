import React, { useState, useEffect } from 'react';

const agentColors = {
  copywriter: 'border-l-blue-500',
  email_sorter: 'border-l-yellow-500',
  xhs_poster: 'border-l-green-500',
  competitive_analysis: 'border-l-purple-500',
  meeting_minutes: 'border-l-red-500',
};

const agentIcons = {
  copywriter: '✍️',
  email_sorter: '📧',
  xhs_poster: '📕',
  competitive_analysis: '📊',
  meeting_minutes: '📋',
};

export default function Agents() {
  const [agents, setAgents] = useState([]);

  useEffect(() => {
    if (window.go?.main?.App?.GetAgentList) {
      window.go.main.App.GetAgentList().then(setAgents).catch(console.warn);
    }
  }, []);

  return (
    <div>
      <div className="flex justify-between items-center mb-5">
        <p className="text-sm text-gray-500">已安装 {agents.length} 个 Agent 插件</p>
      </div>
      <div className="grid md:grid-cols-2 gap-4">
        {agents.map((agent, i) => (
          <div key={i} className={`bg-white rounded-xl border border-l-4 ${agentColors[agent.name] || 'border-l-gray-500'} p-5`}>
            <div className="flex items-center gap-3 mb-3">
              <div className="w-10 h-10 rounded-lg bg-gray-50 flex items-center justify-center text-lg">
                {agentIcons[agent.name] || '🤖'}
              </div>
              <div>
                <h3 className="font-semibold">{agent.name}</h3>
                <p className="text-sm text-gray-500">{agent.summary}</p>
              </div>
              <span className="ml-auto text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">● 在线</span>
            </div>
            <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500 mb-3">
              <span>🤖 {agent.model || '默认模型'}</span>
              <span>{agent.requires_hitl ? '⚡ 需 HITL' : '✅ 自动执行'}</span>
              <span>📦 v{agent.version || '1.0.0'}</span>
            </div>
            {agent.tags && <p className="text-xs text-gray-400 mb-3">🏷️ {agent.tags}</p>}
            <div className="flex gap-2">
              <button className="text-xs border px-3 py-1.5 rounded hover:bg-gray-50">⚙️ 配置</button>
              <button className="text-xs border px-3 py-1.5 rounded hover:bg-gray-50">📊 详情</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
