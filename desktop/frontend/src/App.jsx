import React, { useState, useEffect, useRef } from 'react';

const agents = [
  { value: 'copywriter', label: '✍️ 文案生成', model: 'deepseek-v4-flash' },
  { value: 'email_sorter', label: '📧 邮件分类', model: 'gemini-2.0-flash' },
  { value: 'xhs_poster', label: '📕 小红书', model: 'deepseek-v4-flash' },
  { value: 'competitive_analysis', label: '📊 竞品分析', model: 'deepseek-v4-flash' },
  { value: 'meeting_minutes', label: '📋 会议纪要', model: 'deepseek-v4-flash' },
];

export default function App() {
  const [agent, setAgent] = useState('copywriter');
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [version, setVersion] = useState({ version: '0.1.0', agents: '0' });
  const messagesEnd = useRef(null);

  useEffect(() => {
    // Get version from Go backend
    if (window.go?.main?.App?.GetVersion) {
      window.go.main.App.GetVersion().then(setVersion).catch(() => {});
    }
  }, []);

  useEffect(() => {
    messagesEnd.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const execute = async () => {
    if (!input.trim() || loading) return;

    setLoading(true);
    setMessages(prev => [...prev, { role: 'user', text: input, agent }]);
    const currentInput = input;
    setInput('');

    try {
      // Call Wails Go binding
      let result;
      if (window.go?.main?.App?.Execute) {
        result = await window.go.main.App.Execute({ agent, input: currentInput });
      } else {
        // Fallback for dev mode (no Wails runtime)
        result = { success: false, error: 'Wails runtime not connected. Run with wails dev.' };
      }

      if (result.success) {
        const output = typeof result.data === 'string'
          ? result.data
          : JSON.stringify(result.data, null, 2);
        setMessages(prev => [...prev, {
          role: 'agent',
          text: output,
          agent,
          tokens: `${result.tokens?.input || 0}/${result.tokens?.output || 0}`,
        }]);
      } else {
        setMessages(prev => [...prev, {
          role: 'error',
          text: result.error || '执行失败',
          agent,
        }]);
      }
    } catch (err) {
      setMessages(prev => [...prev, { role: 'error', text: err.message, agent }]);
    }

    setLoading(false);
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Title bar area (macOS inset) */}
      <header className="bg-[#1E3A5F] text-white px-5 pt-8 pb-3 flex items-center justify-between select-none">
        <div>
          <h1 className="font-bold text-base">🤖 OPC-Agent</h1>
          <p className="text-xs text-blue-200">v{version.version} · {version.agents} Agents</p>
        </div>
        <select
          className="bg-white/10 text-white border border-white/20 rounded-lg px-3 py-1.5 text-xs outline-none"
          value={agent} onChange={e => setAgent(e.target.value)}
        >
          {agents.map(a => <option key={a.value} value={a.value}>{a.label}</option>)}
        </select>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-4 space-y-3">
        {messages.length === 0 && (
          <div className="text-center text-gray-400 mt-20">
            <div className="text-3xl mb-3">🤖</div>
            <p className="font-medium text-gray-600">OPC-Agent 桌面版</p>
            <p className="text-xs mt-1">选择 Agent 并输入任务开始使用</p>
          </div>
        )}
        {messages.map((msg, i) => (
          <div key={i} className={`flex gap-2.5 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}>
            <div className={`w-7 h-7 rounded-full flex items-center justify-center text-xs flex-shrink-0 ${
              msg.role === 'user' ? 'bg-[#1E3A5F] text-white' : 
              msg.role === 'error' ? 'bg-red-100 text-red-500' : 'bg-blue-50 text-blue-600'
            }`}>
              {msg.role === 'user' ? 'A' : msg.role === 'error' ? '⚠' : '🤖'}
            </div>
            <div className={`max-w-[75%] ${msg.role === 'user' ? 'text-right' : ''}`}>
              <div className={`rounded-xl px-3.5 py-2.5 text-sm leading-relaxed ${
                msg.role === 'user' ? 'bg-[#1E3A5F] text-white' :
                msg.role === 'error' ? 'bg-red-50 text-red-600 border border-red-100' :
                'bg-white border shadow-sm'
              }`}>
                {msg.text}
              </div>
              <div className="text-xs text-gray-400 mt-1">
                {msg.agent} {msg.tokens ? `· ${msg.tokens} tokens` : ''}
              </div>
            </div>
          </div>
        ))}
        {loading && (
          <div className="flex gap-2.5">
            <div className="w-7 h-7 rounded-full bg-blue-50 flex items-center justify-center text-xs">🤖</div>
            <div className="bg-white border rounded-xl px-4 py-3 text-sm text-gray-400">
              <span className="animate-pulse">⏳ 思考中...</span>
            </div>
          </div>
        )}
        <div ref={messagesEnd} />
      </div>

      {/* Input */}
      <div className="border-t bg-white px-4 py-3">
        <div className="flex gap-2">
          <textarea
            className="flex-1 border rounded-xl px-4 py-2.5 text-sm outline-none resize-none focus:border-blue-400"
            rows={1} value={input} onChange={e => setInput(e.target.value)}
            placeholder={`输入任务，交给 ${agents.find(a => a.value === agent)?.label || ''}...`}
            onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); execute(); } }}
          />
          <button
            className="bg-[#1E3A5F] text-white px-4 rounded-xl text-lg disabled:opacity-40 hover:bg-[#2D5A8E] transition-colors"
            onClick={execute} disabled={loading || !input.trim()}
          >➤</button>
        </div>
        <p className="text-xs text-gray-400 mt-1.5">Enter 发送 · Shift+Enter 换行</p>
      </div>
    </div>
  );
}
