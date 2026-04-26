import { registerRoute, startRouter, currentPath, navigate, onBeforeRender } from './router.js';
import { store } from './store.js';
import { api } from './api.js';
import { toast, openModal, html, escape, initials, readForm } from './ui.js';
import { API_BASE, STORAGE_KEYS } from './config.js';

import { renderLogin } from './pages/login.js';
import { renderDashboard } from './pages/dashboard.js';
import { renderMembers, renderMemberDetail } from './pages/members.js';
import { renderGroups, renderGroupDetail } from './pages/groups.js';
import { renderContributions } from './pages/contributions.js';
import { renderCycles } from './pages/cycles.js';
import { renderPayouts } from './pages/payouts.js';
import { renderInvitations } from './pages/invitations.js';

// ── Routes ────────────────────────────────────────────────────────────────
registerRoute('/login',              renderLogin);
registerRoute('/',                   renderDashboard);
registerRoute('/members',            renderMembers);
registerRoute('/members/:id',        renderMemberDetail);
registerRoute('/groups',             renderGroups);
registerRoute('/groups/:id',         renderGroupDetail);
registerRoute('/contributions',      renderContributions);
registerRoute('/cycles',             renderCycles);
registerRoute('/payouts',            renderPayouts);
registerRoute('/invitations',        renderInvitations);

// ── Route guard + nav highlight ───────────────────────────────────────────
onBeforeRender(({ path }) => {
  const isLoginPage = path === '/login';
  const isLoggedIn  = !!store.token;

  if (!isLoggedIn && !isLoginPage) {
    navigate('/login');
    return false;
  }
  if (isLoggedIn && isLoginPage) {
    navigate('/');
    return false;
  }

  // Highlight active nav link
  const root = '/' + (path.split('/').filter(Boolean)[0] || '');
  document.querySelectorAll('#mainNav a[data-route]').forEach((a) => {
    a.classList.toggle('active', a.dataset.route === root);
  });
});

// ── Header: reflect auth state ────────────────────────────────────────────
function updateHeader() {
  const isLoggedIn = !!store.token;
  document.getElementById('mainNav').style.display         = isLoggedIn ? '' : 'none';
  document.getElementById('currentUserChip').style.display = isLoggedIn ? '' : 'none';
  document.getElementById('logoutBtn').style.display       = isLoggedIn ? '' : 'none';
  document.getElementById('settingsBtn').style.display     = isLoggedIn ? '' : 'none';

  const label  = document.getElementById('currentUserLabel');
  const avatar = document.getElementById('currentUserAvatar');
  if (!label || !avatar) return;
  if (store.userName) {
    label.textContent  = store.userName;
    avatar.textContent = initials(store.userName);
  } else {
    label.textContent  = 'No user';
    avatar.textContent = '?';
  }
}
store.subscribe(updateHeader);
updateHeader();

// ── Logout ────────────────────────────────────────────────────────────────
document.getElementById('logoutBtn').addEventListener('click', async () => {
  try { await api.auth.logout(); } catch (_) { /* stateless — ignore errors */ }
  store.clearUser();
  navigate('/login');
});

// ── Global 401 handler ────────────────────────────────────────────────────
window.addEventListener('auth:unauthorized', () => {
  // Suppress if already on the login page (e.g. wrong-credentials 401)
  if (currentPath() === '/login') return;
  store.clearUser();
  navigate('/login');
  toast('Session expired. Please sign in again.', { type: 'error' });
});

// ── Settings dialog ───────────────────────────────────────────────────────
document.getElementById('settingsBtn').addEventListener('click', () => {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <form id="settingsForm">
      <div class="field full">
        <label for="apiBase">API base URL</label>
        <input class="input" id="apiBase" name="apiBase" value="${escape(API_BASE)}" />
        <div class="hint">Where the Go backend is running. Default: http://localhost:8080</div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="reset">Reset</button>
        <button type="submit" class="btn">Save & reload</button>
      </div>
    </form>
  `;
  const m = openModal({ title: 'Settings', body: wrap });
  wrap.querySelector('[data-act="reset"]').onclick = () => {
    localStorage.removeItem(STORAGE_KEYS.API_BASE);
    location.reload();
  };
  wrap.querySelector('#settingsForm').onsubmit = (e) => {
    e.preventDefault();
    const v = readForm(e.target).apiBase?.trim();
    if (!v) return;
    localStorage.setItem(STORAGE_KEYS.API_BASE, v.replace(/\/+$/, ''));
    location.reload();
  };
});

// Expose api/store for ad-hoc debugging from the console
window.kp = { api, store, navigate, currentPath };

// Boot
startRouter();
