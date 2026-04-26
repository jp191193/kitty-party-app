// Reusable UI primitives: html template tag, toasts, modals, formatters,
// status badges, common empty/loading states, and a tiny form helper.

export function html(strings, ...values) {
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
  el.style.setProperty('--toast-duration', `${duration}ms`);
  el.innerHTML =
    `<div class="toast-content">` +
    (title ? `<div class="toast-title">${escape(title)}</div>` : '') +
    `<div>${escape(message)}</div></div>` +
    `<button class="toast-close" aria-label="Dismiss">×</button>` +
    `<div class="toast-progress"></div>`;

  root.appendChild(el);

  const remove = () => {
    if (el._removing) return;
    el._removing = true;
    el.classList.add('hide');
    setTimeout(() => el.remove(), 200);
  };

  const timer = setTimeout(remove, duration);
  el.querySelector('.toast-close').addEventListener('click', () => { clearTimeout(timer); remove(); });
  el.addEventListener('mouseenter', () => {
    el.querySelector('.toast-progress').style.animationPlayState = 'paused';
  });
  el.addEventListener('mouseleave', () => {
    el.querySelector('.toast-progress').style.animationPlayState = 'running';
  });
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
  const modalEl   = backdrop.querySelector('.modal');
  const modalBody = backdrop.querySelector('.modal-body');
  if (typeof body === 'string') modalBody.innerHTML = body;
  else if (body instanceof HTMLElement) modalBody.appendChild(body);

  const close = () => {
    if (backdrop._closing) return;
    backdrop._closing = true;
    modalEl.classList.add('closing');
    backdrop.classList.add('closing');
    setTimeout(() => backdrop.remove(), 180);
  };

  backdrop.addEventListener('click', (e) => { if (e.target === backdrop) close(); });
  backdrop.querySelector('.modal-close').addEventListener('click', close);
  document.addEventListener('keydown', function esc(e) {
    if (e.key === 'Escape') { close(); document.removeEventListener('keydown', esc); }
  });

  root.appendChild(backdrop);
  // Focus the modal for accessibility
  requestAnimationFrame(() => {
    const first = modalEl.querySelector('input, select, textarea, button:not(.modal-close)');
    if (first) first.focus();
  });
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
    wrap.querySelector('[data-act="ok"]').onclick    = () => { m.close(); resolve(true); };
  });
}

// ── Loading / empty states ────────────────────────────────────────────────
export const loadingHtml = (label = 'Loading…') =>
  `<div class="loader"><div class="spinner"></div><p>${escape(label)}</p></div>`;

// Skeleton table rows — mimics a real data table while loading
export function skeletonTableHtml(rows = 5, widths = ['32px', '28%', '22%', '15%', '12%', '80px']) {
  const cells = widths.map((w) =>
    `<td><div class="skeleton sk-line" style="width:${w};max-width:100%"></div></td>`).join('');
  const rowsHtml = Array.from({ length: rows }, () => `<tr class="skeleton-row">${cells}</tr>`).join('');
  return `<div class="table-wrap"><table class="data-table"><tbody>${rowsHtml}</tbody></table></div>`;
}

// Skeleton for a card body with N lines
export function skeletonLinesHtml(lines = 3) {
  const widths = ['80%', '65%', '75%', '55%', '70%', '60%'];
  return Array.from({ length: lines }, (_, i) =>
    `<div class="skeleton sk-line" style="width:${widths[i % widths.length]};margin-bottom:10px"></div>`
  ).join('');
}

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

// ── Button loading state ───────────────────────────────────────────────────
export function setButtonLoading(btn, loading) {
  if (loading) {
    btn.__savedHTML = btn.innerHTML;
    btn.disabled = true;
    btn.classList.add('btn-loading');
    btn.innerHTML = `<span class="btn-text">${btn.__savedHTML}</span><span class="btn-spinner"></span>`;
  } else {
    btn.disabled = false;
    btn.classList.remove('btn-loading');
    if (btn.__savedHTML !== undefined) {
      btn.innerHTML = btn.__savedHTML;
      delete btn.__savedHTML;
    }
  }
}

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
export function readForm(form) {
  const data = {};
  for (const el of form.elements) {
    if (!el.name) continue;
    let v = el.value;
    if (v === '') { data[el.name] = undefined; continue; }
    if (el.type === 'number') v = Number(v);
    if (el.type === 'datetime-local') {
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

// ── Inline form validation ─────────────────────────────────────────────────
function validateInput(input) {
  const field = input.closest('.field');
  if (!field) return;
  // Don't validate untouched pristine fields
  if (!input._touched && input.validity.valid) return;
  input._touched = true;

  let errorEl = field.querySelector('.field-error');
  if (!errorEl) {
    errorEl = document.createElement('div');
    errorEl.className = 'field-error';
    field.appendChild(errorEl);
  }
  if (!input.validity.valid) {
    field.classList.add('field-invalid');
    field.classList.remove('field-valid');
    errorEl.textContent = input.validationMessage;
  } else {
    field.classList.remove('field-invalid');
    field.classList.add('field-valid');
    errorEl.textContent = '';
  }
}

// ── Global event delegation ────────────────────────────────────────────────
// Ripple on every .btn click
document.addEventListener('click', (e) => {
  const btn = e.target.closest('.btn');
  if (!btn || btn.disabled) return;
  const rect = btn.getBoundingClientRect();
  const size = Math.max(rect.width, rect.height) * 2;
  const ripple = document.createElement('span');
  ripple.className = 'ripple';
  ripple.style.cssText = [
    `width:${size}px`,
    `height:${size}px`,
    `left:${e.clientX - rect.left - size / 2}px`,
    `top:${e.clientY - rect.top - size / 2}px`,
  ].join(';');
  btn.appendChild(ripple);
  ripple.addEventListener('animationend', () => ripple.remove(), { once: true });
}, { passive: true });

// Blur validation on form inputs
document.addEventListener('blur', (e) => {
  const el = e.target;
  if (el.matches('.input, .select, .textarea') && el.closest('form')) {
    validateInput(el);
  }
}, true);

// Clear validation error as user types (once touched)
document.addEventListener('input', (e) => {
  const el = e.target;
  if (el.matches('.input, .select, .textarea') && el._touched) {
    validateInput(el);
  }
}, { passive: true });
