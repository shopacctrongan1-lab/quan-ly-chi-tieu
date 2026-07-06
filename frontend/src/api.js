const API_BASE = import.meta.env.VITE_API_BASE || ''
let token = localStorage.getItem('token') || ''

export function setToken(value) {
  token = value || ''
  if (token) localStorage.setItem('token', token)
  else localStorage.removeItem('token')
}

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      ...(options.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {}),
    },
    ...options,
  })
  if (!response.ok) {
    let message = 'Có lỗi xảy ra'
    try { message = (await response.json()).error || message } catch (_) {}
    throw new Error(message)
  }
  if (response.status === 204) return null
  return response.json()
}

const qs = (query = {}) => {
  const params = new URLSearchParams()
  Object.entries(query).forEach(([k, v]) => { if (v !== '' && v !== null && v !== undefined) params.set(k, v) })
  return params.toString() ? `?${params}` : ''
}

export const api = {
  login: (payload) => request('/api/auth/login', { method: 'POST', body: JSON.stringify(payload) }),
  register: (payload) => request('/api/auth/register', { method: 'POST', body: JSON.stringify(payload) }),
  logout: () => request('/api/auth/logout', { method: 'POST' }),
  me: () => request('/api/me'),
  expenses: (query) => request(`/api/expenses${qs(query)}`),
  summary: (query) => request(`/api/summary${qs(query)}`),
  createExpense: (payload) => request('/api/expenses', { method: 'POST', body: JSON.stringify(payload) }),
  updateExpense: (id, payload) => request(`/api/expenses/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  deleteExpense: (id) => request(`/api/expenses/${id}`, { method: 'DELETE' }),
  categories: () => request('/api/categories'),
  createCategory: (payload) => request('/api/categories', { method: 'POST', body: JSON.stringify(payload) }),
  updateCategory: (id, payload) => request(`/api/categories/${id}`, { method: 'PUT', body: JSON.stringify(payload) }),
  deleteCategory: (id) => request(`/api/categories/${id}`, { method: 'DELETE' }),
  wallets: () => request('/api/wallets'),
  createWallet: (payload) => request('/api/wallets', { method: 'POST', body: JSON.stringify(payload) }),
  transfer: (payload) => request('/api/wallets/transfer', { method: 'POST', body: JSON.stringify(payload) }),
  budgets: (month) => request(`/api/budgets${qs({ month })}`),
  saveBudget: (payload) => request('/api/budgets', { method: 'POST', body: JSON.stringify(payload) }),
  goals: () => request('/api/goals'),
  saveGoal: (payload) => request('/api/goals', { method: 'POST', body: JSON.stringify(payload) }),
  debts: () => request('/api/debts'),
  saveDebt: (payload) => request('/api/debts', { method: 'POST', body: JSON.stringify(payload) }),
  completeDebt: (id, walletId) => request(`/api/debts/complete/${id}`, { method: 'POST', body: JSON.stringify({ walletId }) }),
  deleteDebt: (id) => request(`/api/debts/${id}`, { method: 'DELETE' }),
  deleteAccount: () => request('/api/account', { method: 'DELETE' }),
  reminder: () => request('/api/reminder'),
  saveReminder: (payload) => request('/api/reminder', { method: 'POST', body: JSON.stringify(payload) }),
  testReminder: () => request('/api/reminder/test', { method: 'POST' }),
  importCsv: (file) => { const body = new FormData(); body.append('file', file); return request('/api/import.csv', { method: 'POST', body, headers: {} }) },
  exportUrl: (query) => `${API_BASE}/api/export.xlsx${qs(query)}`,
}
