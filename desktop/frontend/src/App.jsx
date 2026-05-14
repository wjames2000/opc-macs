import React, { useState, useEffect } from 'react';
import Dashboard from './pages/Dashboard';
import Chat from './pages/Chat';
import Agents from './pages/Agents';
import Usage from './pages/Usage';
import Settings from './pages/Settings';

const navItems = [
  { key: 'dashboard', label: '总览', icon: '📊' },
  { key: 'chat', label: '对话', icon: '💬' },
  { key: 'agents', label: 'Agent', icon: '🤖' },
  { key: 'usage', label: '用量', icon: '📈' },
  { key: 'settings', label: '设置', icon: '⚙️' },
];

export default function App() {
  const [page, setPage] = useState('chat');
  const [version, setVersion] = useState({ version: '0.1.0', agents: '0' });

  useEffect(() => {
    if (window.go?.main?.App?.GetVersion) {
      window.go.main.App.GetVersion().then(setVersion).catch(console.warn);
    }
  }, []);

  const renderPage = () => {
    switch (page) {
      case 'dashboard': return <Dashboard onNavigate={setPage} />;
      case 'agents': return <Agents />;
      case 'usage': return <Usage />;
      case 'settings': return <Settings onClose={() => setPage('chat')} />;
      case 'chat':
      default: return <Chat />;
    }
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <header className="bg-[#0F172A] text-white px-5 pt-8 pb-3 flex items-center justify-between select-none drag">
        <div>
          <h1 className="font-bold text-base">OPC-Agent</h1>
          <p className="text-xs text-cyan-400">AI 团队指挥中心</p>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-gray-400">v{version.version} · {version.agents} Agents</span>
          <button className={`text-sm px-3 py-1 rounded ${page === 'settings' ? 'bg-blue-600' : 'text-white/70 hover:text-white'}`}
            onClick={() => setPage(page === 'settings' ? 'chat' : 'settings')}
            title="设置">⚙️</button>
        </div>
      </header>

      {/* Nav */}
      <nav className="bg-white border-b flex px-4 gap-1">
        {navItems.filter(n => n.key !== 'settings').map(item => (
          <button key={item.key}
            className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
              page === item.key
                ? 'border-[#1E3A5F] text-[#1E3A5F]'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
            onClick={() => setPage(item.key)}>
            {item.icon} {item.label}
          </button>
        ))}
      </nav>

      {/* Main content */}
      <main className="flex-1 overflow-y-auto p-6">
        {renderPage()}
      </main>
    </div>
  );
}
