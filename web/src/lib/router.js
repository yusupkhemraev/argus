import { writable, derived } from 'svelte/store';

function getHash() {
  const hash = window.location.hash.slice(1) || '/';
  return hash.startsWith('/') ? hash : '/' + hash;
}

export const route = writable(getHash());

if (typeof window !== 'undefined') {
  window.addEventListener('hashchange', () => {
    route.set(getHash());
  });
}

export function navigate(path) {
  window.location.hash = '#' + path;
}

export const currentPage = derived(route, ($route) => {
  if ($route.startsWith('/alarms')) return 'alarms';
  if ($route.startsWith('/logs')) return 'logs';
  if ($route.startsWith('/settings')) return 'settings';
  return 'dashboard';
});
