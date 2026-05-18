import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';
import StatCard from '../components/StatCard';

export default function Dashboard() {
  const [stats, setStats] = useState(null);

  useEffect(() => {
    api.usageSummary(localStorage.getItem('tenant_id') || 'demo')
      .then(setStats)
      .catch(() => {});
  }, []);

  const cards = [
    { icon: '🤖', label: '活跃 Agent', value: stats?.active_agents?.toString() ?? '5', sub: stats ? `上次活跃: ${stats.last_active ?? '今天'}` : '加载中...', color: 'blue' },
    { icon: '💰', label: '今日 Token', value: stats ? (stats.daily_tokens ?? stats.total_tokens ?? 0).toLocaleString() : '12,450', sub: stats ? `总计: ${(stats.total_tokens ?? 0).toLocaleString()}` : '加载中...', color: 'yellow' },
    { icon: '📊', label: '本月费用', value: stats?.monthly_cost ? `$${stats.monthly_cost.toFixed(2)}` : '$0.00', sub: stats?.budget ? `预算 ¥${stats.budget}` : '预算内', color: 'green' },
    { icon: '📋', label: '任务总数', value: stats?.total_tasks?.toString() ?? '0', sub: stats ? `今日 ${stats.today_tasks ?? 0}` : '加载中...', color: 'purple' },
  ];

  return (
    <div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {cards.map((c, i) => <StatCard key={i} {...c} />)}
      </div>

      <div className="flex gap-3 mb-6">
        <button className="bg-blue-800 text-white px-5 py-2 rounded-lg text-sm font-medium hover:bg-blue-900">💬 新对话</button>
        <button className="border px-5 py-2 rounded-lg text-sm hover:bg-gray-50">🤖 管理 Agent</button>
        <button className="border px-5 py-2 rounded-lg text-sm hover:bg-gray-50">📈 查看用量</button>
      </div>

      <div className="bg-white rounded-xl border p-5">
        <div className="flex justify-between items-center mb-4">
          <h3 className="font-semibold">最近活动</h3>
          <span className="text-sm text-gray-400">{stats?.last_updated ? `更新于 ${stats.last_updated}` : ''}</span>
        </div>
        {[
          { icon: '💬', agent: 'copywriter', text: '文案生成', meta: stats ? `${stats.total_tokens ?? 0} tokens` : '加载中...', time: '最近', color: 'bg-blue-50' },
          { icon: '📊', agent: 'competitive', text: '趋势分析', meta: stats ? `${stats.active_agents ?? 0} active` : '加载中...', time: '最近', color: 'bg-purple-50' },
        ].map((item, i) => (
          <div key={i} className="flex items-center gap-3 py-3 border-b last:border-0">
            <div className={`w-9 h-9 rounded-lg ${item.color} flex items-center justify-center text-sm`}>{item.icon}</div>
            <div className="flex-1">
              <p className="text-sm font-medium">{item.text}</p>
              <p className="text-xs text-gray-400">{item.agent} · {item.meta}</p>
            </div>
            <span className="text-xs text-gray-400">{item.time}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
