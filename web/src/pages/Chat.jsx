import React, { useState } from 'react';

const agents = ['copywriter', 'email_sorter', 'xhs_poster', 'competitive_analysis', 'meeting_minutes'];

const initialMessages = [
  { role: 'agent', text: '您好！我是文案生成助手。请告诉我您需要推广的产品或服务信息。', time: '10:28 AM', tokens: 120 },
  { role: 'user', text: '杭州可人酒店，高端商务度假型酒店，位于西湖景区核心位置。帮我写推广文案。', time: '10:30 AM' },
  { role: 'agent', text: '', time: '10:32 AM', tokens: 1250, structured: true },
];

export default function Chat() {
  const [selectedAgent, setSelectedAgent] = useState('copywriter');
  const [messages] = useState(initialMessages);

  return (
    <div className="flex h-[calc(100vh-9rem)] -m-6">
      {/* Conversation sidebar */}
      <div className="w-72 border-r bg-white flex flex-col flex-shrink-0">
        <div className="p-3 border-b">
          <button className="w-full bg-blue-800 text-white py-2 rounded-lg text-sm font-medium mb-2">＋ 新对话</button>
          <input className="w-full px-3 py-1.5 border rounded-lg text-sm outline-none" placeholder="搜索..." />
        </div>
        <div className="flex-1 overflow-y-auto p-2">
          {['杭州可人酒店推广', '客户投诉处理', '竞品分析报告'].map((title, i) => (
            <div key={i} className={`p-3 rounded-lg cursor-pointer text-sm ${i === 0 ? 'bg-gray-100' : 'hover:bg-gray-50'}`}>
              <div className="font-medium truncate">{title}</div>
              <div className="text-xs text-gray-400 mt-1 truncate">消息预览...</div>
            </div>
          ))}
        </div>
      </div>

      {/* Chat area */}
      <div className="flex-1 flex flex-col bg-white">
        <div className="flex items-center gap-3 px-5 py-3 border-b">
          <span className="w-2 h-2 rounded-full bg-green-500" />
          <span className="font-medium text-sm">杭州可人酒店推广</span>
          <span className="text-xs bg-green-50 text-green-700 px-2 py-0.5 rounded">● 进行中</span>
          <div className="ml-auto flex items-center gap-2 text-xs text-gray-400">
            <span>Agent:</span>
            <select className="border rounded px-2 py-1 text-xs" value={selectedAgent} onChange={e => setSelectedAgent(e.target.value)}>
              {agents.map(a => <option key={a}>{a}</option>)}
            </select>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          {messages.map((msg, i) => (
            <div key={i} className={`flex gap-3 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm flex-shrink-0 ${msg.role === 'agent' ? 'bg-blue-50' : 'bg-blue-800 text-white'}`}>
                {msg.role === 'agent' ? '🤖' : 'A'}
              </div>
              <div className={`max-w-[70%] ${msg.role === 'user' ? 'text-right' : ''}`}>
                {msg.structured ? (
                  <div className="bg-white border rounded-xl p-4 space-y-3">
                    {['短文案', '长文案'].map((label, j) => (
                      <div key={j}><p className="text-xs text-gray-500 mb-1">{j === 0 ? '📝' : '📄'} {label}</p>
                      <div className="bg-gray-50 rounded-lg p-3 text-sm">{label === '短文案' ? '西湖边的隐世桃源——杭州可人酒店' : '在西湖畔醒来，推窗即景。杭州可人酒店...'}</div></div>
                    ))}
                    <div className="flex gap-2"><button className="text-xs bg-green-600 text-white px-3 py-1.5 rounded">✅ 确认</button><button className="text-xs border px-3 py-1.5 rounded">🔄 重试</button></div>
                  </div>
                ) : (
                  <div className={`rounded-xl px-4 py-2.5 text-sm leading-relaxed ${msg.role === 'agent' ? 'bg-white border' : 'bg-blue-800 text-white'}`}>{msg.text}</div>
                )}
                <div className="text-xs text-gray-400 mt-1">{msg.time}{msg.tokens ? ` · ${msg.tokens} tokens` : ''}</div>
              </div>
            </div>
          ))}
        </div>

        <div className="border-t p-4">
          <div className="flex gap-2 items-end">
            <textarea className="flex-1 border rounded-lg px-4 py-2.5 text-sm outline-none focus:ring-2 focus:ring-blue-500 resize-none" rows={1} placeholder="输入 @agent 指定 Agent..." />
            <button className="bg-blue-800 text-white px-4 py-2.5 rounded-lg text-sm">➤</button>
          </div>
          <p className="text-xs text-gray-400 mt-2">@copywriter · Enter 发送</p>
        </div>
      </div>
    </div>
  );
}
