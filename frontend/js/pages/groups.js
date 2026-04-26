// Groups list + group detail (with members panel & quick links to other domains).
import { api } from '../api.js';
import { store } from '../store.js';
import {
  escape, openModal, toast, confirm, readForm,
  loadingHtml, errorHtml, fmtDate, fmtMoney, statusBadge, animateRows,
} from '../ui.js';

export async function renderGroups({ root }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Groups</h1>
        <div class="subtitle">Each group runs a single kitty pool over a fixed number of months.</div>
      </div>
      <div class="page-actions">
        <button class="btn" id="newGroupBtn">+ New group</button>
      </div>
    </div>
    <div class="toolbar">
      <input class="input search" id="groupSearch" placeholder="Search groups…" />
    </div>
    <div class="card"><div id="groupsTable">${loadingHtml()}</div></div>
  `;

  const container = root.querySelector('#groupsTable');
  const search = root.querySelector('#groupSearch');
  let groups = [], counts = {};

  async function reload() {
    container.innerHTML = loadingHtml();
    try {
      const [g, c] = await Promise.all([api.groups.list(), api.groups.memberCounts().catch(() => [])]);
      groups = g || [];
      counts = Object.fromEntries((c || []).map((x) => [x.id, x.count]));
      paint();
    } catch (e) { container.innerHTML = errorHtml(e.message); }
  }

  function paint() {
    const q = search.value.trim().toLowerCase();
    const filtered = q ? groups.filter((g) => (g.name || '').toLowerCase().includes(q)) : groups;
    if (!filtered.length) {
      container.innerHTML = `<div class="empty">
        <div class="empty-title">${groups.length ? 'No matches' : 'No groups yet'}</div>
        <div>${groups.length ? 'Try a different search.' : 'Create your first group to start collecting contributions.'}</div>
      </div>`;
      return;
    }
    container.innerHTML = `
      <div class="table-wrap"><table class="data-table">
        <thead><tr>
          <th>Name</th><th>Monthly</th><th>Duration</th><th>Members</th>
          <th>Pool size</th><th>Start date</th><th></th>
        </tr></thead>
        <tbody>${filtered.map((g) => {
          const members = counts[g.id] ?? '—';
          const pool = (g.monthly_amount || 0) * (g.duration || 0);
          return `<tr>
            <td><a href="#/groups/${escape(g.id)}"><strong>${escape(g.name)}</strong></a>
                <div class="text-muted mono">${escape(g.id.slice(0, 8))}</div></td>
            <td>${fmtMoney(g.monthly_amount)}</td>
            <td>${escape(g.duration)} months</td>
            <td>${escape(members)}</td>
            <td>${fmtMoney(pool)}</td>
            <td class="text-muted">${fmtDate(g.start_date)}</td>
            <td class="actions">
              <a class="btn btn-secondary btn-sm" href="#/groups/${escape(g.id)}">Open</a>
              <button class="btn btn-secondary btn-sm" data-edit="${escape(g.id)}">Edit</button>
              <button class="btn btn-danger btn-sm" data-del="${escape(g.id)}">Delete</button>
            </td>
          </tr>`;
        }).join('')}</tbody></table></div>`;

    const tbody = container.querySelector('tbody');
    animateRows(tbody);
    tbody.querySelectorAll('[data-edit]').forEach((b) =>
      b.addEventListener('click', () => openGroupForm(groups.find((g) => g.id === b.dataset.edit), reload)));
    tbody.querySelectorAll('[data-del]').forEach((b) =>
      b.addEventListener('click', async () => {
        const g = groups.find((x) => x.id === b.dataset.del);
        if (!await confirm({ title: 'Delete group?', message: `Permanently delete "${g.name}"?`, confirmText: 'Delete', danger: true })) return;
        try { await api.groups.remove(g.id); toast('Group deleted', { type: 'success' }); reload(); }
        catch (e) { toast(e.message, { type: 'error', title: 'Delete failed' }); }
      }));
  }

  search.addEventListener('input', paint);
  root.querySelector('#newGroupBtn').addEventListener('click', () => openGroupForm(null, reload));
  reload();
}

function openGroupForm(group, onSaved) {
  const isEdit = !!group;
  const wrap = document.createElement('div');

  // For create we need a creator member id. Try the acting user first; otherwise let them pick.
  wrap.innerHTML = `
    <form id="groupForm">
      <div class="form-grid">
        <div class="field full"><label>Group name</label>
          <input class="input" name="name" required minlength="2" maxlength="100"
                 value="${escape(group?.name || '')}" placeholder="e.g. Diwali Bachat 2026" /></div>
        <div class="field"><label>Monthly amount (₹)</label>
          <input class="input" name="monthly_amount" type="number" min="1" step="1" required
                 value="${escape(group?.monthly_amount || '')}" /></div>
        <div class="field"><label>Duration (months)</label>
          <input class="input" name="duration" type="number" min="1" step="1" required
                 value="${escape(group?.duration || '')}" /></div>
        <div class="field full"><label>Start date</label>
          <input class="input" name="start_date" type="datetime-local" required
                 value="${toDtLocal(group?.start_date)}" /></div>
        ${isEdit ? '' : `<div class="field full" id="creatorField">
            <label>Created by</label>
            <select class="select" name="created_by" required id="creatorSelect">
              <option value="">Loading members…</option>
            </select>
            <div class="hint">The member who is creating this group.</div></div>`}
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">${isEdit ? 'Save changes' : 'Create group'}</button>
      </div>
    </form>`;
  const m = openModal({ title: isEdit ? 'Edit group' : 'Create new group', body: wrap });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();

  if (!isEdit) {
    api.members.list().then((members) => {
      const sel = wrap.querySelector('#creatorSelect');
      sel.innerHTML = `<option value="">— select member —</option>` +
        (members || []).map((u) => `<option value="${escape(u.id)}" ${u.id === store.userId ? 'selected' : ''}>${escape(u.name)}</option>`).join('');
    }).catch(() => {});
  }

  wrap.querySelector('#groupForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      if (isEdit) await api.groups.update(group.id, data);
      else await api.groups.create(data);
      toast(isEdit ? 'Group updated' : 'Group created', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Save failed' });
    } finally { btn.disabled = false; }
  };
}

function toDtLocal(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (isNaN(d)) return '';
  // Convert to local YYYY-MM-DDTHH:MM expected by datetime-local
  const tz = d.getTimezoneOffset();
  const local = new Date(d.getTime() - tz * 60000);
  return local.toISOString().slice(0, 16);
}

// ── Group detail ──────────────────────────────────────────────────────────
export async function renderGroupDetail({ root, params }) {
  const id = params.id;
  root.innerHTML = `
    <div class="page-header">
      <div>
        <a href="#/groups" class="text-muted">← All groups</a>
        <h1 id="groupName">Group</h1>
        <div class="subtitle" id="groupMeta"></div>
      </div>
      <div class="page-actions" id="groupActions"></div>
    </div>
    <div id="groupBody">${loadingHtml()}</div>
  `;

  let group, members = [];
  try {
    [group, members] = await Promise.all([
      api.groups.get(id),
      api.groups.members.list(id).catch(() => []),
    ]);
  } catch (e) {
    root.querySelector('#groupBody').innerHTML = errorHtml(e.message);
    return;
  }

  document.getElementById('groupName').textContent = group.name;
  document.getElementById('groupMeta').innerHTML =
    `<span>${fmtMoney(group.monthly_amount)}/month</span> · <span>${escape(group.duration)} months</span> · <span>Starts ${fmtDate(group.start_date)}</span>`;
  document.getElementById('groupActions').innerHTML = `
    <a class="btn btn-secondary" href="#/contributions?group=${escape(id)}">Contributions</a>
    <a class="btn btn-secondary" href="#/cycles?group=${escape(id)}">Cycles</a>
    <a class="btn btn-secondary" href="#/payouts?group=${escape(id)}">Payouts</a>
    <button class="btn" id="addMemberBtn">+ Add member</button>
  `;

  const pool = (group.monthly_amount || 0) * (group.duration || 0);
  root.querySelector('#groupBody').innerHTML = `
    <div class="grid grid-cols-4" style="margin-bottom:20px">
      ${stat('Members', members.length)}
      ${stat('Monthly amount', fmtMoney(group.monthly_amount))}
      ${stat('Duration', `${group.duration} months`)}
      ${stat('Total pool', fmtMoney(pool))}
    </div>
    <div class="card">
      <div class="card-header"><h3>Members in this group</h3></div>
      ${renderMembersTable(members)}
    </div>
  `;

  document.getElementById('addMemberBtn').onclick = () => openAddMember(id, () => renderGroupDetail({ root, params }));
}

function stat(label, value) {
  return `<div class="stat">
    <div class="stat-label">${escape(label)}</div>
    <div class="stat-value">${typeof value === 'string' ? escape(value) : value}</div>
  </div>`;
}

function renderMembersTable(members) {
  if (!members.length) {
    return `<div class="empty">
      <div class="empty-title">No members yet</div>
      <div>Click "Add member" to enrol someone in this group.</div>
    </div>`;
  }
  return `<div class="table-wrap"><table class="data-table">
    <thead><tr><th>Member</th><th>Role</th><th>Status</th><th>Joined</th></tr></thead>
    <tbody>${members.map((m) => `
      <tr>
        <td><a href="#/members/${escape(m.member_id)}">${escape(m.member_name || m.member_id)}</a></td>
        <td>${statusBadge(m.role)}</td>
        <td>${statusBadge(m.status)}</td>
        <td class="text-muted">${fmtDate(m.joined_at)}</td>
      </tr>`).join('')}
    </tbody></table></div>`;
}

function openAddMember(groupId, onSaved) {
  if (!store.token) {
    toast('Switch to the group creator (header → Switch) before adding members.', { type: 'warning', title: 'Login required', duration: 5000 });
    return;
  }
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <form id="addMemberForm">
      <div class="field"><label>Member to add</label>
        <select class="select" name="member_id" required>
          <option value="">Loading members…</option>
        </select>
        <div class="hint">Only the group creator can add members.</div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Add member</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Add member to group', body: wrap });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();
  api.members.list().then((members) => {
    const sel = wrap.querySelector('select[name="member_id"]');
    sel.innerHTML = `<option value="">— select member —</option>` +
      (members || []).map((u) => `<option value="${escape(u.id)}">${escape(u.name)} — ${escape(u.email || '')}</option>`).join('');
  });
  wrap.querySelector('#addMemberForm').onsubmit = async (e) => {
    e.preventDefault();
    const { member_id } = readForm(e.target);
    if (!member_id) return;
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      await api.groups.members.add(groupId, member_id);
      toast('Member added', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Add failed' });
    } finally { btn.disabled = false; }
  };
}
