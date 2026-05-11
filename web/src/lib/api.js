const API_BASE = '/api/v1';

async function request(path, options = {}) {
  const token = localStorage.getItem('token');
  const headers = { 'Content-Type': 'application/json', ...options.headers };
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || 'Request failed');
  }
  return res.json();
}

export const api = {
  // Auth
  register: (data) => request('/auth/register', { method: 'POST', body: JSON.stringify(data) }),
  login: (data) => request('/auth/login', { method: 'POST', body: JSON.stringify(data) }),

  // Agents
  listAgents: () => request('/agents'),

  // Usage
  recordUsage: (data) => request('/usage/record', { method: 'POST', body: JSON.stringify(data) }),
  usageSummary: (tenantId) => request(`/usage/summary?tenant_id=${tenantId}`),

  // Tenants
  listTenants: () => request('/tenants'),
};
