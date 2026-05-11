import React from 'react';

const colorMap = {
  blue: 'bg-blue-50 text-blue-600',
  yellow: 'bg-yellow-50 text-yellow-600',
  green: 'bg-green-50 text-green-600',
  purple: 'bg-purple-50 text-purple-600',
};

export default function StatCard({ icon, label, value, sub, color = 'blue' }) {
  return (
    <div className="bg-white rounded-xl border p-4 hover:shadow-sm transition-shadow">
      <div className={`w-10 h-10 rounded-xl ${colorMap[color]} flex items-center justify-center text-lg mb-3`}>{icon}</div>
      <div className="text-2xl font-bold font-mono">{value}</div>
      <div className="text-sm text-gray-500 mt-0.5">{label}</div>
      <div className="text-xs mt-2 text-gray-400">{sub}</div>
    </div>
  );
}
