// Minimal hash router. Routes are registered as { path: '/x', render(ctx) }.
// Path params with `:id` work; rest goes into ctx.params.

const routes = [];
let notFoundHandler = null;
let beforeRender = null;

export function registerRoute(path, render) { routes.push({ path, render }); }
export function setNotFound(fn) { notFoundHandler = fn; }
export function onBeforeRender(fn) { beforeRender = fn; }

export function navigate(to) {
  if (!to.startsWith('#')) to = '#' + to;
  if (location.hash === to) renderCurrent();
  else location.hash = to;
}

export function currentPath() {
  const h = location.hash || '#/';
  return h.startsWith('#') ? h.slice(1) : h;
}

function match(path, pattern) {
  const ps = pattern.split('/').filter(Boolean);
  const xs = path.split('?')[0].split('/').filter(Boolean);
  if (ps.length !== xs.length) return null;
  const params = {};
  for (let i = 0; i < ps.length; i++) {
    if (ps[i].startsWith(':')) params[ps[i].slice(1)] = decodeURIComponent(xs[i]);
    else if (ps[i] !== xs[i]) return null;
  }
  return params;
}

function parseQuery(path) {
  const q = path.split('?')[1];
  if (!q) return {};
  return Object.fromEntries(new URLSearchParams(q));
}

export async function renderCurrent() {
  const path = currentPath();
  const main = document.getElementById('appMain');
  if (!main) return;

  for (const r of routes) {
    const params = match(path, r.path);
    if (params) {
      const query = parseQuery(path);
      if (beforeRender && beforeRender({ path, params, query }) === false) return;
      main.classList.remove('page-enter');
      // Force reflow so animation re-runs
      void main.offsetWidth;
      main.classList.add('page-enter');
      try {
        await r.render({ root: main, params, query });
      } catch (err) {
        console.error(err);
        main.innerHTML = `<div class="loader"><p style="color:var(--danger)">Page failed: ${err.message}</p></div>`;
      }
      return;
    }
  }
  if (notFoundHandler) notFoundHandler({ root: main, path });
  else main.innerHTML = `<div class="loader"><p>404 — ${path}</p></div>`;
}

export function startRouter() {
  window.addEventListener('hashchange', renderCurrent);
  if (!location.hash) location.hash = '#/';
  renderCurrent();
}
