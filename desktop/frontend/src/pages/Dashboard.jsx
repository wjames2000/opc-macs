import React, { useState, useEffect } from 'react';

const agentIcons = {
  copywriter: '✍️',
  email_sorter: '📧',
  xhs_poster: '📕',
  competitive_analysis: '📊',
  meeting_minutes: '📋',
};

const agentLabels = {
  copywriter: '文案生成',
  email_sorter: '邮件分类',
  xhs_poster: '小红书',
  competitive_analysis: '竞品分析',
  meeting_minutes: '会议纪要',
};

export default function Dashboard({ onNavigate }) {
  const [stats, setStats] = useState(null);

  useEffect(() => {
    if (window.go?.main?.App?.GetDashboardStats) {
      window.go.main.App.GetDashboardStats().then(setStats).catch(console.warn);
    }
  }, []);

  const cards = [
    { icon: '🤖', label: '活跃 Agent', value: stats?.agent_count ?? '-', sub: '已注册插件', color: 'blue' },
    { icon: '📞', label: '今日调用', value: stats?.today_calls ?? '-', sub: `总计 ${stats?.total_calls ?? 0} 次`, color: 'yellow' },
    { icon: '🔤', label: '今日 Token', value: stats?.today_tokens ? stats.today_tokens.toLocaleString() : '-', sub: '调用消耗', color: 'green' },
    { icon: '💰', label: '总费用', value: stats?.total_cost ?? '-', sub: '按量计费', color: 'purple' },
  ];

  return (
    <div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {cards.map((c, i) => (
          <div key={i} className="bg-white rounded-xl border p-4">
            <div className="text-2xl mb-1">{c.icon}</div>
            <div className="text-xl font-bold">{c.value}</div>
            <div className="text-sm text-gray-500">{c.label}</div>
            <div className="text-xs text-gray-400 mt-0.5">{c.sub}</div>
          </div>
        ))}
      </div>

      <div className="flex gap-3 mb-6">
        <button className="bg-[#1E3A5F] text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-[#2A4A7F]"
          onClick={() => onNavigate('chat')}>💬 新对话</button>
        <button className="border px-5 py-2 rounded-lg text-sm hover:bg-gray-50"
          onClick={() => onNavigate('agents')}>🤖 管理 Agent</button>
        <button className="border px-5 py-2 rounded-lg text-sm hover:bg-gray-50"
          onClick={() => onNavigate('usage')}>📈 查看用量</button>
      </div>

      <div className="bg-white rounded-xl border p-5">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold">最近活动</h3>
        </div>
        {(!stats?.recent_activity || stats.recent_activity.length === 0) ? (
          <p className="text-sm text-gray-400 text-center py-8">暂无活动记录，开始使用 Agent 吧</p>
        ) : (
          stats.recent_activity.slice().reverse().map((item, i) => (
            <div key={i} className="flex items-center gap-3 py-3 border-b last:border-0">
              <div className="w-9 h-9 rounded-lg bg-blue-50 flex items-center justify-center text-sm">
                {agentIcons[item.agent] || '🤖'}
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium truncate">{agentLabels[item.agent] || item.agent}</p>
                <p className="text-xs text-gray-400 truncate">{item.input}</p>
              </div>
              <div className="text-right flex-shrink-0">
                <p className="text-xs text-gray-400">{item.time}</p>
                <p className="text-xs text-gray-400">{item.tokens} tokens</p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
