import React, { useState, useRef, useEffect } from 'react';

const agents = [
  { value: 'copywriter', label: '✍️ 文案生成' },
  { value: 'email_sorter', label: '📧 邮件分类' },
  { value: 'xhs_poster', label: '📕 小红书' },
  { value: 'competitive_analysis', label: '📊 竞品分析' },
  { value: 'meeting_minutes', label: '📋 会议纪要' },
];

function formatOutput(data, agentType) {
  // Handle string data that might be JSON
  if (typeof data === 'string') {
    try { data = JSON.parse(data); } catch (e) { return data; }
  }
  // Handle null/undefined
  if (!data || typeof data !== 'object') return String(data || '');

  const m = data;
  const s = (k) => typeof m[k] === 'string' ? m[k] : '';
  const list = (k) => Array.isArray(m[k]) ? m[k].map(x => typeof x === 'object' ? JSON.stringify(x) : String(x)) : [];

  switch (agentType) {
    case 'copywriter':
      return `📝 短文案\n${s('short_copy')}\n\n📄 长文案\n${s('long_copy')}\n\n📱 社交文案\n${s('social_copy')}\n\n🎨 风格：${s('style')}`;

    case 'email_sorter': {
      const icon = s('category') === '投诉' ? '⚠️' : s('category') === '合作' ? '🤝' : '📧';
      return `${icon} 分类：${s('category')}${s('urgency') === '高' ? ' 🔴 紧急' : ''}\n\n理由：${s('reason')}\n\n💬 回复建议\n${s('reply_suggestion')}`;
    }

    case 'xhs_poster':
      return `📕 ${s('title')}\n\n${s('body')}\n\n标签：${list('hashtags').join('  ')}\n配图：${list('image_suggestions').join('、')}\n风格：${s('style')}`;

    case 'competitive_analysis': {
      let ca = '';
      ca += `📊 竞品：${s('competitor')}\n\n定位：${s('market_position')}\n`;
      ca += `\n✅ 优势\n${list('strengths').map(x => `  • ${x}`).join('\n')}`;
      ca += `\n\n❌ 劣势\n${list('weaknesses').map(x => `  • ${x}`).join('\n')}`;
      ca += `\n\n📈 机会\n${list('opportunities').map(x => `  • ${x}`).join('\n')}`;
      ca += `\n\n⚠️ 威胁\n${list('threats').map(x => `  • ${x}`).join('\n')}`;
      ca += `\n\n💡 差异化：${s('differentiation')}`;
      ca += `\n\n风险：${s('risk_level') === '高' ? '🔴' : s('risk_level') === '中' ? '🟡' : '🟢'} ${s('risk_level')}`;
      ca += `\n\n📋 总结\n${s('summary')}`;
      return ca;
    }

    case 'meeting_minutes': {
      let mm = '';
      mm += `📅 ${s('title')}\n`;
      if (s('time')) mm += `\n时间：${s('time')}`;
      if (list('participants').length) mm += `\n参会：${list('participants').join('、')}`;
      mm += `\n\n📋 议程\n${list('agenda').map(x => `  • ${x}`).join('\n')}`;
      mm += `\n\n✅ 决策\n${list('decisions').map(x => `  • ${x}`).join('\n')}`;
      mm += `\n\n🔧 待办\n${Array.isArray(m['action_items']) ? m['action_items'].map(x => `  ☐ ${x.task || ''}（${x.owner || ''}，${x.deadline || ''}）`).join('\n') : ''}`;
      if (s('next_steps')) mm += `\n\n📌 下一步\n${s('next_steps')}`;
      return mm;
    }

    default:
      return Object.entries(m).map(([k, v]) => {
        const val = Array.isArray(v) ? v.join(', ') : String(v);
        return `${k}：${val}`;
      }).join('\n');
  }
}

export default function Chat() {
  const [agent, setAgent] = useState('copywriter');
  const [input, setInput] = useState('');
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [conversations, setConversations] = useState([]);
  const [activeConv, setActiveConv] = useState(null);
  const messagesEnd = useRef(null);

  useEffect(() => {
    messagesEnd.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const execute = async () => {
    if (!input.trim() || loading) return;
    setLoading(true);

    const text = input;
    const currentAgent = agent;
    setMessages(prev => [...prev, { role: 'user', text, agent: currentAgent }]);
    setInput('');

    try {
      let result;
      if (window.go?.main?.App?.Execute) {
        result = await window.go.main.App.Execute({ agent: currentAgent, input: text });
      } else {
        result = { success: false, error: 'Wails runtime not connected' };
      }

      if (result.success) {
        const formatted = formatOutput(result.data, currentAgent);
        setMessages(prev => [...prev, {
          role: 'agent', text: formatted, agent: currentAgent,
          tokens: `${result.tokens?.input || 0}/${result.tokens?.output || 0}`,
        }]);
        // Add to conversation list
        const title = text.length > 30 ? text.slice(0, 30) + '...' : text;
        const now = new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
        setConversations(prev => {
          const existing = activeConv ? prev.filter(c => c.id !== activeConv) : prev;
          return [{ id: Date.now().toString(), title, preview: formatted.slice(0, 40), time: now }, ...existing];
        });
        if (!activeConv) setActiveConv(Date.now().toString());
      } else {
        setMessages(prev => [...prev, { role: 'error', text: result.error || '执行失败', agent: currentAgent }]);
      }
    } catch (err) {
      setMessages(prev => [...prev, { role: 'error', text: err.message, agent: currentAgent }]);
    }
    setLoading(false);
  };

  const newConversation = () => {
    setMessages([]);
    setActiveConv(null);
  };

  return (
    <div className="flex h-[calc(100vh-9rem)] -m-6">
      {/* Conversation sidebar */}
      <div className="w-64 border-r bg-white flex flex-col flex-shrink-0">
        <div className="p-3 border-b">
          <button className="w-full bg-[#1E3A5F] text-white py-2 rounded-lg text-sm font-medium mb-2"
            onClick={newConversation}>＋ 新对话</button>
        </div>
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {conversations.length === 0 ? (
            <p className="text-xs text-gray-400 text-center pt-6">暂无对话历史</p>
          ) : (
            conversations.map((c, i) => (
              <div key={c.id}
                className={`p-3 rounded-lg cursor-pointer text-sm ${c.id === activeConv ? 'bg-gray-100' : 'hover:bg-gray-50'}`}
                onClick={() => setActiveConv(c.id)}>
                <div className="font-medium truncate">{c.title}</div>
                <div className="text-xs text-gray-400 mt-1 truncate">{c.preview}</div>
                <div className="text-xs text-gray-300 mt-0.5">{c.time}</div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Chat area */}
      <div className="flex-1 flex flex-col bg-white">
        <div className="flex items-center gap-3 px-5 py-3 border-b">
          <span className="w-2 h-2 rounded-full bg-green-500" />
          <span className="font-medium text-sm">{activeConv ? '对话中' : '新对话'}</span>
          <div className="ml-auto flex items-center gap-2 text-xs text-gray-400">
            <span>Agent:</span>
            <select className="border rounded px-2 py-1 text-xs outline-none"
              value={agent} onChange={e => setAgent(e.target.value)}>
              {agents.map(a => <option key={a.value} value={a.value}>{a.label}</option>)}
            </select>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          {messages.length === 0 ? (
            <div className="text-center text-gray-400 mt-16">
              <div className="text-3xl mb-3">🤖</div>
              <p className="font-medium text-gray-600">选择 Agent 并输入任务</p>
              <p className="text-xs mt-1">Enter 发送 · Shift+Enter 换行</p>
            </div>
          ) : (
            messages.map((msg, i) => (
              <div key={i} className={`flex gap-3 ${msg.role === 'user' ? 'flex-row-reverse' : ''}`}>
                <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm flex-shrink-0 ${
                  msg.role === 'user' ? 'bg-[#1E3A5F] text-white' :
                  msg.role === 'error' ? 'bg-red-100 text-red-500' : 'bg-blue-50 text-blue-600'
                }`}>
                  {msg.role === 'user' ? 'A' : msg.role === 'error' ? '⚠' : '🤖'}
                </div>
                <div className={`max-w-[75%] ${msg.role === 'user' ? 'text-right' : ''}`}>
                  <div className={`rounded-xl px-3.5 py-2.5 text-sm leading-relaxed whitespace-pre-wrap ${
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
            ))
          )}
          {loading && (
            <div className="flex gap-3">
              <div className="w-8 h-8 rounded-full bg-blue-50 flex items-center justify-center text-sm">🤖</div>
              <div className="bg-white border rounded-xl px-4 py-3 text-sm text-gray-400 animate-pulse">⏳ 思考中...</div>
            </div>
          )}
          <div ref={messagesEnd} />
        </div>

        <div className="border-t bg-white px-4 py-3">
          <div className="flex gap-2">
            <textarea className="flex-1 border rounded-xl px-4 py-2.5 text-sm outline-none resize-none focus:border-blue-400"
              rows={1} value={input} onChange={e => setInput(e.target.value)}
              placeholder={`输入任务，交给 ${agents.find(a => a.value === agent)?.label || ''}...`}
              onKeyDown={e => { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); execute(); } }} />
            <button className="bg-[#1E3A5F] text-white px-4 rounded-xl text-lg disabled:opacity-40"
              onClick={execute} disabled={loading || !input.trim()}>➤</button>
          </div>
        </div>
      </div>
    </div>
  );
}
