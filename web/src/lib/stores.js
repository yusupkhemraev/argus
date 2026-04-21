import { writable, derived } from 'svelte/store';

export const metrics = writable({});
export const alarms = writable([]);
export const connected = writable(false);
export const filter = writable('');
export const severityFilter = writable('all');
export const startTime = writable(Date.now());

export const filteredAlarms = derived(
  [alarms, filter, severityFilter],
  ([$alarms, $filter, $severityFilter]) => {
    let result = $alarms;

    if ($severityFilter !== 'all') {
      result = result.filter(a => a.severity?.toLowerCase() === $severityFilter);
    }

    if ($filter) {
      const lower = $filter.toLowerCase();
      result = result.filter(a =>
        a.message?.toLowerCase().includes(lower) ||
        a.collector?.toLowerCase().includes(lower) ||
        a.severity?.toLowerCase().includes(lower)
      );
    }

    return result;
  }
);

export const realtimeLogs = writable([]);
export const serverName = writable('');
export const configPath = writable('');

export const alarmCounts = derived(alarms, ($alarms) => {
  const counts = { all: $alarms.length, critical: 0, warning: 0, info: 0 };
  for (const a of $alarms) {
    const sev = a.severity?.toLowerCase();
    if (sev in counts) counts[sev]++;
  }
  return counts;
});
