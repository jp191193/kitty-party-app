// Payouts page: schedule a payout, list per group, disburse, cancel.
import { api } from '../api.js';
import { store } from '../store.js';
import {
  escape, openModal, toast, confirm, readForm,
  loadingHtml, errorHtml, fmtDate, fmtMoney, statusBadge, animateRows,
} from '../ui.js';

export async function renderPayouts({ root, query }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Payouts</h1>
        <div class="subtitle">Schedule when each member receives the pool, then record disbursement.</div>
      </div>
      <div class="page-actions" id="topActions"></div>
    </div>
    <div class="toolbar">
      <div class="field" style="flex:1;min-width:240px">
        <label>Group</label>
        <select class="select" id="groupPicker"><option value="">Loading groups…</option></select>
      </div>
    </div>
    <div class="card"><div id="payoutsContainer">${emptyHint('Pick a group to view payouts.')}</div></div>
  `;

  const picker = root.querySelector('#groupPicker');
  const container = root.querySelector('#payoutsContainer');
  const topActions = root.querySelector('#topActions');

  let groups = [];
  try { groups = (await api.groups.list()) || []; }
  catch (e) { picker.innerHTML = `<option>Failed: ${escape(e.message)}</option>`; return; }

  picker.innerHTML = `<option value="">— select a group —</option>` +
    groups.map((g) => `<option value="${escape(g.id)}">${escape(g.name)}</option>`).join('');

  async function loadFor(groupId) {
    if (!groupId) { container.innerHTML = emptyHint('Pick a group to view payouts.'); topActions.innerHTML = ''; return; }
    const g = groups.find((x) => x.id === groupId);
    topActions.innerHTML = `<button class="btn" id="newPayoutBtn">+ Schedule payout</button>`;
    document.getElementById('newPayoutBtn').onclick = () => openSchedulePayout(g, () => loadFor(groupId));

    container.innerHTML = loadingHtml('Loading payouts…');
    try {
      const payouts = (await api.payouts.listByGroup(groupId)) || [];
      paintPayouts(container, payouts, () => loadFor(groupId));
    } catch (e) { container.innerHTML = errorHtml(e.message); }
  }

  picker.addEventListener('change', () => loadFor(picker.value));
  if (query?.group) { picker.value = query.group; loadFor(query.group); }
}

function paintPayouts(container, payouts, reload) {
  if (!payouts.length) {
    container.innerHTML = `<div class="empty">
      <div class="empty-title">No payouts scheduled</div>
      <div>Click "Schedule payout" to plan one for this group.</div>
    </div>`;
    return;
  }
  const sorted = [...payouts].sort((a, b) => (a.cycle_number || 0) - (b.cycle_number || 0));
  container.innerHTML = `
    <div class="table-wrap"><table class="data-table">
      <thead><tr>
        <th>Cycle</th><th>Recipient</th><th>Scheduled</th><th>Pool</th>
        <th>Status</th><th>Disbursement</th><th></th>
      </tr></thead>
      <tbody>${sorted.map((p) => {
        const d = p.disbursement;
        return `<tr>
          <td><strong>#${escape(p.cycle_number)}</strong>
              <div class="text-muted">${escape(p.cycle_month)}/${escape(p.cycle_year)}</div></td>
          <td><a href="#/members/${escape(p.recipient_id)}" class="mono">${escape(p.recipient_id.slice(0, 8))}</a></td>
          <td class="text-muted">${fmtDate(p.scheduled_date)}</td>
          <td>${fmtMoney(p.pool_amount)}</td>
          <td>${statusBadge(p.status)}</td>
          <td>${d ? `${fmtMoney(d.amount_paid)} via ${statusBadge(d.payment_mode)}<div class="text-muted">${fmtDate(d.payment_date)}</div>` : '<span class="text-muted">—</span>'}</td>
          <td class="actions">
            ${p.status === 'SCHEDULED' ? `<button class="btn btn-sm" data-disburse="${escape(p.id)}" data-amount="${escape(p.pool_amount)}">Disburse</button>` : ''}
            ${p.status === 'SCHEDULED' ? `<button class="btn btn-sm btn-danger" data-cancel="${escape(p.id)}">Cancel</button>` : ''}
          </td>
        </tr>`;
      }).join('')}</tbody></table></div>`;
  const tbody = container.querySelector('tbody');
  animateRows(tbody);
  tbody.querySelectorAll('[data-disburse]').forEach((b) =>
    b.onclick = () => openDisburse(b.dataset.disburse, +b.dataset.amount, reload));
  tbody.querySelectorAll('[data-cancel]').forEach((b) =>
    b.onclick = async () => {
      if (!await confirm({ title: 'Cancel payout?', message: 'Cancel this scheduled payout?', confirmText: 'Cancel payout', danger: true })) return;
      try { await api.payouts.cancel(b.dataset.cancel, 'Cancelled from UI'); toast('Payout cancelled', { type: 'info' }); reload(); }
      catch (e) { toast(e.message, { type: 'error', title: 'Cancel failed' }); }
    });
}

function openSchedulePayout(group, onSaved) {
  const wrap = document.createElement('div');
  const today = new Date();
  wrap.innerHTML = `
    <form id="schedForm">
      <div class="form-grid">
        <div class="field full"><label>Recipient member</label>
          <select class="select" name="recipient_id" required id="recipSel"><option>Loading…</option></select></div>
        <div class="field"><label>Cycle number</label>
          <input class="input" type="number" name="cycle_number" min="1" required value="1" /></div>
        <div class="field"><label>Cycle month</label>
          <input class="input" type="number" name="cycle_month" min="1" max="12" required value="${today.getMonth() + 1}" /></div>
        <div class="field"><label>Cycle year</label>
          <input class="input" type="number" name="cycle_year" min="2000" required value="${today.getFullYear()}" /></div>
        <div class="field"><label>Pool amount (₹)</label>
          <input class="input" type="number" name="pool_amount" min="1" required value="${escape((group.monthly_amount || 0) * (group.duration || 0) || '')}" /></div>
        <div class="field"><label>Scheduled date</label>
          <input class="input" type="datetime-local" name="scheduled_date" required value="${toDtLocal(today)}" /></div>
        <div class="field full"><label>Notes</label>
          <textarea class="textarea" name="notes" placeholder="Optional"></textarea></div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Schedule payout</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Schedule a payout', body: wrap, size: 'lg' });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();

  api.groups.members.list(group.id).then((mem) => {
    const sel = wrap.querySelector('#recipSel');
    sel.innerHTML = `<option value="">— select recipient —</option>` +
      (mem || []).map((m) => `<option value="${escape(m.member_id)}">${escape(m.member_name || m.member_id)}</option>`).join('');
  });

  wrap.querySelector('#schedForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    data.group_id = group.id;
    if (store.userId) data.created_by = store.userId; // best-effort; backend may default
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      await api.payouts.schedule(data);
      toast('Payout scheduled', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) { toast(err.message, { type: 'error', title: 'Schedule failed' }); }
    finally { btn.disabled = false; }
  };
}

function openDisburse(payoutId, amount, onSaved) {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <form id="disburseForm">
      <div class="form-grid">
        <div class="field"><label>Amount paid (₹)</label>
          <input class="input" type="number" name="amount_paid" min="1" required value="${escape(amount || '')}" /></div>
        <div class="field"><label>Payment mode</label>
          <select class="select" name="payment_mode" required>
            <option value="UPI">UPI</option>
            <option value="CASH">Cash</option>
            <option value="BANK">Bank transfer</option>
          </select></div>
        <div class="field full"><label>Transaction reference</label>
          <input class="input" name="transaction_ref" placeholder="UPI ref / cheque no. (optional)" /></div>
        <div class="field full"><label>Notes</label>
          <textarea class="textarea" name="notes"></textarea></div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Record disbursement</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Record disbursement', body: wrap });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();
  wrap.querySelector('#disburseForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    data.payout_id = payoutId;
    const btn = e.target.querySelector('button[type=submit]');
    btn.disabled = true;
    try {
      await api.payouts.disburse(data);
      toast('Disbursement recorded', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) { toast(err.message, { type: 'error', title: 'Save failed' }); }
    finally { btn.disabled = false; }
  };
}

function emptyHint(msg) { return `<div class="empty"><div>${escape(msg)}</div></div>`; }
function toDtLocal(d) {
  if (!d) return '';
  const dt = d instanceof Date ? d : new Date(d);
  if (isNaN(dt)) return '';
  const tz = dt.getTimezoneOffset();
  return new Date(dt.getTime() - tz * 60000).toISOString().slice(0, 16);
}
