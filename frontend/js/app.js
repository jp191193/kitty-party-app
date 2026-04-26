// Bootstraps the app: registers routes, paints the user chip, wires header buttons.
import { registerRoute, startRouter, currentPath, navigate, onBeforeRender } from './router.js';
import { store } from './store.js';
import { api } from './api.js';
import { toast, openModal, html, escape, initials, readForm } from './ui.js';
import { API_BASE, STORAGE_KEYS } from './config.js';

import { renderDashboard } from './pages/dashboard.js';
import { renderMembers, renderMemberDetail } from './pages/members.js';
import { renderGroups, renderGroupDetail } from './pages/groups.js';
import { renderContributions } from './pages/contributions.js';
import { renderCycles } from './pages/cycles.js';
import { renderPayouts } from './pages/payouts.js';

// ── Routes ────────────────────────────────────────────────────────────────
registerRoute('/',                     renderDashboard);
registerRoute('/members',              renderMembers);
registerRoute('/members/:id',          renderMemberDetail);
registerRoute('/groups',               renderGroups);
registerRoute('/groups/:id',           renderGroupDetail);
registerRoute('/contributions',        renderContributions);
registerRoute('/cycles',               renderCycles);
registerRoute('/payouts',              renderPayouts);

// Highlight active nav link on every render
onBeforeRender(({ path }) => {
  const root = '/' + (path.split('/').filter(Boolean)[0] || '');
  document.querySelectorAll('#mainNav a[data-route]').forEach((a) => {
    a.classList.toggle('active', a.dataset.route === root);
  });
});

// ── Header user chip ──────────────────────────────────────────────────────
function paintUserChip() {
  const label = document.getElementById('currentUserLabel');
  const avatar = document.getElementById('currentUserAvatar');
  if (!label || !avatar) return;
  if (store.userName) {
    label.textContent = store.userName;
    avatar.textContent = initials(store.userName);
  } else {
    label.textContent = 'No user selected';
    avatar.textContent = '?';
  }
}
store.subscribe(paintUserChip);
paintUserChip();

// ── Switch user dialog ────────────────────────────────────────────────────
async function openSwitchUser() {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <p class="text-muted" style="margin-top:0">
      Some endpoints (creating a profile, adding members to a group) require an authenticated user.
      Pick an existing member to act as — we'll request a JWT for you.
    </p>
    <div class="loader"><div class="spinner"></div><p>Loading members…</p></div>
  `;
  const m = openModal({ title: 'Switch acting user', body: wrap });
  let members = [];
  try {
    members = await api.members.list() || [];
  } catch (e) {
    wrap.innerHTML = `<p class="text-danger">Could not load members: ${escape(e.message)}</p>`;
    return;
  }
  if (!members.length) {
    wrap.innerHTML = `
      <p>No members exist yet. Create one first from the <a href="#/members" data-close>Members</a> page.</p>`;
    wrap.querySelector('[data-close]').onclick = () => m.close();
    return;
  }

  wrap.innerHTML = `
    <div class="field">
      <label for="userPicker">Member</label>
      <select id="userPicker" class="select">
        ${members.map((u) => `<option value="${escape(u.id)}" data-name="${escape(u.name)}">${escape(u.name)} — ${escape(u.email || '')}</option>`).join('')}
      </select>
      <div class="hint">A bearer token will be issued and stored in your browser's localStorage.</div>
    </div>
    <div class="form-actions">
      <button class="btn btn-secondary" data-act="clear">Clear</button>
      <button class="btn" data-act="ok">Use this user</button>
    </div>
  `;
  wrap.querySelector('[data-act="clear"]').onclick = () => { store.clearUser(); m.close(); toast('User cleared', { type: 'info' }); };
  wrap.querySelector('[data-act="ok"]').onclick = async () => {
    const sel = wrap.querySelector('#userPicker');
    const id = sel.value;
    const name = sel.options[sel.selectedIndex].dataset.name;
    try {
      const res = await api.auth.token(id);
      store.setUser({ id, name, token: res.token });
      m.close();
      toast(`Acting as ${name}`, { type: 'success' });
    } catch (e) {
      toast(e.message, { type: 'error', title: 'Token request failed' });
    }
  };
}
document.getElementById('switchUserBtn').addEventListener('click', openSwitchUser);

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
