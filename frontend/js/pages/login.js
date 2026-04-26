import { api } from '../api.js';
import { store } from '../store.js';
import { escape, readForm, setButtonLoading } from '../ui.js';
import { navigate } from '../router.js';

export async function renderLogin({ root }) {
  root.innerHTML = `
    <div class="login-wrap">
      <div class="card login-card">
        <div class="login-brand">
          <span class="brand-logo" style="font-size:2rem;width:3rem;height:3rem;">K</span>
          <h1 class="login-title">Sign in to Kitty Party</h1>
          <p class="text-muted login-subtitle">Welcome back! Please enter your details.</p>
        </div>
        <form id="loginForm" novalidate>
          <div class="field full">
            <label for="loginEmail">Email</label>
            <input class="input" id="loginEmail" name="email" type="email" required
                   autocomplete="email" placeholder="you@example.com" />
          </div>
          <div class="field full">
            <label for="loginPassword">Password</label>
            <input class="input" id="loginPassword" name="password" type="password" required
                   autocomplete="current-password" placeholder="••••••••" />
          </div>
          <p id="loginError" class="login-error" style="display:none"></p>
          <div class="form-actions" style="margin-top:1.25rem">
            <button type="submit" class="btn" style="width:100%">Sign in</button>
          </div>
        </form>
        <p class="login-hint">Don't have an account? Create a member from the Members page after signing in.</p>
      </div>
    </div>
  `;

  const form = root.querySelector('#loginForm');
  const errorEl = root.querySelector('#loginError');

  form.onsubmit = async (e) => {
    e.preventDefault();
    errorEl.style.display = 'none';
    const data = readForm(form);
    const btn = form.querySelector('[type=submit]');
    setButtonLoading(btn, true);
    try {
      const res = await api.auth.login(data);
      store.setUser({ id: res.user.id, name: res.user.name, token: res.token, globalRole: res.user.global_role });
      navigate('/');
    } catch (err) {
      errorEl.textContent = escape(err.message || 'Invalid email or password');
      errorEl.style.display = 'block';
    } finally {
      setButtonLoading(btn, false);
    }
  };
}
