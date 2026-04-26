// Thin API client for the Kitty Party backend.
// Adds Bearer token automatically when present, unwraps the {success,data} envelope,
// and throws ApiError on non-2xx so pages can render real messages.
import { API_BASE } from './config.js';
import { store } from './store.js';

export class ApiError extends Error {
  constructor(message, { status, code, payload } = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.payload = payload;
  }
}

async function request(path, { method = 'GET', body, query, auth = false } = {}) {
  const url = new URL(API_BASE + path);
  if (query) {
    Object.entries(query).forEach(([k, v]) => {
      if (v !== undefined && v !== null && v !== '') url.searchParams.set(k, v);
    });
  }

  const headers = { 'Content-Type': 'application/json' };
  if (auth && store.token) headers['Authorization'] = `Bearer ${store.token}`;

  let res;
  try {
    res = await fetch(url, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (e) {
    throw new ApiError(`Network error: ${e.message}. Is the backend running at ${API_BASE}?`, { status: 0 });
  }

  if (res.status === 204) return null;

  let payload = null;
  const text = await res.text();
  if (text) {
    try { payload = JSON.parse(text); } catch (_) { payload = { raw: text }; }
  }

  if (!res.ok) {
    if (res.status === 401) {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    }
    const msg = payload?.error?.message || payload?.message || res.statusText || 'Request failed';
    throw new ApiError(msg, { status: res.status, code: payload?.error?.code, payload });
  }

  // Backend wraps successful responses in {success, data}
  return payload && Object.prototype.hasOwnProperty.call(payload, 'data') ? payload.data : payload;
}

const j = (params) => request(...params);

export const api = {
  health: () => request('/health'),

  // ─── Auth ──
  auth: {
    login:  (data) => request('/api/v1/auth/login',  { method: 'POST', body: data }),
    logout: ()     => request('/api/v1/auth/logout', { method: 'POST' }),
  },

  // ─── Members ──
  members: {
    list: () => request('/api/v1/members'),
    get: (id) => request(`/api/v1/members/${id}`),
    create: (data) => request('/api/v1/members', { method: 'POST', body: data }),
    update: (id, data) => request(`/api/v1/members/${id}`, { method: 'PUT', body: data }),
    remove: (id) => request(`/api/v1/members/${id}`, { method: 'DELETE' }),
  },

  // ─── Profiles ──
  profiles: {
    get: (memberId) => request(`/api/v1/members/${memberId}/profile`),
    summary: (memberId) => request(`/api/v1/members/${memberId}/profile/summary`),
    upsert: (memberId, data) => request(`/api/v1/members/${memberId}/profile`, { method: 'PUT', body: data, auth: true }),
  },

  // ─── Groups ──
  groups: {
    list: () => request('/api/v1/groups'),
    get: (id) => request(`/api/v1/groups/${id}`),
    create: (data) => request('/api/v1/groups', { method: 'POST', body: data }),
    update: (id, data) => request(`/api/v1/groups/${id}`, { method: 'PUT', body: data }),
    remove: (id) => request(`/api/v1/groups/${id}`, { method: 'DELETE' }),
    memberCounts: () => request('/api/v1/groups-member-counts'),
    myRole: (groupId) => request(`/api/v1/groups/${groupId}/my-role`, { auth: true }),
    members: {
      list: (groupId) => request(`/api/v1/groups/${groupId}/members`),
      add: (groupId, memberId) =>
        request(`/api/v1/groups/${groupId}/members`, { method: 'POST', body: { member_id: memberId }, auth: true }),
      count: (groupId) => request(`/api/v1/groups/${groupId}/members/count`),
      updateRole: (groupId, memberId, role) =>
        request(`/api/v1/groups/${groupId}/members/${memberId}/role`, { method: 'PATCH', body: { role }, auth: true }),
    },
  },

  // ─── Contributions ──
  contributions: {
    generateDues: (data) => request('/api/v1/contributions/dues/generate', { method: 'POST', body: data }),
    listDuesByGroup: (groupId) => request(`/api/v1/contributions/groups/${groupId}/dues`),
    recordPayment: (data) => request('/api/v1/contributions/payments', { method: 'POST', body: data }),
  },

  // ─── Tracking ──
  tracking: {
    create: (data) => request('/api/v1/contribution-tracking', { method: 'POST', body: data }),
    get: (id) => request(`/api/v1/contribution-tracking/${id}`),
    applyPayment: (id, data) => request(`/api/v1/contribution-tracking/${id}/payment`, { method: 'PATCH', body: data }),
    getByDue: (dueId) => request(`/api/v1/contribution-tracking/dues/${dueId}`),
    listByGroup: (groupId) => request(`/api/v1/contribution-tracking/groups/${groupId}`),
    listByMember: (memberId) => request(`/api/v1/contribution-tracking/members/${memberId}`),
  },

  // ─── Payouts ──
  payouts: {
    schedule: (data) => request('/api/v1/payouts/schedule', { method: 'POST', body: data }),
    disburse: (data) => request('/api/v1/payouts/disburse', { method: 'POST', body: data }),
    listByGroup: (groupId) => request(`/api/v1/payouts/groups/${groupId}`),
    listByRecipient: (memberId) => request(`/api/v1/payouts/recipients/${memberId}`),
    get: (id) => request(`/api/v1/payouts/${id}`),
    cancel: (id, notes) => request(`/api/v1/payouts/${id}/cancel`, { method: 'PATCH', body: { notes } }),
  },

  // ─── Kitty Cycles ──
  cycles: {
    schedule: (data) => request('/api/v1/kitty-cycles/schedule', { method: 'POST', body: data, auth: true }),
    listByGroup: (groupId) => request(`/api/v1/kitty-cycles/groups/${groupId}`, { auth: true }),
    get: (id) => request(`/api/v1/kitty-cycles/${id}`, { auth: true }),
    start: (id) => request(`/api/v1/kitty-cycles/${id}/start`, { method: 'PATCH', auth: true }),
    cancel: (id, notes) => request(`/api/v1/kitty-cycles/${id}/cancel`, { method: 'PATCH', body: { notes }, auth: true }),
    rollHost: (scheduleId) => request(`/api/v1/kitty-cycles/schedule-entries/${scheduleId}/roll-host`, { method: 'POST', auth: true }),
  },

  // ─── Schedule Invitations ──
  invitations: {
    send: (scheduleId) =>
      request(`/api/v1/kitty-cycles/schedule-entries/${scheduleId}/invitations`, { method: 'POST', auth: true }),
    listBySchedule: (scheduleId) =>
      request(`/api/v1/kitty-cycles/schedule-entries/${scheduleId}/invitations`, { auth: true }),
    my: () => request('/api/v1/invitations/my', { auth: true }),
    accept: (id) => request(`/api/v1/invitations/${id}/accept`, { method: 'PATCH', auth: true }),
    proposeDate: (id) => request(`/api/v1/invitations/${id}/propose-date`, { method: 'PATCH', auth: true }),
  },

  // ─── Date Proposals ──
  proposals: {
    create: (data) => request('/api/v1/date-proposals', { method: 'POST', body: data, auth: true }),
    get: (id) => request(`/api/v1/date-proposals/${id}`, { auth: true }),
    listBySchedule: (scheduleId) => request(`/api/v1/date-proposals/schedule/${scheduleId}`, { auth: true }),
    vote: (id, vote) => request(`/api/v1/date-proposals/${id}/votes`, { method: 'POST', body: { vote }, auth: true }),
    accept: (id) => request(`/api/v1/date-proposals/${id}/accept`, { method: 'PATCH', auth: true }),
  },
};
