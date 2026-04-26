// Animated host-picker dice roller. Opens a modal, tumbles a 3D die while
// cycling member names in the foreground, calls the backend to pick the
// real host, then settles on the result with a confetti burst.

import { api } from './api.js';
import { escape, openModal, toast } from './ui.js';

const CONFETTI_COLORS = ['#c2410c', '#e97947', '#16a34a', '#2563eb', '#d97706', '#0f766e'];

// Build the static SVG/HTML for one cube face — `pips` controls how many.
function diceFaceHtml(faceClass, pips) {
  return `<div class="dice-face ${faceClass}">${
    Array.from({ length: pips }, () => '<span class="pip"></span>').join('')
  }</div>`;
}

function diceCubeHtml() {
  return `
    <div class="dice tumbling">
      ${diceFaceHtml('f1', 1)}
      ${diceFaceHtml('f2', 2)}
      ${diceFaceHtml('f3', 3)}
      ${diceFaceHtml('f4', 4)}
      ${diceFaceHtml('f5', 5)}
      ${diceFaceHtml('f6', 6)}
    </div>`;
}

function spawnConfetti(layer, count = 36) {
  const frag = document.createDocumentFragment();
  for (let i = 0; i < count; i++) {
    const piece = document.createElement('span');
    piece.className = 'confetti';
    const angle = Math.random() * Math.PI * 2;
    const dist = 90 + Math.random() * 90;
    const dx = Math.cos(angle) * dist;
    const dy = Math.sin(angle) * dist - 30;
    const rot = (Math.random() * 720 - 360) | 0;
    piece.style.background = CONFETTI_COLORS[i % CONFETTI_COLORS.length];
    piece.style.setProperty('--confetti-end', `translate(${dx.toFixed(0)}px, ${dy.toFixed(0)}px) rotate(${rot}deg)`);
    piece.style.animationDelay = `${(Math.random() * 120) | 0}ms`;
    frag.appendChild(piece);
  }
  layer.appendChild(frag);
}

/**
 * Open the dice roller modal.
 *
 * @param {object} opts
 * @param {string} opts.scheduleId   – kitty_schedule entry to assign a host to
 * @param {string} opts.cycleLabel   – e.g. "Cycle #3 · Mar 2026"
 * @param {Array}  opts.candidates   – [{member_id, member_name}] – cycled in the ticker
 * @param {() => void=} opts.onAssigned – called after the host is successfully assigned
 */
export function openDiceRoller({ scheduleId, cycleLabel, candidates, onAssigned }) {
  const wrap = document.createElement('div');
  wrap.innerHTML = `
    <p class="text-muted" style="margin-top:0;text-align:center">
      Rolling for <strong style="color:var(--text)">${escape(cycleLabel)}</strong>.
      The dice will pick this month's host from the group.
    </p>

    <div class="dice-stage" id="diceStage">
      <span class="confetti-layer" id="confettiLayer"></span>
      ${diceCubeHtml()}
    </div>

    <div class="dice-ticker" id="diceTicker">…</div>

    <div id="diceResult" style="display:none"></div>

    <div class="form-actions" style="justify-content:center">
      <button type="button" class="btn" id="rollBtn">
        🎲 Roll the dice
      </button>
      <button type="button" class="btn btn-secondary" id="closeBtn" style="display:none">Done</button>
    </div>
  `;

  const m = openModal({ title: 'Pick this month\'s host', body: wrap });

  const stage = wrap.querySelector('#diceStage');
  const dice = stage.querySelector('.dice');
  const confettiLayer = wrap.querySelector('#confettiLayer');
  const ticker = wrap.querySelector('#diceTicker');
  const result = wrap.querySelector('#diceResult');
  const rollBtn = wrap.querySelector('#rollBtn');
  const closeBtn = wrap.querySelector('#closeBtn');

  // Idle state: gentle preview names.
  const names = (candidates || []).map((c) => c.member_name || c.member_id.slice(0, 8));
  ticker.textContent = names.length ? names[0] : 'Ready to roll';

  let tickerTimer = null;
  function startTicker() {
    if (!names.length) return;
    let i = 0;
    tickerTimer = setInterval(() => {
      i = (i + 1) % names.length;
      ticker.textContent = names[i];
    }, 110);
  }
  function stopTicker() {
    if (tickerTimer) { clearInterval(tickerTimer); tickerTimer = null; }
  }

  rollBtn.onclick = async () => {
    rollBtn.disabled = true;
    rollBtn.textContent = '🎲 Rolling…';

    // Show shake → tumble.
    dice.classList.remove('settling');
    dice.classList.add('shaking');
    setTimeout(() => {
      dice.classList.remove('shaking');
      dice.classList.add('tumbling');
    }, 350);

    startTicker();

    // Kick off the API call alongside the animation. Race the visual
    // dwell time so the dice always tumbles for at least ~1.6s.
    const minDwell = new Promise((r) => setTimeout(r, 1600));
    let rolled, apiErr;
    try { rolled = await api.cycles.rollHost(scheduleId); }
    catch (e) { apiErr = e; }
    await minDwell;

    stopTicker();
    dice.classList.remove('tumbling');

    if (apiErr) {
      ticker.textContent = '';
      result.style.display = 'block';
      result.innerHTML = `
        <div class="dice-winner" style="color:var(--danger)">
          <div class="dice-winner-label">Couldn't roll</div>
          <div style="font-size:16px;font-weight:600;color:var(--danger)">${escape(apiErr.message)}</div>
        </div>`;
      rollBtn.style.display = 'none';
      closeBtn.style.display = 'inline-flex';
      return;
    }

    // Settle and reveal.
    dice.classList.add('settling');
    ticker.textContent = '';

    const winnerName = rolled.host_member_name || rolled.host_member_id?.slice(0, 8) || 'Member';
    result.style.display = 'block';
    result.innerHTML = `
      <div class="dice-winner">
        <div class="dice-winner-label">🎉 This month's host</div>
        <div class="dice-winner-name">${escape(winnerName)}</div>
        <div class="dice-winner-sub">Locked in for ${escape(cycleLabel)}</div>
      </div>`;

    spawnConfetti(confettiLayer, 42);

    rollBtn.style.display = 'none';
    closeBtn.style.display = 'inline-flex';

    toast(`${winnerName} hosts ${cycleLabel}`, { type: 'success', title: 'Host picked' });
    onAssigned?.(rolled);
  };

  closeBtn.onclick = () => m.close();
}
