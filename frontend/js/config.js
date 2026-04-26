// Central runtime configuration. Change API_BASE to point at a deployed backend.
// You can also override via localStorage key `kp.apiBase`.

const DEFAULT_API_BASE = 'http://localhost:8080';

export const API_BASE =
  (typeof localStorage !== 'undefined' && localStorage.getItem('kp.apiBase')) ||
  DEFAULT_API_BASE;

export const STORAGE_KEYS = {
  TOKEN: 'kp.token',
  USER_ID: 'kp.userId',
  USER_NAME: 'kp.userName',
  API_BASE: 'kp.apiBase',
};
