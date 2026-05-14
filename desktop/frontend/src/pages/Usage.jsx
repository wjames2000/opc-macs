import React, { useState, useEffect } from 'react';

const agentLabels = {
  copywriter: '✍️ 文案生成',
  email_sorter: '📧 邮件分类',
  xhs_poster: '📕 小红书',
  competitive_analysis: '📊 竞品分析',
  meeting_minutes: '📋 会议纪要',
};

export default function Usage() {
  const [usage, setUsage] = useState([]);

  useEffect(() => {
    if (window.go?.main?.App?.GetUsageStats) {
      window.go.main.App.GetUsageStats().then(setUsage).catch(console.warn);
    }
  }, []);

  const totalTokens = usage.reduce((s, u) => s + u.in_tokens + u.out_tokens, 0);
  const totalCalls = usage.reduce((s, u) => s + u.calls, 0);
  const totalCost = usage.reduce((s, u) => s + u.cost, 0);

  return (
    <div>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        {[
          { label: '总 Token', value: totalTokens ? totalTokens.toLocaleString() : '-' },
          { label: '总调用', value: totalCalls ? `${totalCalls} 次` : '-' },
          { label: '总费用', value: totalCost ? `$${totalCost.toFixed(4)}` : '-' },
          { label: 'Agent 数', value: `${usage.length} 个` },
        ].map((s, i) => (
          <div key={i} className="bg-white rounded-xl border p-4">
            <div className="text-xl font-bold font-mono">{s.value}</div>
            <div className="text-sm text-gray-500">{s.label}</div>
          </div>
        ))}
      </div>

      <div className="bg-white rounded-xl border overflow-hidden">
        <div className="p-4 border-b">
          <h3 className="font-semibold">Agent 用量排行</h3>
        </div>
        {usage.length === 0 ? (
          <p className="text-sm text-gray-400 text-center py-8">暂无用量数据</p>
        ) : (
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
                </tr>
              </thead>
              <tbody>
                {usage.map((r, i) => (
                  <tr key={i} className="border-b last:border-0 hover:bg-gray-50">
                    <td className="p-3 font-medium">{agentLabels[r.agent] || r.agent}</td>
                    <td className="p-3 text-gray-500">{r.model || '-'}</td>
                    <td className="p-3">{r.calls}</td>
                    <td className="p-3 font-mono text-xs">{r.in_tokens.toLocaleString()}</td>
                    <td className="p-3 font-mono text-xs">{r.out_tokens.toLocaleString()}</td>
                    <td className="p-3 font-mono">${r.cost.toFixed(4)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
