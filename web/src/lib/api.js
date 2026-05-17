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
  executeAgent: (data) => request('/agents/execute', { method: 'POST', body: JSON.stringify(data) }),

  // Usage
  recordUsage: (data) => request('/usage/record', { method: 'POST', body: JSON.stringify(data) }),
  usageSummary: (tenantId) => request(`/usage/summary?tenant_id=${tenantId}`),

  // Tenants
  listTenants: () => request('/tenants'),

  // Marketplace - Tasks
  listOpenTasks: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/marketplace/open-tasks${qs ? '?' + qs : ''}`);
  },
  listMyTasks: () => request('/marketplace/tasks'),
  getTask: (id) => request(`/marketplace/tasks/${id}`),
  createTask: (data) => request('/marketplace/tasks', { method: 'POST', body: JSON.stringify(data) }),
  cancelTask: (id) => request(`/marketplace/tasks/${id}/cancel`, { method: 'POST' }),

  // Marketplace - Orders
  applyTask: (id) => request(`/marketplace/tasks/${id}/apply`, { method: 'POST' }),
  listMyOrders: () => request('/marketplace/my-orders'),
  getOrder: (id) => request(`/marketplace/orders/${id}`),
  deliverOrder: (id, data) => request(`/marketplace/orders/${id}/deliver`, { method: 'POST', body: JSON.stringify(data) }),
  approveOrder: (id, data) => request(`/marketplace/orders/${id}/approve`, { method: 'POST', body: JSON.stringify(data) }),
  rejectOrder: (id, data) => request(`/marketplace/orders/${id}/reject`, { method: 'POST', body: JSON.stringify(data) }),

  // Wallet
  getWallet: () => request('/wallet'),
  getTransactions: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/wallet/transactions${qs ? '?' + qs : ''}`);
  },
  createWithdrawal: (data) => request('/wallet/withdrawals', { method: 'POST', body: JSON.stringify(data) }),
  listWithdrawals: () => request('/wallet/withdrawals'),

  // Settlement
  calculateSettlement: (data) => request('/settlement/calculate', { method: 'POST', body: JSON.stringify(data) }),

  // Credit & Reviews
  addReview: (data) => request('/credit/reviews', { method: 'POST', body: JSON.stringify(data) }),
  getMyScore: () => request('/credit/score'),
  getUserScore: (userId) => request(`/credit/score/${userId}`),

  // Notifications
  listNotifications: (unreadOnly = false) => request(`/notifications${unreadOnly ? '?unread_only=true' : ''}`),
  countUnread: () => request('/notifications/unread-count'),
  markRead: (id) => request(`/notifications/read/${id}`, { method: 'POST' }),
  createNotification: (data) => request('/notifications', { method: 'POST', body: JSON.stringify(data) }),
  deleteNotification: (id) => request(`/notifications/${id}`, { method: 'DELETE' }),

  // Lifecycle
  completeOrder: (data) => request('/lifecycle/complete-order', { method: 'POST', body: JSON.stringify(data) }),
  cancelOrder: (data) => request('/lifecycle/cancel-order', { method: 'POST', body: JSON.stringify(data) }),

  // Publish
  publish: (data) => request('/publish', { method: 'POST', body: JSON.stringify(data) }),
  listSchedules: () => request('/publish/schedules'),

  // Video Script
  generateVideoScript: (data) => request('/video-script/generate', { method: 'POST', body: JSON.stringify(data) }),
  getVideoScriptInfo: () => request('/video-script/info'),

  // Tag Generator
  generateTags: (data) => request('/tags/generate', { method: 'POST', body: JSON.stringify(data) }),
  getTagGeneratorInfo: () => request('/tags/info'),

  // Trend Radar
  getTrendReport: (data) => request('/trend-radar/report', { method: 'POST', body: JSON.stringify(data) }),
  getTrendRadarInfo: () => request('/trend-radar/info'),

  // Invite
  inviteCreator: (data) => request('/marketplace/invite', { method: 'POST', body: JSON.stringify(data) }),
  listInvitations: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/marketplace/invitations${qs ? '?' + qs : ''}`);
  },
  acceptInvite: (id) => request(`/marketplace/invitations/${id}/accept`, { method: 'POST' }),
  declineInvite: (id) => request(`/marketplace/invitations/${id}/decline`, { method: 'POST' }),

  // Recommendations
  getRecommendations: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/marketplace/recommendations${qs ? '?' + qs : ''}`);
  },

  // Engage
  engageReply: (data) => request('/engage/reply', { method: 'POST', body: JSON.stringify(data) }),
  engageAnalyze: (data) => request('/engage/analyze', { method: 'POST', body: JSON.stringify(data) }),

  // Publish
  listSchedules: () => request('/publish/schedules'),
  publish: (data) => request('/publish', { method: 'POST', body: JSON.stringify(data) }),

  // Admin
  adminPendingWithdrawals: () => request('/admin/withdrawals'),
  adminProcessWithdrawal: (id, data) => request(`/admin/withdrawals/${id}/process`, { method: 'POST', body: JSON.stringify(data) }),

  // Invoices
  createInvoice: (data) => request('/invoices', { method: 'POST', body: JSON.stringify(data) }),
  listInvoices: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/invoices${qs ? '?' + qs : ''}`);
  },
  getInvoice: (id) => request(`/invoices/${id}`),
  payInvoice: (id) => request(`/invoices/${id}/pay`, { method: 'POST' }),
  cancelInvoice: (id) => request(`/invoices/${id}/cancel`, { method: 'POST' }),

  // Payment Methods
  listPaymentMethods: () => request('/payment-methods'),
  upsertPaymentMethod: (data) => request('/payment-methods', { method: 'POST', body: JSON.stringify(data) }),
  deletePaymentMethod: (id) => request(`/payment-methods/${id}`, { method: 'DELETE' }),
  setDefaultPaymentMethod: (id) => request(`/payment-methods/${id}/default`, { method: 'POST' }),

  // Report Export
  exportOrdersCSV: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/reports/export/orders${qs ? '?' + qs : ''}`);
  },
  exportTasksCSV: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/reports/export/tasks${qs ? '?' + qs : ''}`);
  },

  // Fraud Detection
  listFraudRules: () => request('/fraud/rules'),
  createFraudRule: (data) => request('/fraud/rules', { method: 'POST', body: JSON.stringify(data) }),
  toggleFraudRule: (id, data) => request(`/fraud/rules/${id}/toggle`, { method: 'POST', body: JSON.stringify(data) }),
  listFraudFlags: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/fraud/flags${qs ? '?' + qs : ''}`);
  },

  // Fraud Account
  checkDuplicateAccount: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/fraud/accounts/check${qs ? '?' + qs : ''}`);
  },
  reportSuspiciousActivity: (data) => request('/fraud/accounts/report', { method: 'POST', body: JSON.stringify(data) }),
  listSuspiciousActivities: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/fraud/accounts/activities${qs ? '?' + qs : ''}`);
  },

  // Fraud Content
  checkFraudContent: (data) => request('/fraud/content/check', { method: 'POST', body: JSON.stringify(data) }),
  listContentChecks: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/fraud/content/checks${qs ? '?' + qs : ''}`);
  },

  // Fraud Alerts
  listFraudAlerts: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/fraud/alerts${qs ? '?' + qs : ''}`);
  },
  createFraudAlert: (data) => request('/fraud/alerts', { method: 'POST', body: JSON.stringify(data) }),
  updateFraudAlertStatus: (id, data) => request(`/fraud/alerts/${id}/status`, { method: 'POST', body: JSON.stringify(data) }),
  getFraudAlertStats: () => request('/fraud/alerts/stats'),

  // Withdraw Payments (vendor-initiated)
  createWithdrawPayment: (data) => request('/withdraw-payments', { method: 'POST', body: JSON.stringify(data) }),
  listWithdrawPayments: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/withdraw-payments${qs ? '?' + qs : ''}`);
  },
  getWithdrawPayment: (id) => request(`/withdraw-payments/${id}`),
  processWithdrawPayment: (id, data) => request(`/withdraw-payments/${id}/process`, { method: 'POST', body: JSON.stringify(data) }),

  // Reports - Advertiser
  getAdvertiserReport: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/reports/advertiser/summary${qs ? '?' + qs : ''}`);
  },

  // Reports - Creator
  getCreatorReportSummary: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/reports/creator/summary${qs ? '?' + qs : ''}`);
  },
  getCreatorCompletedOrders: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/reports/creator/orders${qs ? '?' + qs : ''}`);
  },
  getCreatorMonthlyEarnings: (params = {}) => {
    const qs = new URLSearchParams(params).toString();
    return request(`/reports/creator/monthly${qs ? '?' + qs : ''}`);
  },
};
