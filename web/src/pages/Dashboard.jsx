import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';
import StatCard from '../components/StatCard';

export default function Dashboard() {
  const [stats, setStats] = useState(null);

  useEffect(() => {
    api.usageSummary(localStorage.getItem('tenant_id') || 'demo').then(setStats).catch(() => {});
  }, []);

  const cards = [
    { icon: '🤖', label: '活跃 Agent', value: '5', sub: '+2 本周新增', color: 'blue' },
    { icon: '💰', label: '今日 Token', value: '12,450', sub: '↓ 8% 较昨日', color: 'yellow' },
    { icon: '📊', label: '本月费用', value: '$0.32', sub: '预算内 ✅', color: 'green' },
    { icon: '📋', label: '任务总数', value: '47', sub: '今日 12', color: 'purple' },
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
          <a href="#" className="text-sm text-cyan-600">查看全部 →</a>
        </div>
        {[
          { icon: '💬', agent: 'copywriter', text: '文案生成 · 杭州可人酒店', meta: '1,250 tokens', time: '5 分钟前', color: 'bg-blue-50' },
          { icon: '📧', agent: 'email_sorter', text: '邮件分类 · 客户投诉', meta: '340 tokens', time: '12 分钟前', color: 'bg-green-50' },
          { icon: '📊', agent: 'competitive', text: '竞品分析 · 竞品报告', meta: '4,500 tokens', time: '32 分钟前', color: 'bg-purple-50' },
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
