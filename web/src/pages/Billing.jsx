import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';

const PLANS = [
  { name: 'Free', price: '0', desc: '适合个人试用', features: ['10 万 Token/月', '3 个 Agent', '1 个用户', '2 个工作流'], limit: false, color: 'gray', popular: false },
  { name: 'Pro', price: '99', desc: '适合小型团队', features: ['100 万 Token/月', '10 个 Agent', '5 个用户', '20 个工作流', '超额 ¥0.008/千 tokens'], limit: true, color: 'cyan', popular: true },
  { name: 'Enterprise', price: '499', desc: '适合企业', features: ['1000 万 Token/月', '不限 Agent', '不限用户', '不限工作流', '超额 ¥0.005/千 tokens', '专属技术支持'], limit: false, color: 'blue', popular: false },
];

export default function Billing() {
  const [usage, setUsage] = useState(null);

  useEffect(() => {
    api.usageSummary(localStorage.getItem('tenant_id') || 'demo')
      .then(setUsage)
      .catch(() => {}); // falls back to null
  }, []);

  const used = usage?.total_tokens ?? 845200;
  const limit = 1000000;
  const pct = Math.min(100, (used / limit * 100)).toFixed(1);
  const currentPlan = 'Pro';

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold">💎 套餐与账单</h2>
          <p className="text-sm text-gray-500">当前套餐: {currentPlan}</p>
        </div>
      </div>

      {/* Usage gauge */}
      <div className="bg-white rounded-xl border p-5 mb-6">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          {[
            { label: '本月已用 Token', value: used.toLocaleString(), sub: usage ? `↑ ${usage.monthly_change ?? 0}% 较上月` : '加载中...' },
            { label: '套餐限额', value: limit.toLocaleString(), sub: `${currentPlan} 套餐` },
            { label: '已用比例', value: `${pct}%`, sub: `剩余 ${(limit - used).toLocaleString()}` },
            { label: '预估费用', value: '¥99', sub: '套餐费内' },
          ].map((s, i) => (
            <div key={i}>
              <div className="text-xl font-bold font-mono">{s.value}</div>
              <div className="text-xs text-gray-500 mt-0.5">{s.label}</div>
              <div className="text-xs text-gray-400 mt-0.5">{s.sub}</div>
            </div>
          ))}
        </div>
        <div className="w-full h-2 bg-gray-100 rounded-full overflow-hidden">
          <div className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-amber-500" style={{ width: `${pct}%` }} />
        </div>
      </div>

      {/* Plans */}
      <div className="grid md:grid-cols-3 gap-4 mb-6">
        {PLANS.map((plan, i) => (
          <div key={i} className={`bg-white rounded-xl border p-6 ${plan.popular ? 'border-cyan-500 ring-1 ring-cyan-500 relative' : ''}`}>
            {plan.popular && <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-cyan-500 text-white text-xs px-3 py-0.5 rounded-full font-medium">最受欢迎</div>}
            <div className="text-xs font-semibold text-gray-500 uppercase tracking-wider">{plan.name}</div>
            <div className="mt-3 mb-1"><span className="text-3xl font-bold">¥{plan.price}</span>{plan.price !== '0' && <span className="text-gray-400 text-sm">/月</span>}</div>
            <div className="text-xs text-gray-400 mb-4">{plan.desc}</div>
            <ul className="space-y-2 mb-6">
              {plan.features.map((f, j) => <li key={j} className="flex items-start gap-2 text-sm"><span className="text-green-500 mt-0.5">✓</span>{f}</li>)}
            </ul>
            <button className={`w-full py-2 rounded-lg text-sm font-medium border ${plan.popular ? 'bg-cyan-500 text-white border-cyan-500' : 'hover:bg-gray-50'}`}>
              {plan.price === '0' ? '当前套餐' : '升级'}
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
