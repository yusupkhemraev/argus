import { metrics, alarms, connected, realtimeLogs, serverName, configPath } from './stores.js';

let es = null;

async function loadInitialData() {
  try {
    const [alarmsRes, configRes] = await Promise.all([
      fetch('/api/alarms?limit=200'),
      fetch('/api/config')
    ]);
    const alarmsData = await alarmsRes.json();
    if (Array.isArray(alarmsData) && alarmsData.length > 0) {
      alarms.set(alarmsData.reverse());
    }
    const configData = await configRes.json();
    if (configData?.name) {
      serverName.set(configData.name);
    }
    if (configData?.config_path) {
      configPath.set(configData.config_path);
    }
  } catch {}
}

export function connect() {
  es = new EventSource('/api/events');

  es.addEventListener('metric', (e) => {
    const data = JSON.parse(e.data);
    metrics.update(m => ({ ...m, [data.collector]: data }));
  });

  es.addEventListener('alarm', (e) => {
    const data = JSON.parse(e.data);
    alarms.update(list => [data, ...list].slice(0, 200));
  });

  es.addEventListener('log', (e) => {
    const data = JSON.parse(e.data);
    realtimeLogs.update(list => [data, ...list].slice(0, 200));
  });

  es.onopen = () => {
    connected.set(true);
    loadInitialData();
  };
  es.onerror = () => connected.set(false);
}

export function disconnect() {
  if (es) {
    es.close();
    es = null;
    connected.set(false);
  }
}
