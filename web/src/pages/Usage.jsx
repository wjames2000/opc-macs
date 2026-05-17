import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';

export default function Usage() {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const tenantId = localStorage.getItem('tenant_id') || 'demo';
    api.usageSummary(tenantId).then(data => {
      setStats(data);
    }).catch(() => {}).finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="p-8 text-gray-500">加载中...</div>;

  const statCards = [
    { label: '总 Token', value: stats?.total_tokens?.toLocaleString() || '--', sub: '累计用量' },
    { label: '总费用', value: stats?.total_cost ? `¥${stats.total_cost}` : '--', sub: '本月' },
    { label: 'Agent 数量', value: stats?.agent_count?.toString() || '--', sub: '已安装' },
    { label: '调用次数', value: stats?.total_calls?.toLocaleString() || '--', sub: '累计' },
  ];

  const rows = stats?.agent_stats || [];

  return (
    <div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {statCards.map((s, i) => (
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
                <tr key={r.agent || i} className="border-b last:border-0 hover:bg-gray-50">
                  <td className="p-3 font-medium">{r.agent}</td>
                  <td className="p-3 text-gray-500">{r.model || '--'}</td>
                  <td className="p-3 font-mono">{r.calls?.toLocaleString() || '--'}</td>
                  <td className="p-3 font-mono">{r.input_tokens?.toLocaleString() || '--'}</td>
                  <td className="p-3 font-mono">{r.output_tokens?.toLocaleString() || '--'}</td>
                  <td className="p-3">{r.cost || '--'}</td>
                  <td className="p-3">
                    <div className="flex items-center gap-2">
                      <div className="flex-1 h-1.5 bg-gray-100 rounded-full">
                        <div className="h-full rounded-full bg-blue-800" style={{ width: `${r.percentage || 0}%` }} />
                      </div>
                      <span className="text-xs text-gray-500 w-8 text-right">{r.percentage || 0}%</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
