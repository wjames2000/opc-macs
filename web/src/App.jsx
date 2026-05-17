import React, { useState } from 'react';
import { Routes, Route, Link, useLocation } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Chat from './pages/Chat';
import Agents from './pages/Agents';
import Usage from './pages/Usage';
import Billing from './pages/Billing';
import Settings from './pages/Settings';
import Login from './pages/Login';
import Marketplace from './pages/Marketplace';
import Wallet from './pages/Wallet';
import AdminDashboard from './pages/AdminDashboard';
import Publish from './pages/Publish';
import VideoScript from './pages/VideoScript';
import TagGenerator from './pages/TagGenerator';
import TrendRadar from './pages/TrendRadar';

const navItems = [
  { path: '/', label: '总览', icon: '📊' },
  { path: '/chat', label: '对话', icon: '💬' },
  { path: '/agents', label: 'Agent', icon: '🤖' },
  { path: '/video-script', label: '视频脚本', icon: '🎬' },
  { path: '/tags', label: '标签推荐', icon: '🏷️' },
  { path: '/trends', label: '热点雷达', icon: '📡' },
  { path: '/publish', label: '发布', icon: '📤' },
  { path: '/marketplace', label: '市场', icon: '🏪' },
  { path: '/usage', label: '用量', icon: '📈' },
  { path: '/billing', label: '套餐', icon: '💎' },
  { path: '/wallet', label: '钱包', icon: '💰' },
  { path: '/admin', label: '管理', icon: '⚙️' },
  { path: '/settings', label: '设置', icon: '🔧' },
];

function Layout({ children }) {
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className={`fixed inset-y-0 left-0 z-50 w-60 bg-[#0F172A] text-white transform transition-transform lg:translate-x-0 ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'} lg:static lg:inset-auto`}>
        <div className="p-5 border-b border-white/10">
          <h1 className="text-lg font-bold">OPC-Agent</h1>
          <p className="text-xs text-cyan-400">AI 团队指挥中心</p>
        </div>
        <nav className="p-3 space-y-1">
          {navItems.map(item => (
            <Link key={item.path} to={item.path}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                location.pathname === item.path ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-white hover:bg-white/5'
              }`}
              onClick={() => setSidebarOpen(false)}
            >
              <span>{item.icon}</span> {item.label}
            </Link>
          ))}
        </nav>
        <div className="absolute bottom-0 p-4 border-t border-white/10 w-full text-xs text-gray-500">
          v0.1.0 · 5 Agents
        </div>
      </aside>

      {/* Overlay */}
      {sidebarOpen && <div className="fixed inset-0 bg-black/50 z-40 lg:hidden" onClick={() => setSidebarOpen(false)} />}

      {/* Main */}
      <div className="flex-1 flex flex-col min-h-screen">
        <header className="bg-white border-b px-6 py-3 flex items-center justify-between sticky top-0 z-30">
          <div className="flex items-center gap-3">
            <button className="lg:hidden text-xl" onClick={() => setSidebarOpen(true)}>☰</button>
            <h2 className="font-semibold text-lg">{navItems.find(i => i.path === location.pathname)?.label || 'OPC-Agent'}</h2>
          </div>
          <div className="flex items-center gap-4">
            <span className="flex items-center gap-1.5 text-xs bg-green-50 text-green-700 px-3 py-1 rounded-full">
              <span className="w-1.5 h-1.5 rounded-full bg-green-500" /> 系统正常
            </span>
            <span className="text-xs text-gray-400">admin@opc.com</span>
          </div>
        </header>
        <main className="flex-1 p-6 bg-gray-50/50">{children}</main>
      </div>
    </div>
  );
}

export default function App() {
  const [loggedIn, setLoggedIn] = useState(!!localStorage.getItem('token'));

  if (!loggedIn) return <Login onLogin={() => setLoggedIn(true)} />;

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/chat" element={<Chat />} />
        <Route path="/agents" element={<Agents />} />
        <Route path="/usage" element={<Usage />} />
          <Route path="/billing" element={<Billing />} />
          <Route path="/video-script" element={<VideoScript />} />
          <Route path="/tags" element={<TagGenerator />} />
          <Route path="/trends" element={<TrendRadar />} />
          <Route path="/publish" element={<Publish />} />
          <Route path="/marketplace" element={<Marketplace />} />
          <Route path="/wallet" element={<Wallet />} />
          <Route path="/admin" element={<AdminDashboard />} />
          <Route path="/settings" element={<Settings />} />
      </Routes>
    </Layout>
  );
}
