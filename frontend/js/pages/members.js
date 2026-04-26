// Members listing + individual member detail (with profile + summary).
import { api } from '../api.js';
import { store } from '../store.js';
import {
  escape, openModal, toast, confirm, readForm,
  loadingHtml, errorHtml, fmtDate, fmtMoney, initials, statusBadge, animateRows,
} from '../ui.js';

export async function renderMembers({ root }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Members</h1>
        <div class="subtitle">People who can join kitty groups.</div>
      </div>
      <div class="page-actions">
        <button class="btn" id="newMemberBtn">+ Add member</button>
      </div>
    </div>
    <div class="toolbar">
      <input class="input search" id="memberSearch" placeholder="Search by name, email or phone…" />
    </div>
    <div class="card"><div id="membersTable">${loadingHtml()}</div></div>
  `;

  const container = root.querySelector('#membersTable');
  const search = root.querySelector('#memberSearch');
  let members = [];

  async function reload() {
    container.innerHTML = loadingHtml();
    try {
      members = (await api.members.list()) || [];
      paint();
    } catch (e) { container.innerHTML = errorHtml(e.message); }
  }

  function paint() {
    const q = search.value.trim().toLowerCase();
    const filtered = q
      ? members.filter((m) =>
          (m.name || '').toLowerCase().includes(q) ||
          (m.email || '').toLowerCase().includes(q) ||
          (m.phone || '').toLowerCase().includes(q))
      : members;

    if (!filtered.length) {
      container.innerHTML = `<div class="empty">
        <div class="empty-title">${members.length ? 'No matches' : 'No members yet'}</div>
        <div>${members.length ? 'Try a different search term.' : 'Click "Add member" to get started.'}</div>
      </div>`;
      return;
    }

    container.innerHTML = `
      <div class="table-wrap"><table class="data-table">
        <thead><tr>
          <th></th><th>Name</th><th>Email</th><th>Phone</th><th>Joined</th><th></th>
        </tr></thead>
        <tbody>
          ${filtered.map((m) => `
            <tr>
              <td><span class="avatar">${escape(initials(m.name))}</span></td>
              <td><a href="#/members/${escape(m.id)}"><strong>${escape(m.name)}</strong></a><div class="text-muted mono">${escape(m.id.slice(0, 8))}</div></td>
              <td>${escape(m.email)}</td>
              <td>${escape(m.phone)}</td>
              <td class="text-muted">${fmtDate(m.created_at)}</td>
              <td class="actions">
                <a class="btn btn-secondary btn-sm" href="#/members/${escape(m.id)}">Open</a>
                <button class="btn btn-secondary btn-sm" data-edit="${escape(m.id)}">Edit</button>
                <button class="btn btn-danger btn-sm" data-del="${escape(m.id)}">Delete</button>
              </td>
            </tr>`).join('')}
        </tbody></table></div>`;

    const tbody = container.querySelector('tbody');
    animateRows(tbody);

    tbody.querySelectorAll('[data-edit]').forEach((b) =>
      b.addEventListener('click', () => openMemberForm(members.find((m) => m.id === b.dataset.edit), reload)));
    tbody.querySelectorAll('[data-del]').forEach((b) =>
      b.addEventListener('click', async () => {
        const m = members.find((x) => x.id === b.dataset.del);
        if (!await confirm({ title: 'Delete member?', message: `Permanently delete "${m.name}"?`, confirmText: 'Delete', danger: true })) return;
        try {
          await api.members.remove(m.id);
          toast('Member deleted', { type: 'success' });
          reload();
        } catch (e) { toast(e.message, { type: 'error', title: 'Delete failed' }); }
      }));
  }

  search.addEventListener('input', paint);
  root.querySelector('#newMemberBtn').addEventListener('click', () => openMemberForm(null, reload));

  reload();
}

// ── Member form (create / edit) ───────────────────────────────────────────
function openMemberForm(member, onSaved) {
  const isEdit = !!member;
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <form id="memberForm">
      <div class="form-grid">
        <div class="field full">
          <label for="name">Full name</label>
          <input class="input" id="name" name="name" required minlength="2" maxlength="100"
                 value="${escape(member?.name || '')}" placeholder="e.g. Anjali Sharma" />
        </div>
        <div class="field">
          <label for="email">Email</label>
          <input class="input" id="email" name="email" type="email" required
                 value="${escape(member?.email || '')}" placeholder="anjali@example.com" />
        </div>
        <div class="field">
          <label for="phone">Phone</label>
          <input class="input" id="phone" name="phone" required minlength="10" maxlength="15"
                 value="${escape(member?.phone || '')}" placeholder="9876543210" />
        </div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">${isEdit ? 'Save changes' : 'Create member'}</button>
      </div>
    </form>`;
  const m = openModal({ title: isEdit ? 'Edit member' : 'Add a new member', body: wrap });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();
  wrap.querySelector('#memberForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      if (isEdit) await api.members.update(member.id, data);
      else await api.members.create(data);
      toast(isEdit ? 'Member updated' : 'Member created', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Save failed' });
    } finally { btn.disabled = false; }
  };
}

// ── Member detail page ────────────────────────────────────────────────────
export async function renderMemberDetail({ root, params }) {
  const id = params.id;
  root.innerHTML = `
    <div class="page-header">
      <div>
        <a href="#/members" class="text-muted">← All members</a>
        <h1 id="memberName">Member</h1>
        <div class="subtitle" id="memberMeta"></div>
      </div>
      <div class="page-actions" id="memberActions"></div>
    </div>
    <div id="memberBody">${loadingHtml()}</div>
  `;
  try {
    const summary = await api.profiles.summary(id);
    paintMemberDetail(root, summary, id);
  } catch (e) {
    root.querySelector('#memberBody').innerHTML = errorHtml(e.message);
  }
}

function paintMemberDetail(root, s, id) {
  document.getElementById('memberName').textContent = s.name || 'Member';
  document.getElementById('memberMeta').innerHTML =
    `<span>${escape(s.email || '')}</span> · <span>${escape(s.phone || '')}</span>`;
  document.getElementById('memberActions').innerHTML = `
    <button class="btn btn-secondary" id="editProfileBtn">Edit profile</button>
  `;
  document.getElementById('editProfileBtn').onclick = () => openProfileForm(id, s.profile, () => renderMemberDetail({ root, params: { id } }));

  const p = s.profile || {};
  const groupRows = (s.groups || []).map((g) => `
    <tr>
      <td><a href="#/groups/${escape(g.group_id)}">${escape(g.group_name)}</a></td>
      <td>${statusBadge(g.role)}</td>
      <td>${statusBadge(g.status)}</td>
      <td class="text-muted">${fmtDate(g.joined_at)}</td>
    </tr>`).join('');

  root.querySelector('#memberBody').innerHTML = `
    <div class="grid grid-cols-3" style="margin-bottom:20px">
      ${[
        ['Groups', s.group_count ?? 0, 'Active memberships'],
        ['Lifetime paid', fmtMoney(s.total_paid), 'Across all groups'],
        ['Pending dues', s.pending_dues ?? 0, 'Outstanding contributions'],
      ].map(([l, v, sub]) => `
        <div class="stat"><div class="stat-label">${escape(l)}</div>
          <div class="stat-value">${typeof v === 'string' ? escape(v) : v}</div>
          <div class="stat-sub">${escape(sub)}</div></div>`).join('')}
    </div>
    <div class="grid grid-cols-2">
      <div class="card">
        <div class="card-header"><h3>Profile details</h3></div>
        <div class="card-body">
          <dl style="display:grid;grid-template-columns:140px 1fr;gap:8px 16px;margin:0">
            ${profileRow('Date of birth', p.date_of_birth ? fmtDate(p.date_of_birth) : '—')}
            ${profileRow('Occupation', p.occupation)}
            ${profileRow('Address', p.address)}
            ${profileRow('City', p.city)}
            ${profileRow('State', p.state)}
            ${profileRow('Pincode', p.pincode)}
            ${profileRow('About', p.about)}
          </dl>
        </div>
      </div>
      <div class="card">
        <div class="card-header"><h3>Groups</h3></div>
        ${groupRows
          ? `<div class="table-wrap"><table class="data-table">
              <thead><tr><th>Group</th><th>Role</th><th>Status</th><th>Joined</th></tr></thead>
              <tbody>${groupRows}</tbody></table></div>`
          : `<div class="empty"><div class="empty-title">Not in any group yet</div><div>Add to a group from the Groups page.</div></div>`}
      </div>
    </div>
  `;
}

function profileRow(label, value) {
  return `<dt class="text-muted">${escape(label)}</dt><dd style="margin:0">${value ? escape(value) : '<span class="text-muted">—</span>'}</dd>`;
}

function openProfileForm(memberId, profile, onSaved) {
  if (!store.token) {
    toast('Click "Switch" in the header and pick this user to get an auth token.', { type: 'warning', title: 'Login required', duration: 5000 });
    return;
  }
  const p = profile || {};
  const dob = p.date_of_birth ? p.date_of_birth.slice(0, 10) : '';
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <form id="profileForm">
      <div class="form-grid">
        <div class="field"><label>Date of birth</label>
          <input class="input" name="date_of_birth" type="date" value="${escape(dob)}" /></div>
        <div class="field"><label>Occupation</label>
          <input class="input" name="occupation" maxlength="100" value="${escape(p.occupation || '')}" /></div>
        <div class="field full"><label>Profile picture URL</label>
          <input class="input" name="profile_picture_url" maxlength="2048" value="${escape(p.profile_picture_url || '')}" /></div>
        <div class="field full"><label>Address</label>
          <input class="input" name="address" maxlength="255" value="${escape(p.address || '')}" /></div>
        <div class="field"><label>City</label>
          <input class="input" name="city" maxlength="100" value="${escape(p.city || '')}" /></div>
        <div class="field"><label>State</label>
          <input class="input" name="state" maxlength="100" value="${escape(p.state || '')}" /></div>
        <div class="field"><label>Pincode</label>
          <input class="input" name="pincode" maxlength="10" value="${escape(p.pincode || '')}" /></div>
        <div class="field full"><label>About</label>
          <textarea class="textarea" name="about" maxlength="500">${escape(p.about || '')}</textarea></div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Save profile</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Edit profile', body: wrap, size: 'lg' });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();
  wrap.querySelector('#profileForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    // Convert YYYY-MM-DD: backend accepts the literal "YYYY-MM-DD" string.
    if (data.date_of_birth instanceof Date) data.date_of_birth = data.date_of_birth.toISOString().slice(0, 10);
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      await api.profiles.upsert(memberId, data);
      toast('Profile updated', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Save failed' });
    } finally { btn.disabled = false; }
  };
}
