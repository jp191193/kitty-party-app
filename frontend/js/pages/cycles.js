// Kitty cycles page: schedule a multi-month cycle, list, start, cancel.
import { api } from '../api.js';
import {
  escape, openModal, toast, confirm, readForm,
  loadingHtml, skeletonTableHtml, errorHtml, fmtDate, fmtMoney, statusBadge, setButtonLoading,
} from '../ui.js';
import { openDiceRoller } from '../dice.js';
import { sendInvitationsForSchedule } from './invitations.js';

export async function renderCycles({ root, query }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Kitty cycles</h1>
        <div class="subtitle">Plan a full cycle with one host per month — auto-generates the schedule.</div>
      </div>
      <div class="page-actions" id="topActions"></div>
    </div>
    <div class="toolbar">
      <div class="field" style="flex:1;min-width:240px">
        <label>Group</label>
        <select class="select" id="groupPicker"><option value="">Loading groups…</option></select>
      </div>
    </div>
    <div id="cyclesContainer">${emptyHint('Pick a group to view its kitty cycles.')}</div>
  `;

  const picker = root.querySelector('#groupPicker');
  const container = root.querySelector('#cyclesContainer');
  const topActions = root.querySelector('#topActions');

  let groups = [];
  try { groups = (await api.groups.list()) || []; }
  catch (e) { picker.innerHTML = `<option>Failed: ${escape(e.message)}</option>`; return; }

  picker.innerHTML = `<option value="">— select a group —</option>` +
    groups.map((g) => `<option value="${escape(g.id)}">${escape(g.name)}</option>`).join('');

  async function loadFor(groupId) {
    if (!groupId) { container.innerHTML = emptyHint('Pick a group to view its cycles.'); topActions.innerHTML = ''; return; }
    const g = groups.find((x) => x.id === groupId);
    topActions.innerHTML = `<button class="btn" id="newCycleBtn">+ Schedule cycle</button>`;
    document.getElementById('newCycleBtn').onclick = () => openScheduleCycle(g, () => loadFor(groupId));

    container.innerHTML = `<div style="display:flex;flex-direction:column;gap:14px">
      ${[1,2,3].map(() => `
        <div class="card">
          <div class="card-header" style="gap:16px">
            <div style="flex:1">
              <div class="skeleton sk-line" style="width:55%;margin-bottom:8px"></div>
              <div class="skeleton sk-sm" style="width:80%"></div>
            </div>
            <div class="skeleton sk-line" style="width:120px;height:30px;border-radius:8px"></div>
          </div>
        </div>`).join('')}
    </div>`;
    try {
      const cycles = (await api.cycles.listByGroup(groupId)) || [];
      paintCycles(container, cycles, groupId, () => loadFor(groupId));
    } catch (e) { container.innerHTML = errorHtml(e.message); }
  }

  picker.addEventListener('change', () => loadFor(picker.value));
  if (query?.group) { picker.value = query.group; loadFor(query.group); }
}

function paintCycles(container, cycles, groupId, reload) {
  if (!cycles.length) {
    container.innerHTML = `<div class="empty">
      <div class="empty-title">No cycles scheduled</div>
      <div>Click "Schedule cycle" to plan a new round.</div>
    </div>`;
    return;
  }

  container.innerHTML = cycles.map((c) => `
    <div class="card" style="margin-bottom:14px">
      <div class="card-header">
        <div>
          <h3 style="margin:0">${escape(c.name || `Cycle starting ${fmtDate(c.start_date)}`)} ${statusBadge(c.status)}</h3>
          <div class="subtitle text-muted">
            ${escape(c.total_cycles)} months · ${fmtMoney(c.monthly_amount)}/mo · pool ${fmtMoney(c.pool_amount)}
            · ${fmtDate(c.start_date)} → ${fmtDate(c.end_date)}
          </div>
        </div>
        <div class="flex">
          ${c.status === 'PENDING' ? `<button class="btn btn-sm" data-start="${escape(c.id)}">Start</button>` : ''}
          ${c.status !== 'CANCELLED' && c.status !== 'COMPLETED' ? `<button class="btn btn-sm btn-danger" data-cancel="${escape(c.id)}">Cancel</button>` : ''}
          <button class="btn btn-sm btn-secondary" data-view="${escape(c.id)}">View schedule</button>
        </div>
      </div>
      ${c.notes ? `<div class="card-body text-muted">${escape(c.notes)}</div>` : ''}
    </div>
  `).join('');

  container.querySelectorAll('[data-start]').forEach((b) =>
    b.onclick = () => openStartCycle(b.dataset.start, groupId, reload)
  );
  container.querySelectorAll('[data-cancel]').forEach((b) => b.onclick = async () => {
    if (!await confirm({ title: 'Cancel cycle?', message: 'This will cancel the entire kitty cycle.', confirmText: 'Cancel cycle', danger: true })) return;
    try { await api.cycles.cancel(b.dataset.cancel, 'Cancelled from UI'); toast('Cycle cancelled', { type: 'info' }); reload(); }
    catch (e) { toast(e.message, { type: 'error', title: 'Cancel failed' }); }
  });
  container.querySelectorAll('[data-view]').forEach((b) => b.onclick = () => openCycleSchedule(b.dataset.view));
}

// ── Start cycle — confirmation modal ─────────────────────────────────────────

async function openStartCycle(cycleId, groupId, reload) {
  const wrap = document.createElement('div');
  wrap.innerHTML = loadingHtml('Loading cycle details…');
  const m = openModal({ title: 'Start kitty cycle', body: wrap, size: 'lg' });

  let cycle, members;
  try {
    [cycle, members] = await Promise.all([
      api.cycles.get(cycleId),
      api.groups.members.list(groupId),
    ]);
  } catch (e) {
    wrap.innerHTML = errorHtml(e.message);
    return;
  }

  const memberMap = Object.fromEntries((members || []).map((mem) => [mem.member_id, mem.member_name || mem.member_id]));
  const sorted = [...(cycle.schedule || [])].sort((a, b) => a.cycle_number - b.cycle_number);
  const first = sorted[0];
  const firstHostId = first?.host_member_id;
  const hostName = firstHostId
    ? (memberMap[firstHostId] || firstHostId.slice(0, 8))
    : 'TBD — roll the dice';
  const cycleName = cycle.name || `Cycle starting ${fmtDate(cycle.start_date)}`;

  wrap.innerHTML = `
    <p class="text-muted" style="margin-top:0">
      Review the cycle details below, then click <strong>Start cycle</strong> to activate.
    </p>

    <div class="card" style="padding:14px 16px;margin-bottom:14px">
      <div style="font-weight:600;margin-bottom:10px">${escape(cycleName)}</div>
      <div style="display:grid;grid-template-columns:repeat(2,1fr);gap:8px 24px">
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Duration</div>
          <div>${escape(cycle.total_cycles)} months</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Status</div>
          <div>${statusBadge(cycle.status)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Monthly amount</div>
          <div>${fmtMoney(cycle.monthly_amount)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Pool amount</div>
          <div style="font-weight:600">${fmtMoney(cycle.pool_amount)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Start</div>
          <div>${fmtDate(cycle.start_date)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">End</div>
          <div>${fmtDate(cycle.end_date)}</div>
        </div>
      </div>
    </div>

    ${first ? `
    <h4 style="margin:0 0 8px">First month — Cycle #${escape(first.cycle_number)}</h4>
    <div class="card" style="padding:14px 16px;margin-bottom:14px">
      <div style="display:grid;grid-template-columns:repeat(2,1fr);gap:8px 24px">
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Host / Recipient</div>
          <div style="font-weight:600">${escape(hostName)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Month</div>
          <div>${escape(first.cycle_month)}/${escape(first.cycle_year)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Kitty date</div>
          <div>${fmtDate(first.scheduled_date)}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Due date</div>
          <div>${fmtDate(first.due_date)}</div>
        </div>
      </div>
    </div>` : ''}

    <div style="background:var(--info-bg);border:1px solid var(--info);border-radius:var(--radius-sm);padding:12px 16px;margin-bottom:18px">
      <div style="font-weight:600;margin-bottom:6px">Starting this cycle will:</div>
      <ul style="margin:0;padding-left:18px;color:var(--text-muted);line-height:1.9">
        <li>Activate the cycle (PENDING → ACTIVE)</li>
        <li>Mark month #1 schedule as IN_PROGRESS</li>
        ${firstHostId
          ? `<li>Generate contribution dues for all group members</li>
             <li>Schedule a payout of <strong>${fmtMoney(cycle.pool_amount)}</strong> for <strong>${escape(hostName)}</strong></li>`
          : `<li class="text-muted"><strong>Heads up:</strong> month #1 has no host yet — roll the dice first to also auto-generate dues and the payout.</li>`}
      </ul>
    </div>

    <div class="form-actions">
      <button class="btn btn-secondary" id="cancelStartBtn">Cancel</button>
      <button class="btn" id="confirmStartBtn">Start cycle</button>
    </div>
  `;

  wrap.querySelector('#cancelStartBtn').onclick = () => m.close();

  const confirmBtn = wrap.querySelector('#confirmStartBtn');
  confirmBtn.onclick = async () => {
    setButtonLoading(confirmBtn, true);
    try {
      const result = await api.cycles.start(cycleId);
      m.close();
      showActivationResult(result, hostName);
      reload();
    } catch (e) {
      toast(e.message, { type: 'error', title: 'Start failed' });
      setButtonLoading(confirmBtn, false);
    }
  };
}

// ── Activation result modal ───────────────────────────────────────────────────

function showActivationResult(result, hostName) {
  const wrap = document.createElement('div');
  const duesCount = result.dues_generated || 0;
  const payoutScheduled = !!result.payout_id;

  wrap.innerHTML = `
    <div style="display:flex;align-items:center;gap:14px;padding:4px 0 18px">
      <div style="width:44px;height:44px;border-radius:50%;background:var(--success-bg);display:flex;align-items:center;justify-content:center;flex-shrink:0">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="var(--success)" stroke-width="2.5">
          <polyline points="20 6 9 17 4 12"/>
        </svg>
      </div>
      <div>
        <div style="font-weight:600;font-size:1.05em">Cycle started successfully</div>
        <div style="color:var(--text-muted)">The kitty cycle is now active.</div>
      </div>
    </div>

    <div class="card" style="padding:14px 16px;margin-bottom:16px">
      <div style="display:grid;grid-template-columns:repeat(2,1fr);gap:10px 24px">
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Cycle status</div>
          <div>${statusBadge('ACTIVE')}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Month #1 status</div>
          <div>${statusBadge('IN_PROGRESS')}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Dues generated</div>
          <div>${escape(duesCount)} member${duesCount !== 1 ? 's' : ''}</div>
        </div>
        <div>
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Pool amount</div>
          <div style="font-weight:600">${fmtMoney(result.pool_amount)}</div>
        </div>
        <div style="grid-column:span 2">
          <div style="font-size:0.78em;color:var(--text-muted);margin-bottom:2px">Payout scheduled for</div>
          <div>${payoutScheduled ? escape(hostName) : '—'}</div>
        </div>
      </div>
    </div>

    <p style="color:var(--text-muted);margin:0 0 18px;font-size:0.9em">
      Members can now record their contributions against the generated dues.
      The host will receive the pool once all dues are collected.
    </p>

    <div class="form-actions">
      <button class="btn" id="closeResultBtn">Done</button>
    </div>
  `;

  const m = openModal({ title: 'Cycle activated', body: wrap });
  wrap.querySelector('#closeResultBtn').onclick = () => m.close();
}

// ── View schedule modal ───────────────────────────────────────────────────────

async function openCycleSchedule(cycleId) {
  const wrap = document.createElement('div');
  wrap.innerHTML = skeletonTableHtml(5, ['40px', '12%', '22%', '14%', '14%', '12%', '10%', '120px']);
  const m = openModal({ title: 'Cycle schedule', body: wrap, size: 'lg' });

  async function paint() {
    let c, members;
    try {
      c = await api.cycles.get(cycleId);
      members = await api.groups.members.list(c.group_id);
    } catch (e) {
      wrap.innerHTML = errorHtml(e.message);
      return;
    }
    const sched = (c.schedule || []).slice().sort((a, b) => a.cycle_number - b.cycle_number);
    const memberMap = Object.fromEntries((members || []).map((mem) => [mem.member_id, mem.member_name || mem.member_id]));
    const takenIds = new Set(sched.filter((s) => s.host_member_id).map((s) => s.host_member_id));

    if (!sched.length) {
      wrap.innerHTML = `<div class="empty"><div>No schedule entries.</div></div>`;
      return;
    }

    wrap.innerHTML = `
      <div class="text-muted" style="margin-bottom:10px;font-size:13px">
        Roll the dice for each upcoming month to pick its host.
      </div>
      <div class="table-wrap"><table class="data-table">
        <thead><tr><th>#</th><th>Month</th><th>Host</th><th>Scheduled</th><th>Due</th><th>Pool</th><th>Status</th><th></th></tr></thead>
        <tbody>${sched.map((s) => {
          const hostCell = s.host_member_id
            ? `<a href="#/members/${escape(s.host_member_id)}">${escape(memberMap[s.host_member_id] || s.host_member_id.slice(0, 8))}</a>`
            : `<span class="dice-host-pill unrolled"><span class="dot"></span>To be rolled</span>`;

          const canRoll = !s.host_member_id && s.status !== 'CANCELLED' && s.status !== 'COMPLETED';
          const rollBtn = canRoll
            ? `<button class="btn btn-sm" data-roll="${escape(s.id)}" data-cycle-num="${escape(s.cycle_number)}" data-cycle-month="${escape(s.cycle_month)}" data-cycle-year="${escape(s.cycle_year)}">🎲 Roll dice</button>`
            : '';
          const canInvite = s.status !== 'CANCELLED' && s.status !== 'COMPLETED';
          const inviteBtn = canInvite
            ? `<button class="btn btn-sm btn-secondary" data-invite="${escape(s.id)}">✉ Invite</button>`
            : '';
          return `
            <tr>
              <td>${escape(s.cycle_number)}</td>
              <td>${escape(s.cycle_month)}/${escape(s.cycle_year)}</td>
              <td>${hostCell}</td>
              <td class="text-muted">${fmtDate(s.scheduled_date)}</td>
              <td class="text-muted">${fmtDate(s.due_date)}</td>
              <td>${fmtMoney(s.pool_amount)}</td>
              <td>${statusBadge(s.status)}</td>
              <td class="text-right" style="white-space:nowrap">${rollBtn} ${inviteBtn}</td>
            </tr>`;
        }).join('')}</tbody></table></div>`;
    wrap.querySelectorAll('[data-roll]').forEach((btn) => {
      btn.onclick = () => {
        const candidates = (members || []).filter((mem) => !takenIds.has(mem.member_id));
        if (!candidates.length) {
          toast('Every member has already hosted in this cycle.', { type: 'warning' });
          return;
        }
        openDiceRoller({
          scheduleId: btn.dataset.roll,
          cycleLabel: `Cycle #${btn.dataset.cycleNum} · ${btn.dataset.cycleMonth}/${btn.dataset.cycleYear}`,
          candidates,
          onAssigned: () => paint(),
        });
      };
    });

    wrap.querySelectorAll('[data-invite]').forEach((btn) => {
      btn.onclick = async () => {
        setButtonLoading(btn, true);
        try {
          await sendInvitationsForSchedule(btn.dataset.invite);
        } finally {
          setButtonLoading(btn, false);
        }
      };
    });
  }
  paint();
}

// ── Schedule new cycle modal ──────────────────────────────────────────────────

function openScheduleCycle(group, onSaved) {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <p class="text-muted" style="margin-top:0">
      Pick the start month, set the monthly amount, and adjust each month's
      scheduled & due dates. Hosts are picked later by rolling the dice — one
      month at a time, just before that cycle.
    </p>
    <form id="scheduleForm">
      <div class="form-grid">
        <div class="field full"><label>Cycle name (optional)</label>
          <input class="input" name="name" placeholder="e.g. ${escape(group.name)} 2026" /></div>
        <div class="field"><label>Monthly amount (₹)</label>
          <input class="input" type="number" name="monthly_amount" min="1" required value="${escape(group.monthly_amount || '')}" /></div>
        <div class="field"><label>Number of months</label>
          <input class="input" type="number" id="durationInput" min="1" max="60" required value="${escape(group.duration || 6)}" /></div>
        <div class="field"><label>Start date</label>
          <input class="input" type="datetime-local" name="start_date" id="startDateInput" required value="${toDtLocal(new Date())}" /></div>
        <div class="field"><label>End date</label>
          <input class="input" type="datetime-local" name="end_date" id="endDateInput" required /></div>
        <div class="field full"><label>Notes</label>
          <textarea class="textarea" name="notes" placeholder="Optional"></textarea></div>
      </div>

      <h4 style="margin:18px 0 8px">Monthly schedule</h4>
      <div class="text-muted" style="font-size:12.5px;margin-bottom:8px">
        🎲 Hosts will be assigned one-by-one via the dice roller before each cycle.
      </div>
      <div id="scheduleRows" class="grid" style="gap:8px"></div>

      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Schedule cycle</button>
      </div>
    </form>`;
  const m = openModal({ title: 'Schedule a kitty cycle', body: wrap, size: 'lg' });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();

  const rowsEl = wrap.querySelector('#scheduleRows');
  const durationEl = wrap.querySelector('#durationInput');
  const startEl = wrap.querySelector('#startDateInput');
  const endEl = wrap.querySelector('#endDateInput');

  function renderRows() {
    const n = Math.min(1, +durationEl.value || 1);
    const start = new Date(startEl.value || Date.now());
    if (!endEl.value) {
      const end = new Date(start.getFullYear(), start.getMonth() + n, start.getDate());
      endEl.value = toDtLocal(end);
    }
    rowsEl.innerHTML = Array.from({ length: n }, (_, i) => {
      const sched = new Date(start.getFullYear(), start.getMonth() + i, start.getDate());
      const due = new Date(sched.getFullYear(), sched.getMonth(), sched.getDate() + 14);
      return `
        <div class="card" style="padding:10px 12px;display:grid;grid-template-columns:60px 1fr 1.2fr 1.2fr;gap:10px;align-items:end">
          <div class="field"><label>#${i + 1}</label><div class="text-muted mono">${sched.getMonth() + 1}/${sched.getFullYear()}</div></div>
          <div class="field"><label>Host</label>
            <span class="dice-host-pill unrolled"><span class="dot"></span>To be rolled</span></div>
          <div class="field"><label>Scheduled</label>
            <input class="input" type="datetime-local" data-sched="${i}" value="${toDtLocal(sched)}" /></div>
          <div class="field"><label>Due</label>
            <input class="input" type="datetime-local" data-due="${i}" value="${toDtLocal(due)}" /></div>
        </div>`;
    }).join('');
  }
  durationEl.addEventListener('input', renderRows);
  startEl.addEventListener('change', () => { endEl.value = ''; renderRows(); });
  renderRows();

  wrap.querySelector('#scheduleForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    delete data['']; // safety

    const n = Math.max(1, +durationEl.value || 1);
    const schedule = [];
    let invalid = false;
    for (let i = 0; i < n; i++) {
      const sched = wrap.querySelector(`[data-sched="${i}"]`).value;
      const due   = wrap.querySelector(`[data-due="${i}"]`).value;
      if (!sched || !due) { invalid = true; break; }
      const sd = new Date(sched);
      schedule.push({
        cycle_number: i + 1,
        cycle_month: sd.getMonth() + 1,
        cycle_year: sd.getFullYear(),
        scheduled_date: sd.toISOString(),
        due_date: new Date(due).toISOString(),
      });
    }
    if (invalid) { toast('Set a scheduled and due date for every month.', { type: 'warning' }); return; }

    const payload = {
      group_id: group.id,
      name: data.name,
      monthly_amount: data.monthly_amount,
      start_date: data.start_date,
      end_date: data.end_date,
      notes: data.notes,
      schedule,
    };
    const btn = e.target.querySelector('button[type=submit]');
    setButtonLoading(btn, true);
    try {
      await api.cycles.schedule(payload);
      toast('Cycle scheduled — roll the dice when you\'re ready to pick each host.', { type: 'success' });
      m.close(); onSaved?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Schedule failed' });
    } finally { setButtonLoading(btn, false); }
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
