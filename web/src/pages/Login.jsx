import React, { useState } from 'react';
import { api } from '../lib/api';

export default function Login({ onLogin }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleLogin = async (e) => {
    e.preventDefault();
    try {
      const res = await api.login({ email, password });
      localStorage.setItem('token', res.token);
      localStorage.setItem('tenant_id', res.tenant_id);
      onLogin();
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-900 to-slate-900">
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">OPC-Agent</h1>
          <p className="text-sm text-gray-500 mt-1">登录你的 AI 团队指挥中心</p>
        </div>
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
            <input type="email" required value={email} onChange={e => setEmail(e.target.value)}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none" placeholder="admin@example.com" />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">密码</label>
            <input type="password" required value={password} onChange={e => setPassword(e.target.value)}
              className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none" />
          </div>
          {error && <p className="text-red-500 text-sm">{error}</p>}
          <button type="submit" className="w-full bg-blue-800 hover:bg-blue-900 text-white py-2.5 rounded-lg font-medium transition-colors">
            登 录
          </button>
        </form>
        <p className="text-xs text-gray-400 text-center mt-6">
          还没有账号？<a href="#" className="text-blue-600" onClick={e => { e.preventDefault(); alert('请联系管理员注册'); }}>联系管理员</a>
        </p>
      </div>
    </div>
  );
}
