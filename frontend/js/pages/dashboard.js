// Dashboard: an at-a-glance summary of members, groups, recent activity.
import { api } from '../api.js';
import { escape, fmtMoney, loadingHtml, errorHtml, statusBadge, fmtDate } from '../ui.js';

export async function renderDashboard({ root }) {
  root.innerHTML = `
    <div class="page-header">
      <div>
        <h1>Welcome back</h1>
        <div class="subtitle">A quick overview of your kitty parties.</div>
      </div>
      <div class="page-actions">
        <a class="btn btn-secondary" href="#/members">Manage members</a>
        <a class="btn" href="#/groups">View groups</a>
      </div>
    </div>
    <div id="statsGrid">${loadingHtml('Loading stats…')}</div>
    <div class="grid grid-cols-2" style="margin-top:20px">
      <div class="card">
        <div class="card-header"><h3>Recent groups</h3><a class="btn btn-ghost btn-sm" href="#/groups">View all →</a></div>
        <div id="recentGroups">${loadingHtml()}</div>
      </div>
      <div class="card">
        <div class="card-header"><h3>Recent members</h3><a class="btn btn-ghost btn-sm" href="#/members">View all →</a></div>
        <div id="recentMembers">${loadingHtml()}</div>
      </div>
    </div>
  `;

  const [membersRes, groupsRes, countsRes] = await Promise.allSettled([
    api.members.list(),
    api.groups.list(),
    api.groups.memberCounts(),
  ]);

  const members = membersRes.status === 'fulfilled' ? (membersRes.value || []) : [];
  const groups  = groupsRes.status  === 'fulfilled' ? (groupsRes.value  || []) : [];
  const counts  = countsRes.status  === 'fulfilled' ? (countsRes.value  || []) : [];

  if (membersRes.status === 'rejected' && groupsRes.status === 'rejected') {
    root.querySelector('#statsGrid').innerHTML = errorHtml(membersRes.reason?.message || 'Backend not reachable');
    return;
  }

  const totalPool = groups.reduce((s, g) => s + (Number(g.monthly_amount) || 0) * (Number(g.duration) || 0), 0);
  const totalMembers = counts.reduce((s, g) => s + (g.count || 0), 0);

  root.querySelector('#statsGrid').innerHTML = `
    <div class="grid grid-cols-4">
      ${statCard('Members', members.length, 'Registered members')}
      ${statCard('Groups', groups.length, 'Active kitty groups')}
      ${statCard('Memberships', totalMembers, 'Total enrolments across groups')}
      ${statCard('Lifetime pool', fmtMoney(totalPool), 'Sum of monthly × duration')}
    </div>
  `;

  // Recent groups (newest first)
  const sortedGroups = [...groups].sort((a, b) => new Date(b.created_at) - new Date(a.created_at)).slice(0, 5);
  root.querySelector('#recentGroups').innerHTML = sortedGroups.length
    ? `<div class="table-wrap"><table class="data-table">
        <thead><tr><th>Name</th><th>Monthly</th><th>Duration</th><th>Created</th></tr></thead>
        <tbody>${sortedGroups.map((g) => `
          <tr>
            <td><a href="#/groups/${escape(g.id)}">${escape(g.name)}</a></td>
            <td>${fmtMoney(g.monthly_amount)}</td>
            <td>${escape(g.duration)} mo</td>
            <td class="text-muted">${fmtDate(g.created_at)}</td>
          </tr>`).join('')}</tbody></table></div>`
    : emptyMini('No groups yet', 'Create your first group from the Groups page.');

  // Recent members
  const sortedMembers = [...members].sort((a, b) => new Date(b.created_at) - new Date(a.created_at)).slice(0, 5);
  root.querySelector('#recentMembers').innerHTML = sortedMembers.length
    ? `<div class="table-wrap"><table class="data-table">
        <thead><tr><th>Name</th><th>Email</th><th>Joined</th></tr></thead>
        <tbody>${sortedMembers.map((m) => `
          <tr>
            <td><a href="#/members/${escape(m.id)}">${escape(m.name)}</a></td>
            <td class="text-muted">${escape(m.email)}</td>
            <td class="text-muted">${fmtDate(m.created_at)}</td>
          </tr>`).join('')}</tbody></table></div>`
    : emptyMini('No members yet', 'Add your first member from the Members page.');
}

function statCard(label, value, sub) {
  return `
    <div class="stat">
      <div class="stat-label">${escape(label)}</div>
      <div class="stat-value">${typeof value === 'string' ? escape(value) : value}</div>
      <div class="stat-sub">${escape(sub)}</div>
    </div>`;
}

function emptyMini(t, s) {
  return `<div class="empty"><div class="empty-title">${escape(t)}</div><div>${escape(s)}</div></div>`;
}
