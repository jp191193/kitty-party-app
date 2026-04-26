// Tiny app-wide store: acting user, token, and a pub/sub for change events.
import { STORAGE_KEYS } from './config.js';

const listeners = new Set();

export const store = {
  get token()      { return localStorage.getItem(STORAGE_KEYS.TOKEN)       || ''; },
  get userId()     { return localStorage.getItem(STORAGE_KEYS.USER_ID)     || ''; },
  get userName()   { return localStorage.getItem(STORAGE_KEYS.USER_NAME)   || ''; },
  get globalRole() { return localStorage.getItem(STORAGE_KEYS.GLOBAL_ROLE) || 'User'; },

  get isSuperAdmin() { return this.globalRole === 'SuperAdmin'; },

  setUser({ id, name, token, globalRole }) {
    if (id)         localStorage.setItem(STORAGE_KEYS.USER_ID,     id);
    if (name)       localStorage.setItem(STORAGE_KEYS.USER_NAME,   name);
    if (token)      localStorage.setItem(STORAGE_KEYS.TOKEN,       token);
    if (globalRole) localStorage.setItem(STORAGE_KEYS.GLOBAL_ROLE, globalRole);
    emit();
  },

  clearUser() {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
    localStorage.removeItem(STORAGE_KEYS.USER_ID);
    localStorage.removeItem(STORAGE_KEYS.USER_NAME);
    localStorage.removeItem(STORAGE_KEYS.GLOBAL_ROLE);
    emit();
  },

  subscribe(fn) { listeners.add(fn); return () => listeners.delete(fn); },
};

function emit() { listeners.forEach((fn) => { try { fn(); } catch (_) {} }); }
