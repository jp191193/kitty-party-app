// Tiny app-wide store: acting user, token, and a pub/sub for change events.
import { STORAGE_KEYS } from './config.js';

const listeners = new Set();

export const store = {
  get token() { return localStorage.getItem(STORAGE_KEYS.TOKEN) || ''; },
  get userId() { return localStorage.getItem(STORAGE_KEYS.USER_ID) || ''; },
  get userName() { return localStorage.getItem(STORAGE_KEYS.USER_NAME) || ''; },

  setUser({ id, name, token }) {
    if (id) localStorage.setItem(STORAGE_KEYS.USER_ID, id);
    if (name) localStorage.setItem(STORAGE_KEYS.USER_NAME, name);
    if (token) localStorage.setItem(STORAGE_KEYS.TOKEN, token);
    emit();
  },

  clearUser() {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
    localStorage.removeItem(STORAGE_KEYS.USER_ID);
    localStorage.removeItem(STORAGE_KEYS.USER_NAME);
    emit();
  },

  subscribe(fn) { listeners.add(fn); return () => listeners.delete(fn); },
};

function emit() { listeners.forEach((fn) => { try { fn(); } catch (_) {} }); }
