import React from 'react';

const plans = [
  { name: 'Free', price: '0', desc: '适合个人试用', features: ['10 万 Token/月', '3 个 Agent', '1 个用户', '2 个工作流'], limit: false, color: 'gray', users: 28 },
  { name: 'Pro', price: '99', desc: '适合小型团队', features: ['100 万 Token/月', '10 个 Agent', '5 个用户', '20 个工作流', '超额 ¥0.008/千 tokens'], limit: true, color: 'cyan', users: 10, popular: true },
  { name: 'Enterprise', price: '499', desc: '适合企业', features: ['1000 万 Token/月', '不限 Agent', '不限用户', '不限工作流', '超额 ¥0.005/千 tokens', '专属技术支持'], limit: false, color: 'blue', users: 4 },
];

export default function Billing() {
  const used = 845200, limit = 1000000, pct = (used / limit * 100).toFixed(1);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold">💎 套餐与账单</h2>
          <p className="text-sm text-gray-500">当前套餐: Pro · 续费: 2026-06-11</p>
        </div>
        <button className="bg-blue-800 text-white px-4 py-2 rounded-lg text-sm">⬆️ 升级套餐</button>
      </div>

      {/* Usage gauge */}
      <div className="bg-white rounded-xl border p-5 mb-6">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
          {[
            { label: '本月已用 Token', value: '845,200', sub: '↑ 12% 较上月' },
            { label: '套餐限额', value: '1,000,000', sub: 'Pro 套餐' },
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
        {plans.map((plan, i) => (
          <div key={i} className={`bg-white rounded-xl border p-6 ${plan.popular ? 'border-cyan-500 ring-1 ring-cyan-500 relative' : ''}`}>
            {plan.popular && <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-cyan-500 text-white text-xs px-3 py-0.5 rounded-full font-medium">最受欢迎</div>}
            <div className="text-xs font-semibold text-gray-500 uppercase tracking-wider">{plan.name}</div>
            <div className="mt-3 mb-1"><span className="text-3xl font-bold">¥{plan.price}</span>{plan.price !== '0' && <span className="text-sm text-gray-500">/月</span>}</div>
            <p className="text-xs text-gray-500 mb-4">{plan.desc}</p>
            <div className="space-y-2 text-sm">
              {plan.features.map((f, j) => (
                <div key={j} className="flex items-center gap-2"><span className="text-green-500">✅</span> {f}</div>
              ))}
            </div>
            <div className="text-xs text-gray-400 mt-4">当前 {plan.users} 个租户使用</div>
          </div>
        ))}
      </div>

      {/* Usage table */}
      <div className="bg-white rounded-xl border overflow-hidden">
        <div className="p-4 border-b font-semibold text-sm">用量明细</div>
        <table className="w-full text-sm">
          <thead><tr className="text-left text-xs text-gray-500 border-b">
            <th className="p-3">Agent</th><th className="p-3">调用</th><th className="p-3 font-mono">输入 Token</th>
            <th className="p-3 font-mono">输出 Token</th><th className="p-3">费用</th>
          </tr></thead>
          <tbody>
            {[
              { agent: 'copywriter', calls: 156, in: '45,200', out: '28,300', cost: '$1.42' },
              { agent: 'email_sorter', calls: 89, in: '12,400', out: '8,100', cost: '$0.41' },
              { agent: 'xhs_poster', calls: 67, in: '32,100', out: '19,800', cost: '$0.99' },
              { agent: 'competitive_analysis', calls: 23, in: '18,500', out: '12,400', cost: '$0.62' },
            ].map((r, i) => (
              <tr key={i} className="border-b last:border-0 hover:bg-gray-50">
                <td className="p-3 font-medium">{r.agent}</td>
                <td className="p-3">{r.calls}</td>
                <td className="p-3 font-mono text-xs">{r.in}</td>
                <td className="p-3 font-mono text-xs">{r.out}</td>
                <td className="p-3 font-mono text-xs">{r.cost}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
