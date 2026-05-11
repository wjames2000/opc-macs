import React from 'react';

const rows = [
  { agent: 'copywriter', model: 'deepseek-v4-flash', calls: 156, inTokens: '45,200', outTokens: '28,300', cost: '$1.42', pct: 80 },
  { agent: 'xhs_poster', model: 'deepseek-v4-flash', calls: 67, inTokens: '32,100', outTokens: '19,800', cost: '$0.99', pct: 55 },
  { agent: 'email_sorter', model: 'gemini-2.0-flash', calls: 89, inTokens: '12,400', outTokens: '8,100', cost: '$0.41', pct: 28 },
  { agent: 'competitive_analysis', model: 'deepseek-v4-flash', calls: 23, inTokens: '18,500', outTokens: '12,400', cost: '$0.62', pct: 35 },
  { agent: 'meeting_minutes', model: 'deepseek-v4-flash', calls: 12, inTokens: '6,800', outTokens: '4,200', cost: '$0.21', pct: 15 },
];

export default function Usage() {
  return (
    <div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {[
          { label: '总 Token', value: '124,500', sub: '↑ 12% 较上月' },
          { label: '总费用', value: '$3.12', sub: '预算内 ✅' },
          { label: '平均/任务', value: '2,650', sub: '↓ 5%' },
          { label: '缓存命中', value: '87%', sub: '↑ 缓存优化' },
        ].map((s, i) => (
          <div key={i} className="bg-white rounded-xl border p-4">
            <div className="text-xl font-bold font-mono">{s.value}</div>
            <div className="text-sm text-gray-500">{s.label}</div>
            <div className="text-xs text-gray-400 mt-1">{s.sub}</div>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-xl border overflow-hidden">
        <div className="p-4 border-b flex justify-between items-center">
          <h3 className="font-semibold">Agent 用量排行</h3>
          <div className="flex gap-2">
            <button className="text-xs border px-3 py-1 rounded">7 天</button>
            <button className="text-xs bg-blue-800 text-white px-3 py-1 rounded">30 天</button>
          </div>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs text-gray-500 border-b">
                <th className="p-3 font-medium">Agent</th>
                <th className="p-3 font-medium">模型</th>
                <th className="p-3 font-medium">调用</th>
                <th className="p-3 font-medium font-mono">输入 Token</th>
                <th className="p-3 font-medium font-mono">输出 Token</th>
                <th className="p-3 font-medium">费用</th>
                <th className="p-3 font-medium">占比</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r, i) => (
                <tr key={i} className="border-b last:border-0 hover:bg-gray-50">
                  <td className="p-3 font-medium">{r.agent}</td>
                  <td className="p-3"><span className="text-xs bg-blue-50 text-blue-700 px-2 py-0.5 rounded">{r.model}</span></td>
                  <td className="p-3">{r.calls}</td>
                  <td className="p-3 font-mono text-xs">{r.inTokens}</td>
                  <td className="p-3 font-mono text-xs">{r.outTokens}</td>
                  <td className="p-3 font-mono text-xs">{r.cost}</td>
                  <td className="p-3"><div className="w-20 h-1.5 bg-gray-100 rounded"><div className="h-full rounded bg-cyan-500" style={{width: `${r.pct}%`}} /></div></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
