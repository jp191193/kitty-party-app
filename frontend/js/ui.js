// Reusable UI primitives: html template tag, toasts, modals, formatters,
// status badges, common empty/loading states, and a tiny form helper.

export function html(strings, ...values) {
  // Tagged template that returns a string. Values are NOT auto-escaped — use escape() for user data.
  let out = '';
  strings.forEach((s, i) => {
    out += s;
    if (i < values.length) {
      const v = values[i];
      out += v === null || v === undefined ? '' : Array.isArray(v) ? v.join('') : String(v);
    }
  });
  return out;
}

export function escape(s) {
  if (s === null || s === undefined) return '';
  return String(s)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

// ── Toasts ────────────────────────────────────────────────────────────────
export function toast(message, { type = 'info', title, duration = 3500 } = {}) {
  const root = document.getElementById('toastContainer');
  if (!root) { console.log(`[${type}]`, title || '', message); return; }
  const el = document.createElement('div');
  el.className = `toast ${type}`;
  el.innerHTML =
    (title ? `<div class="toast-title">${escape(title)}</div>` : '') +
    `<div>${escape(message)}</div>`;
  root.appendChild(el);
  const remove = () => {
    el.classList.add('hide');
    setTimeout(() => el.remove(), 200);
  };
  setTimeout(remove, duration);
  el.addEventListener('click', remove);
}

// ── Modals ────────────────────────────────────────────────────────────────
export function openModal({ title, body, size = 'md' }) {
  const root = document.getElementById('modalRoot');
  const backdrop = document.createElement('div');
  backdrop.className = 'modal-backdrop';
  backdrop.innerHTML = `
    <div class="modal ${size === 'lg' ? 'lg' : ''}" role="dialog" aria-modal="true">
      <div class="modal-header">
        <h3>${escape(title || '')}</h3>
        <button class="modal-close" aria-label="Close">&times;</button>
      </div>
      <div class="modal-body"></div>
    </div>`;
  const modalBody = backdrop.querySelector('.modal-body');
  if (typeof body === 'string') modalBody.innerHTML = body;
  else if (body instanceof HTMLElement) modalBody.appendChild(body);

  const close = () => {
    backdrop.style.opacity = '0';
    setTimeout(() => backdrop.remove(), 150);
  };
  backdrop.addEventListener('click', (e) => { if (e.target === backdrop) close(); });
  backdrop.querySelector('.modal-close').addEventListener('click', close);
  document.addEventListener('keydown', function esc(e) {
    if (e.key === 'Escape') { close(); document.removeEventListener('keydown', esc); }
  });

  root.appendChild(backdrop);
  return { close, backdrop, body: modalBody };
}

export async function confirm({ title = 'Confirm', message, confirmText = 'Confirm', danger = false } = {}) {
  return new Promise((resolve) => {
    const wrap = document.createElement('div');
    wrap.innerHTML = `
      <p style="margin-top:0">${escape(message || 'Are you sure?')}</p>
      <div class="form-actions">
        <button class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button class="btn ${danger ? 'btn-danger' : ''}" data-act="ok">${escape(confirmText)}</button>
      </div>`;
    const m = openModal({ title, body: wrap });
    wrap.querySelector('[data-act="cancel"]').onclick = () => { m.close(); resolve(false); };
    wrap.querySelector('[data-act="ok"]').onclick = () => { m.close(); resolve(true); };
  });
}

// ── Loading / empty states ────────────────────────────────────────────────
export const loadingHtml = (label = 'Loading…') =>
  `<div class="loader"><div class="spinner"></div><p>${escape(label)}</p></div>`;

export const emptyHtml = (title, subtitle) => `
  <div class="empty">
    <div class="empty-icon">
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>
    </div>
    <div class="empty-title">${escape(title)}</div>
    ${subtitle ? `<div>${escape(subtitle)}</div>` : ''}
  </div>`;

export const errorHtml = (msg) => `
  <div class="empty">
    <div class="empty-icon" style="background:var(--danger-bg);color:var(--danger)">
      <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
      </svg>
    </div>
    <div class="empty-title">Something went wrong</div>
    <div>${escape(msg)}</div>
  </div>`;

// ── Status badges (centralised so colors stay consistent) ─────────────────
const STATUS_BADGE = {
  ACTIVE:      'success',
  PAID:        'success',
  SUCCESS:     'success',
  COMPLETED:   'success',
  DISBURSED:   'success',
  PENDING:     'warning',
  PARTIAL:     'warning',
  SCHEDULED:   'info',
  IN_PROGRESS: 'info',
  OVERDUE:     'danger',
  FAILED:      'danger',
  REJECTED:    'danger',
  CANCELLED:   'muted',
  ADMIN:       'primary',
  MEMBER:      'muted',
  UPI:         'info',
  BANK:        'info',
  CASH:        'muted',
};
export function statusBadge(status) {
  if (!status) return '';
  const cls = STATUS_BADGE[status] || 'muted';
  return `<span class="badge badge-${cls}">${escape(status)}</span>`;
}

// ── Formatters ────────────────────────────────────────────────────────────
const inr = new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 });
export const fmtMoney = (n) => (n === null || n === undefined || isNaN(+n)) ? '—' : inr.format(+n);

export function fmtDate(s, { withTime = false } = {}) {
  if (!s) return '—';
  const d = new Date(s);
  if (isNaN(d)) return s;
  const opts = withTime
    ? { day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }
    : { day: '2-digit', month: 'short', year: 'numeric' };
  return d.toLocaleString('en-IN', opts);
}

export function relTime(s) {
  if (!s) return '';
  const diff = (Date.now() - new Date(s).getTime()) / 1000;
  if (isNaN(diff)) return '';
  const abs = Math.abs(diff);
  const sign = diff >= 0 ? 'ago' : 'from now';
  if (abs < 60) return `${Math.round(abs)}s ${sign}`;
  if (abs < 3600) return `${Math.round(abs / 60)}m ${sign}`;
  if (abs < 86400) return `${Math.round(abs / 3600)}h ${sign}`;
  return `${Math.round(abs / 86400)}d ${sign}`;
}

export const initials = (name) => {
  if (!name) return '?';
  return name.trim().split(/\s+/).slice(0, 2).map((p) => p[0]?.toUpperCase() || '').join('') || '?';
};

export const shortId = (id) => id ? id.slice(0, 8) : '—';

// ── Form helper ───────────────────────────────────────────────────────────
// Reads named inputs from a <form> element into a plain object,
// converting blank strings to undefined and `[type=number]` to numbers.
export function readForm(form) {
  const data = {};
  for (const el of form.elements) {
    if (!el.name) continue;
    let v = el.value;
    if (v === '') { data[el.name] = undefined; continue; }
    if (el.type === 'number') v = Number(v);
    if (el.type === 'datetime-local') {
      // <input type=datetime-local> has no timezone — append local TZ for an ISO string the API accepts.
      v = new Date(v).toISOString();
    }
    data[el.name] = v;
  }
  return data;
}

// Animate table rows on insert (subtle stagger). Call after appending rows.
export function animateRows(tbody, max = 12) {
  const rows = tbody.querySelectorAll('tr');
  rows.forEach((r, i) => {
    if (i >= max) return;
    r.classList.add('anim');
    r.style.animationDelay = `${i * 18}ms`;
  });
}
