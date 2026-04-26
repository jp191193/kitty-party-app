// Invitations page: view pending invites, accept dates, propose new dates, vote on proposals.
import { api } from '../api.js';
import { store } from '../store.js';
import {
  escape, openModal, toast, confirm, readForm,
  loadingHtml, skeletonTableHtml, errorHtml, fmtDate, statusBadge, setButtonLoading,
} from '../ui.js';

export async function renderInvitations({ root }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>My invitations</h1>
        <div class="subtitle">Respond to scheduled kitty party events — accept the date or propose a new one.</div>
      </div>
    </div>
    <div id="invitationsContainer">${loadingHtml('Loading invitations…')}</div>
  `;

  const container = root.querySelector('#invitationsContainer');
  await loadInvitations(container);
}

async function loadInvitations(container) {
  container.innerHTML = `<div style="display:flex;flex-direction:column;gap:14px">
    ${[1,2].map(() => `
      <div class="card">
        <div class="card-header" style="gap:16px">
          <div style="flex:1">
            <div class="skeleton sk-line" style="width:60%;margin-bottom:8px"></div>
            <div class="skeleton sk-sm" style="width:45%"></div>
          </div>
          <div style="display:flex;gap:8px">
            <div class="skeleton sk-line" style="width:90px;height:28px;border-radius:8px"></div>
            <div class="skeleton sk-line" style="width:110px;height:28px;border-radius:8px"></div>
          </div>
        </div>
      </div>`).join('')}
  </div>`;
  try {
    const invitations = (await api.invitations.my()) || [];
    paintInvitations(container, invitations);
  } catch (e) {
    container.innerHTML = errorHtml(e.message);
  }
}

function paintInvitations(container, invitations) {
  if (!invitations.length) {
    container.innerHTML = `
      <div class="empty">
        <div class="empty-title">No invitations</div>
        <div>You have no pending invitations. When a host sends one, it will appear here.</div>
      </div>`;
    return;
  }

  container.innerHTML = invitations.map((inv) => {
    const statusColor = {
      PENDING: 'warning',
      ACCEPTED: 'success',
      DATE_PROPOSED: 'info',
    }[inv.status] || 'default';

    return `
      <div class="card" style="margin-bottom:14px" data-inv-id="${escape(inv.id)}">
        <div class="card-header">
          <div>
            <h3 style="margin:0">
              ${escape(inv.group_name || 'Kitty Group')} — Cycle #${escape(inv.cycle_number)}
              ${statusBadge(inv.status)}
            </h3>
            <div class="subtitle text-muted">
              ${inv.cycle_month}/${inv.cycle_year} · Scheduled: <strong>${fmtDate(inv.scheduled_date)}</strong>
            </div>
          </div>
          <div class="flex" style="gap:8px">
            ${inv.status === 'PENDING' ? `
              <button class="btn btn-sm" data-act="accept" data-id="${escape(inv.id)}">Accept date</button>
              <button class="btn btn-sm btn-secondary" data-act="propose" data-id="${escape(inv.id)}" data-schedule="${escape(inv.schedule_id)}">Propose new date</button>
            ` : ''}
            <button class="btn btn-sm btn-ghost" data-act="view-proposals" data-schedule="${escape(inv.schedule_id)}">View proposals</button>
          </div>
        </div>
      </div>`;
  }).join('');

  container.querySelectorAll('[data-act="accept"]').forEach((btn) => {
    btn.onclick = async () => {
      setButtonLoading(btn, true);
      try {
        await api.invitations.accept(btn.dataset.id);
        toast('Date accepted!', { type: 'success' });
        await loadInvitations(container);
      } catch (e) {
        toast(e.message, { type: 'error', title: 'Accept failed' });
        setButtonLoading(btn, false);
      }
    };
  });

  container.querySelectorAll('[data-act="propose"]').forEach((btn) => {
    btn.onclick = () => openProposeDateModal(btn.dataset.id, btn.dataset.schedule, () => loadInvitations(container));
  });

  container.querySelectorAll('[data-act="view-proposals"]').forEach((btn) => {
    btn.onclick = () => openProposalsModal(btn.dataset.schedule);
  });
}

// ── Propose date modal ────────────────────────────────────────────────────────

function openProposeDateModal(invitationId, scheduleId, onDone) {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <p class="text-muted" style="margin-top:0">
      Pick an alternative date. All group members will be notified and can vote on it.
    </p>
    <form id="proposeDateForm">
      <div class="field">
        <label>Proposed date</label>
        <input class="input" type="date" name="proposed_date" required />
      </div>
      <div class="form-actions">
        <button type="button" class="btn btn-secondary" data-act="cancel">Cancel</button>
        <button type="submit" class="btn">Submit proposal</button>
      </div>
    </form>
  `;

  const m = openModal({ title: 'Propose a new date', body: wrap });
  wrap.querySelector('[data-act="cancel"]').onclick = () => m.close();

  wrap.querySelector('#proposeDateForm').onsubmit = async (e) => {
    e.preventDefault();
    const data = readForm(e.target);
    if (!data.proposed_date) {
      toast('Please pick a date.', { type: 'warning' });
      return;
    }

    const btn = e.target.querySelector('button[type=submit]');
    setButtonLoading(btn, true);
    try {
      // Mark invitation as DATE_PROPOSED and create the proposal in parallel.
      await Promise.all([
        api.invitations.proposeDate(invitationId),
        api.proposals.create({
          schedule_id: scheduleId,
          proposed_date: new Date(data.proposed_date).toISOString(),
        }),
      ]);
      toast('Proposal submitted — group members can now vote.', { type: 'success' });
      m.close();
      onDone?.();
    } catch (err) {
      toast(err.message, { type: 'error', title: 'Proposal failed' });
      setButtonLoading(btn, false);
    }
  };
}

// ── View proposals modal ──────────────────────────────────────────────────────

async function openProposalsModal(scheduleId) {
  const wrap = document.createElement('div');
  wrap.innerHTML = `<div style="display:flex;flex-direction:column;gap:14px">
    ${[1,2].map(() => `
      <div class="card">
        <div class="card-header">
          <div style="flex:1">
            <div class="skeleton sk-line" style="width:45%;margin-bottom:8px"></div>
            <div class="skeleton sk-sm" style="width:30%"></div>
          </div>
          <div style="display:flex;gap:6px">
            <div class="skeleton" style="width:88px;height:28px;border-radius:8px"></div>
            <div class="skeleton" style="width:88px;height:28px;border-radius:8px"></div>
          </div>
        </div>
      </div>`).join('')}
  </div>`;
  const m = openModal({ title: 'Date proposals', body: wrap, size: 'lg' });

  async function paint() {
    wrap.innerHTML = `<div style="display:flex;flex-direction:column;gap:14px">
      ${[1].map(() => `<div class="card"><div class="card-header">
        <div class="skeleton sk-line" style="width:40%;flex:1"></div>
      </div></div>`).join('')}
    </div>`;
    try {
      const proposals = (await api.proposals.listBySchedule(scheduleId)) || [];
      paintProposals(wrap, proposals, scheduleId, paint);
    } catch (e) {
      wrap.innerHTML = errorHtml(e.message);
    }
  }
  paint();
}

function paintProposals(wrap, proposals, scheduleId, reload) {
  if (!proposals.length) {
    wrap.innerHTML = `
      <div class="empty">
        <div>No proposals yet.</div>
        <div class="text-muted" style="margin-top:6px">Members can propose an alternative date from their invitations page.</div>
      </div>`;
    return;
  }

  wrap.innerHTML = proposals.map((p) => {
    const isOpen = p.status === 'OPEN';
    return `
      <div class="card" style="margin-bottom:14px">
        <div class="card-header">
          <div>
            <div style="font-weight:600">
              ${fmtDate(p.proposed_date)} ${statusBadge(p.status)}
            </div>
            <div class="text-muted" style="font-size:13px">
              Proposed by ${escape(p.proposer_name || p.proposed_by.slice(0, 8))}
            </div>
          </div>
          <div class="flex" style="gap:8px">
            ${isOpen ? `
              <button class="btn btn-sm" data-vote="ACCEPT" data-id="${escape(p.id)}">Vote Accept</button>
              <button class="btn btn-sm btn-secondary" data-vote="REJECT" data-id="${escape(p.id)}">Vote Reject</button>
              <button class="btn btn-sm btn-ghost" data-host-accept data-id="${escape(p.id)}">Host accept</button>
            ` : ''}
            <button class="btn btn-sm btn-ghost" data-view-votes data-id="${escape(p.id)}">View votes</button>
          </div>
        </div>
      </div>`;
  }).join('');

  wrap.querySelectorAll('[data-vote]').forEach((btn) => {
    btn.onclick = async () => {
      setButtonLoading(btn, true);
      try {
        const result = await api.proposals.vote(btn.dataset.id, btn.dataset.vote);
        const autoAccepted = result.status === 'ACCEPTED';
        toast(
          autoAccepted
            ? `All members accepted — schedule date updated to ${fmtDate(result.proposed_date)}!`
            : `Vote recorded (${result.accept_count} accept / ${result.reject_count} reject)`,
          { type: autoAccepted ? 'success' : 'info' }
        );
        reload();
      } catch (e) {
        toast(e.message, { type: 'error', title: 'Vote failed' });
        setButtonLoading(btn, false);
      }
    };
  });

  wrap.querySelectorAll('[data-host-accept]').forEach((btn) => {
    btn.onclick = async () => {
      if (!await confirm({
        title: 'Accept proposal?',
        message: 'This will update the schedule entry to the proposed date.',
        confirmText: 'Accept',
      })) return;
      setButtonLoading(btn, true);
      try {
        const result = await api.proposals.accept(btn.dataset.id);
        toast(`Schedule date updated to ${fmtDate(result.proposed_date)}`, { type: 'success' });
        reload();
      } catch (e) {
        toast(e.message, { type: 'error', title: 'Accept failed' });
        setButtonLoading(btn, false);
      }
    };
  });

  wrap.querySelectorAll('[data-view-votes]').forEach((btn) => {
    btn.onclick = () => openVotesModal(btn.dataset.id);
  });
}

// ── View votes breakdown modal ────────────────────────────────────────────────

async function openVotesModal(proposalId) {
  const wrap = document.createElement('div');
  wrap.innerHTML = skeletonTableHtml(3, ['35%', '25%', '30%']);
  openModal({ title: 'Votes breakdown', body: wrap });

  try {
    const result = await api.proposals.get(proposalId);
    const votes = result.votes || [];

    if (!votes.length) {
      wrap.innerHTML = `<div class="empty"><div>No votes cast yet.</div></div>`;
      return;
    }

    wrap.innerHTML = `
      <div style="display:flex;gap:24px;margin-bottom:16px">
        <div style="color:var(--success);font-weight:600">✓ Accept: ${escape(result.accept_count)}</div>
        <div style="color:var(--danger);font-weight:600">✗ Reject: ${escape(result.reject_count)}</div>
      </div>
      <div class="table-wrap">
        <table class="data-table">
          <thead><tr><th>Member</th><th>Vote</th><th>Time</th></tr></thead>
          <tbody>
            ${votes.map((v) => `
              <tr>
                <td>${escape(v.member_name || v.member_id.slice(0, 8))}</td>
                <td>${v.vote === 'ACCEPT'
                  ? `<span style="color:var(--success);font-weight:600">✓ Accept</span>`
                  : `<span style="color:var(--danger);font-weight:600">✗ Reject</span>`}
                </td>
                <td class="text-muted">${fmtDate(v.voted_at)}</td>
              </tr>`).join('')}
          </tbody>
        </table>
      </div>`;
  } catch (e) {
    wrap.innerHTML = errorHtml(e.message);
  }
}

// ── Send invitations helper (called from cycles.js) ───────────────────────────

export async function sendInvitationsForSchedule(scheduleId) {
  debugger;
  try {
    const result = await api.invitations.send(scheduleId);
    toast(`Invitations sent to ${result.count} member${result.count !== 1 ? 's' : ''}.`, { type: 'success' });
    return result;
  } catch (e) {
    toast(e.message, { type: 'error', title: 'Send failed' });
    throw e;
  }
}
