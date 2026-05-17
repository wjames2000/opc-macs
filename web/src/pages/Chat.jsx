import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';

const initialMessages = [
  { role: 'agent', text: '您好！我是文案生成助手。请告诉我您需要推广的产品或服务信息。', time: '10:28 AM', tokens: 120 },
  { role: 'user', text: '杭州可人酒店，高端商务度假型酒店，位于西湖景区核心位置。帮我写推广文案。', time: '10:30 AM' },
  { role: 'agent', text: '', time: '10:32 AM', tokens: 1250, structured: true },
];

export default function Chat() {
  const [agents, setAgents] = useState(['copywriter', 'email_sorter', 'xhs_poster', 'competitive_analysis', 'meeting_minutes']);
  const [selectedAgent, setSelectedAgent] = useState('copywriter');
  const [messages, setMessages] = useState(initialMessages);
  const [input, setInput] = useState('');
  const [sending, setSending] = useState(false);

  useEffect(() => {
    api.listAgents().then(data => {
      if (Array.isArray(data) && data.length > 0) {
        setAgents(data.map(a => a.name));
        setSelectedAgent(data[0].name);
      }
    }).catch(() => {});
  }, []);

  const handleSend = async () => {
    if (!input.trim() || sending) return;
    const userMsg = { role: 'user', text: input, time: new Date().toLocaleTimeString() };
    setMessages(prev => [...prev, userMsg]);
    setInput('');
    setSending(true);

    try {
      const result = await api.executeAgent({ name: selectedAgent, input });
      setMessages(prev => [...prev, { role: 'agent', text: result?.result || JSON.stringify(result), time: new Date().toLocaleTimeString(), tokens: result?.token_usage?.total_tokens || 0 }]);
    } catch (err) {
      setMessages(prev => [...prev, { role: 'agent', text: '执行出错：' + err.message, time: new Date().toLocaleTimeString() }]);
    }
    setSending(false);
  };

  return (
    <div className="flex h-[calc(100vh-9rem)] -m-6">
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
              <div className={`max-w-lg ${msg.role === 'user' ? 'bg-blue-800 text-white' : 'bg-gray-50'} rounded-xl px-4 py-3`}>
                {msg.structured ? (
                  <div><p className="text-sm font-medium mb-1">📋 文案生成结果</p><p className="text-sm text-gray-600">[标题] 栖居西湖畔——杭州可人酒店推广文案</p><p className="text-sm text-gray-600 mt-1">[正文] 西湖美景，近在咫尺。可人酒店坐落于西湖景区核心...</p></div>
                ) : (
                  <p className="text-sm">{msg.text}</p>
                )}
                <div className={`text-xs mt-1 ${msg.role === 'user' ? 'text-blue-200' : 'text-gray-400'}`}>
                  {msg.time}{msg.tokens ? ` · ${msg.tokens} tokens` : ''}
                </div>
              </div>
            </div>
          ))}
          {sending && <div className="text-center text-sm text-gray-400">Agent 思考中...</div>}
        </div>

        <div className="border-t p-3 flex gap-2">
          <input className="flex-1 px-4 py-2 border rounded-lg text-sm outline-none focus:border-blue-500" placeholder="输入任务描述..." value={input} onChange={e => setInput(e.target.value)} onKeyDown={e => e.key === 'Enter' && handleSend()} />
          <button className="bg-blue-800 text-white px-5 py-2 rounded-lg text-sm font-medium disabled:opacity-50" disabled={sending || !input.trim()} onClick={handleSend}>{sending ? '发送中...' : '发送'}</button>
        </div>
      </div>
    </div>
  );
}
