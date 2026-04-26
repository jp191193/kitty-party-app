// Contributions page: pick a group, view dues, generate dues, record payments.
import { api } from '../api.js';
import {
  escape, openModal, toast, readForm,
  loadingHtml, errorHtml, fmtDate, fmtMoney, statusBadge, animateRows,
} from '../ui.js';

export async function renderContributions({ root, query }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Contributions</h1>
        <div class="subtitle">Generate monthly dues for a group, then record member payments.</div>
      </div>
      <div class="page-actions" id="topActions"></div>
    </div>
    <div class="toolbar">
      <div class="field" style="flex:1;min-width:240px">
        <label>Group</label>
        <select class="select" id="groupPicker"><option value="">Loading groups…</option></select>
      </div>
    </div>
    <div class="card"><div id="duesContainer">${emptyHint('Pick a group to view its dues.')}</div></div>
  `;

  const picker = root.querySelector('#groupPicker');
  const container = root.querySelector('#duesContainer');
  const topActions = root.querySelector('#topActions');

  let groups = [];
  try { groups = (await api.groups.list()) || []; }
  catch (e) { picker.innerHTML = `<option>Failed: ${escape(e.message)}</option>`; return; }

  if (!groups.length) {
    container.innerHTML = `<div class="empty">
      <div class="empty-title">No groups yet</div>
      <div>Create a group on the <a href="#/groups">Groups page</a> first.</div>
    </div>`;
    return;
  }

  picker.innerHTML = `<option value="">— select a group —</option>` +
    groups.map((g) => `<option value="${escape(g.id)}">${escape(g.name)} · ${fmtMoney(g.monthly_amount)}/mo</option>`).join('');

  async function loadFor(groupId) {
    if (!groupId) { container.innerHTML = emptyHint('Pick a group to view its dues.'); topActions.innerHTML = ''; return; }
    const g = groups.find((x) => x.id === groupId);
    topActions.innerHTML = `<button class="btn" id="genDuesBtn">+ Generate dues</button>`;
    document.getElementById('genDuesBtn').onclick = () => openGenerateDues(g, () => loadFor(groupId));

    container.innerHTML = loadingHtml('Loading dues…');
    try {
      const [dues, members] = await Promise.all([
        api.contributions.listDuesByGroup(groupId),
        api.groups.members.list(groupId),
      ]);
      const memberMap = Object.fromEntries((members || []).map((m) => [m.member_id, m.member_name || m.member_id]));
      paintDues(container, dues || [], g, memberMap);
    } catch (e) { container.innerHTML = errorHtml(e.message); }
  }

  picker.addEventListener('change', () => loadFor(picker.value));

  // Honour ?group=... from URL
  if (query?.group) {
    picker.value = query.group;
    loadFor(query.group);
  }
}

function paintDues(container, dues, group, memberMap = {}) {
  if (!dues.length) {
    container.innerHTML = `<div class="empty">
      <div class="empty-title">No dues generated yet</div>
      <div>Click "Generate dues" to create monthly contribution records for this group.</div>
    </div>`;
    return;
  }

  // Group by cycle for easier reading
  const byCycle = new Map();
  dues.forEach((d) => {
    const k = `${d.cycle_year}-${String(d.cycle_month).padStart(2, '0')}-c${d.cycle_number}`;
    if (!byCycle.has(k)) byCycle.set(k, []);
    byCycle.get(k).push(d);
  });

  const sections = [...byCycle.entries()]
    .sort((a, b) => a[0] < b[0] ? 1 : -1)
    .map(([k, items]) => {
      const first = items[0];
      const totalDue = items.reduce((s, d) => s + (Number(d.amount) || 0), 0);
      const paid = items.filter((d) => d.status === 'PAID').length;
      return `
        <div class="card" style="margin-bottom:14px">
          <div class="card-header">
            <div>
              <h3 style="margin:0">Cycle ${escape(first.cycle_number)} · ${monthName(first.cycle_month)} ${escape(first.cycle_year)}</h3>
              <div class="subtitle text-muted">Kitty: ${fmtDate(first.kitty_party_date)} · Due: ${fmtDate(first.due_date)} · Pool ${fmtMoney(totalDue)}</div>
            </div>
            <div class="text-muted">${paid}/${items.length} paid</div>
          </div>
          <div class="table-wrap"><table class="data-table">
            <thead><tr>
              <th>Member</th><th>Amount</th><th>Status</th><th>Updated</th><th></th>
            </tr></thead>
            <tbody>${items.map((d) => `
              <tr>
                <td><a href="#/members/${escape(d.member_id)}">${escape(memberMap[d.member_id] || d.member_id.slice(0, 8))}</a></td>
                <td>${fmtMoney(d.amount)}</td>
                <td>${statusBadge(d.status)}</td>
                <td class="text-muted">${fmtDate(d.updated_at || d.created_at)}</td>
                <td class="actions">
                  ${d.status === 'PAID' ? '' : `<button class="btn btn-sm" data-pay="${escape(d.id)}" data-amount="${escape(d.amount)}">Record payment</button>`}
                </td>
              </tr>`).join('')}
            </tbody></table></div>
        </div>`;
    }).join('');

  container.innerHTML = sections;
  container.querySelectorAll('[data-pay]').forEach((btn) =>
    btn.addEventListener('click', () => openRecordPayment(
      btn.dataset.pay, Number(btn.dataset.amount),
      () => paintAfterPayment(container, group),
    )));
  container.querySelectorAll('table tbody').forEach((tb) => animateRows(tb));
}

async function paintAfterPayment(container, group) {
  try {
    const [dues, members] = await Promise.all([
      api.contributions.listDuesByGroup(group.id),
      api.groups.members.list(group.id),
    ]);
    const memberMap = Object.fromEntries((members || []).map((m) => [m.member_id, m.member_name || m.member_id]));
    paintDues(container, dues || [], group, memberMap);
  } catch (e) { container.innerHTML = errorHtml(e.message); }
}

function openGenerateDues(group, onSaved) {
  const wrap = document.createElement('div');
  const today = new Date();
  const due = new Date(today.getTime() + 14 * 86400000);
  wrap.innerHTML = `
    <p class="text-muted" style="margin-top:0">
      Creates a contribution due record for every member of <strong>${escape(group.name)}</strong>
      for the cycle below. Pick the host member who receives the pool this cycle.
    </p>
    <form id="genDuesForm">
      <div class="form-grid">
        <div class="field full"><label>Host member</label>
          <select class="select" name="host_member_id" required id="hostSel"><option>Loading…</option></select>
          <div class="hint">The member who hosts this cycle and receives the pool.</div></div>
        <div class="field"><label>Cycle number</label>
          <input class="input" type="number" name="cycle_number" min="1" required value="1" /></div>
        <div class="field"><label>Cycle month</label>
          <input class="input" type="number" name="cycle_month" min="1" max="12" required value="${today.getMonth() + 1}" /></div>
        <div class="field"><label>Cycle year</label>
          <input class="input" type="number" name="cycle_year" min="2000" required value="${today.getFullYear()}" /></div>
        <div class="field"><label>Amount per member (₹)</label>
          <input class="input" type="number" name="amount" min="1" step="1" required value="${escape(group.monthly_amount)}" /></div>
        <div class="field"><label>Kitty party date</label>
          <input class="input" type="datetime-local" name="kitty_party_date" required value="${toDtLocal(today)}" /></div>
        <div class="field"><label>Due date</label>
          <input class="input" type="datetime-local" name="due_date" required value="${toDtLocal(due)}" /></div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Generate dues</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Generate monthly dues', body: wrap, size: 'lg' });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();

  api.groups.members.list(group.id).then((mem) => {
    const sel = wrap.querySelector('#hostSel');
    sel.innerHTML = `<option value="">— select host —</option>` +
      (mem || []).map((m) => `<option value="${escape(m.member_id)}">${escape(m.member_name || m.member_id)}</option>`).join('');
  }).catch(() => {});

  wrap.querySelector('#genDuesForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    data.group_id = group.id;
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      await api.contributions.generateDues(data);
      toast('Dues generated', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Generate failed' });
    } finally { btn.disabled = false; }
  };
}

function openRecordPayment(dueId, amount, onSaved) {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <form id="payForm">
      <div class="form-grid">
        <div class="field"><label>Amount paid (₹)</label>
          <input class="input" type="number" name="amount_paid" min="1" step="1" required value="${escape(amount || '')}" /></div>
        <div class="field"><label>Payment mode</label>
          <select class="select" name="payment_mode" required>
            <option value="UPI">UPI</option>
            <option value="CASH">Cash</option>
            <option value="BANK">Bank transfer</option>
          </select></div>
        <div class="field full"><label>Transaction reference</label>
          <input class="input" name="transaction_ref" placeholder="UPI ref / cheque no. (optional)" /></div>
        <div class="field"><label>Status</label>
          <select class="select" name="status" required>
            <option value="SUCCESS">Success</option>
            <option value="FAILED">Failed</option>
          </select></div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Record payment</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Record payment', body: wrap });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();
  wrap.querySelector('#payForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    data.due_id = dueId;
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      await api.contributions.recordPayment(data);
      toast('Payment recorded', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Save failed' });
    } finally { btn.disabled = false; }
  };
}

function emptyHint(msg) { return `<div class="empty"><div>${escape(msg)}</div></div>`; }
function monthName(m) { return ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'][((+m || 1) - 1)] || m; }
function toDtLocal(d) {
  if (!d) return '';
  const dt = d instanceof Date ? d : new Date(d);
  if (isNaN(dt)) return '';
  const tz = dt.getTimezoneOffset();
  return new Date(dt.getTime() - tz * 60000).toISOString().slice(0, 16);
}
