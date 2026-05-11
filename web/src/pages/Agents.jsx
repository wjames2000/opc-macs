import React from 'react';

const agents = [
  { name: 'copywriter', summary: '营销文案生成', model: 'deepseek-v4-flash', hitl: true, calls: 12, tokens: '3,450', color: 'border-l-blue-500' },
  { name: 'email_sorter', summary: '邮件分类与回复', model: 'gemini-2.0-flash', hitl: true, calls: 8, tokens: '1,200', color: 'border-l-yellow-500' },
  { name: 'xhs_poster', summary: '小红书种草笔记', model: 'deepseek-v4-flash', hitl: true, calls: 5, tokens: '2,100', color: 'border-l-green-500' },
  { name: 'competitive_analysis', summary: '竞品 SWOT 分析', model: 'deepseek-v4-flash', hitl: false, calls: 3, tokens: '4,500', color: 'border-l-purple-500' },
  { name: 'meeting_minutes', summary: '会议纪要整理', model: 'deepseek-v4-flash', hitl: false, calls: 1, tokens: '2,800', color: 'border-l-red-500' },
];

export default function Agents() {
  return (
    <div>
      <div className="flex justify-between items-center mb-5">
        <p className="text-sm text-gray-500">管理已安装的 Agent，配置运行参数</p>
        <button className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm">+ 添加 Agent</button>
      </div>
      <div className="grid md:grid-cols-2 gap-4">
        {agents.map((agent, i) => (
          <div key={i} className={`bg-white rounded-xl border border-l-4 ${agent.color} p-5`}>
            <div className="flex items-center gap-3 mb-3">
              <span className="w-2.5 h-2.5 rounded-full bg-green-500" />
              <div>
                <h3 className="font-semibold">{agent.name}</h3>
                <p className="text-sm text-gray-500">{agent.summary}</p>
              </div>
              <span className="ml-auto text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">● 在线</span>
            </div>
            <div className="flex gap-4 text-xs text-gray-500 mb-3">
              <span>🤖 {agent.model}</span>
              <span>{agent.hitl ? '⚡ 需 HITL' : '✅ 自动'}</span>
              <span>📊 今日 {agent.calls} 次</span>
            </div>
            <p className="text-xs text-gray-400 mb-3">Token: {agent.tokens}</p>
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
