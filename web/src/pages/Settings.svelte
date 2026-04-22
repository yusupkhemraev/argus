<script>
  import { onMount } from 'svelte';
  import { SlidersHorizontal, Activity, Send, Bell, Globe, Cpu, HardDrive, Info, Save, TriangleAlert, Plus, Trash2, RotateCcw, Download, Monitor } from 'lucide-svelte';
  import Topbar from '../components/Topbar.svelte';
  import { metrics, alarms, configPath } from '../lib/stores.js';

  const subNavItems = [
    { id: 'general', label: 'General', icon: SlidersHorizontal },
    { id: 'collectors', label: 'Collectors', icon: Activity },
    { id: 'notifiers', label: 'Notifiers', icon: Send },
    { id: 'history', label: 'History', icon: Download },
    { id: 'alarms', label: 'Alarms', icon: Bell },
  ];

  let activeSection = $state('general');
  let config = $state(null);
  let loading = $state(true);
  let saving = $state(false);
  let saveMsg = $state('');

  async function fetchConfig() {
    try {
      const res = await fetch('/api/config');
      config = await res.json();
    } catch {
      config = null;
    }
    loading = false;
  }

  async function saveConfig() {
    if (!config) return;
    saving = true;
    saveMsg = '';
    try {
      const res = await fetch('/api/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      if (res.ok) {
        saveMsg = 'Saved';
        setTimeout(() => { saveMsg = ''; }, 2000);
      } else {
        saveMsg = 'Error saving';
      }
    } catch {
      saveMsg = 'Error saving';
    }
    saving = false;
  }

  let testMsg = $state('');
  let testing = $state(false);

  async function testNotification() {
    testing = true;
    testMsg = '';
    try {
      const res = await fetch('/api/test-notification', { method: 'POST' });
      const data = await res.json();
      testMsg = res.ok ? 'Sent!' : `Error: ${data.errors?.join(', ')}`;
    } catch {
      testMsg = 'Connection error';
    }
    testing = false;
    setTimeout(() => { testMsg = ''; }, 3000);
  }

  async function resetAlarms() {
    try {
      await fetch('/api/reset-alarms', { method: 'POST' });
      alarms.set([]);
    } catch {}
  }

  function exportConfig() {
    if (!config) return;
    const data = JSON.stringify(config, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `argus-config-${new Date().toISOString().slice(0,10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  onMount(() => { fetchConfig(); });

  function fmtDur(ns) {
    if (!ns) return '';
    const sec = ns / 1e9;
    if (sec >= 60) return `${Math.round(sec / 60)}m`;
    return `${Math.round(sec)}s`;
  }

  function parseDur(s) {
    if (!s) return 0;
    const m = s.match(/^(\d+)(s|m|h)$/);
    if (!m) return 0;
    const v = parseInt(m[1]);
    if (m[2] === 'm') return v * 60 * 1e9;
    if (m[2] === 'h') return v * 3600 * 1e9;
    return v * 1e9;
  }

  function maskToken(s) {
    if (!s || s.length < 10) return s || '';
    return s.slice(0, 8) + '***' + s.slice(-4);
  }

  function displayVal(v) {
    if (v === undefined || v === null || v === '') return '-';
    return v;
  }

  function arrToStr(arr) {
    if (!arr || !Array.isArray(arr) || arr.length === 0) return '';
    return arr.join(', ');
  }

  function strToArr(s) {
    if (!s) return [];
    return s.split(',').map(v => v.trim()).filter(Boolean);
  }

  function addPriorityRoute() {
    if (!config.collectors.nginx.priority_routes) config.collectors.nginx.priority_routes = [];
    config.collectors.nginx.priority_routes = [...config.collectors.nginx.priority_routes, { method: 'GET', pattern: '', min_count: 1 }];
  }

  function removePriorityRoute(i) {
    config.collectors.nginx.priority_routes = config.collectors.nginx.priority_routes.filter((_, idx) => idx !== i);
  }

  let ignoreInput = $state('');

  function addIgnoreRoute() {
    const raw = ignoreInput.trim();
    if (!raw) return;
    const methods = ['GET','POST','PUT','DELETE','PATCH','HEAD','OPTIONS'];
    const parts = raw.split(/\s+/);
    let method = '*', pattern = raw;
    if (parts.length >= 2 && (methods.includes(parts[0].toUpperCase()) || parts[0] === '*')) {
      method = parts[0].toUpperCase();
      pattern = parts.slice(1).join(' ');
    }
    if (!config.collectors.nginx.ignore_routes) config.collectors.nginx.ignore_routes = [];
    config.collectors.nginx.ignore_routes = [...config.collectors.nginx.ignore_routes, { method, pattern }];
    ignoreInput = '';
  }

  function removeIgnoreRoute(i) {
    config.collectors.nginx.ignore_routes = config.collectors.nginx.ignore_routes.filter((_, idx) => idx !== i);
  }

  function ignoreRouteLabel(r) {
    return r.method === '*' ? r.pattern : `${r.method} ${r.pattern}`;
  }

  function addRabbitQueue() {
    if (!config.collectors.rabbitmq.queues) config.collectors.rabbitmq.queues = [];
    config.collectors.rabbitmq.queues = [...config.collectors.rabbitmq.queues, { name: '', threshold: 1000, unacked_threshold: 50, severity: 'warning' }];
  }

  function removeRabbitQueue(i) {
    config.collectors.rabbitmq.queues = config.collectors.rabbitmq.queues.filter((_, idx) => idx !== i);
  }

  let testCollectorMsg = $state({});
  let testCollectorLoading = $state({});

  async function testCollector(name) {
    testCollectorLoading[name] = true;
    testCollectorMsg[name] = '';
    try {
      const res = await fetch(`/api/test-collector?name=${name}`, { method: 'POST' });
      const data = await res.json();
      if (!res.ok) {
        testCollectorMsg[name] = `Error: ${data.error || 'unknown'}`;
      } else if (data.has_alarm) {
        testCollectorMsg[name] = `Alarm: ${data.alarm.message}`;
      } else {
        testCollectorMsg[name] = `OK — value: ${data.metric.value?.toFixed?.(1) ?? data.metric.value}`;
      }
    } catch {
      testCollectorMsg[name] = 'Connection error';
    }
    testCollectorLoading[name] = false;
    setTimeout(() => { testCollectorMsg[name] = ''; }, 5000);
  }

  let memVal = $derived($metrics.memory?.value ?? 0);
  let cpuVal = $derived($metrics.cpu?.value ?? 0);
  let diskVal = $derived($metrics.disk?.value ?? 0);
  let nginxVal = $derived($metrics.nginx?.value ?? 0);
  let nginxLabels = $derived($metrics.nginx?.labels ?? {});
</script>

<Topbar title="Settings" subtitle="config" />

<div class="settings-body">
  <nav class="sub-nav">
    <span class="sub-label">SECTIONS</span>
    <div class="sub-spacer"></div>
    {#each subNavItems as item}
      <button
        class="sub-item"
        class:active={activeSection === item.id}
        onclick={() => { activeSection = item.id; }}
      >
        <item.icon size={14} />
        <span>{item.label}</span>
      </button>
    {/each}
  </nav>

  <div class="settings-content">
    {#if loading}
      <div class="loading-msg">Loading config...</div>

    {:else if activeSection === 'general'}
      <div class="section-header">
        <span class="section-title">General</span>
        <span class="section-file">{$configPath || 'config.yaml'}</span>
      </div>

      <div class="card">
        <div class="card-header">
          <Monitor size={14} color="var(--text-label)" />
          <span class="card-key">name</span>
          <span class="card-desc">Server identity</span>
        </div>
        <div class="card-body">
          <div class="field-row">
            <span class="field-label">name</span>
            <span class="field-sublabel">Server name shown in alerts and sidebar (defaults to hostname)</span>
            <div class="spacer"></div>
            <input class="field-input" bind:value={config.name} placeholder="hostname" />
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <Globe size={14} color="var(--text-label)" />
          <span class="card-key">web</span>
          <span class="card-desc">HTTP server configuration</span>
        </div>
        <div class="card-body">
          <div class="field-row">
            <span class="field-label">listen</span>
            <span class="field-sublabel">Address and port for web dashboard (host:port)</span>
            <div class="spacer"></div>
            <input class="field-input" bind:value={config.server.listen} />
          </div>
          <div class="field-row">
            <span class="field-label">username</span>
            <span class="field-sublabel">Basic auth username for web dashboard</span>
            <div class="spacer"></div>
            <input class="field-input" bind:value={config.server.username} placeholder="-" />
          </div>
          <div class="field-row">
            <span class="field-label">password</span>
            <span class="field-sublabel">Basic auth password for web dashboard</span>
            <div class="spacer"></div>
            <input class="field-input" type="password" bind:value={config.server.password} placeholder="-" />
          </div>
          <div class="field-row">
            <span class="field-label">enabled</span>
            <span class="field-sublabel">Takes effect after <code>argus restart</code></span>
            <div class="spacer"></div>
            <button class="toggle" aria-label="Toggle server" class:on={config.server.enabled} onclick={() => config.server.enabled = !config.server.enabled}>
              <div class="toggle-thumb"></div>
            </button>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <Cpu size={14} color="var(--text-label)" />
          <span class="card-key">collectors</span>
          <span class="card-desc">Metric collection intervals & thresholds</span>
        </div>
        <div class="card-body no-gap">
          {#each [
            { key: 'memory', color: 'var(--bar-memory)' },
            { key: 'cpu', color: 'var(--bar-cpu)' },
            { key: 'disk', color: 'var(--bar-disk)' },
            { key: 'nginx', color: 'var(--accent-amber)' }
          ] as col, i}
            {@const c = config.collectors[col.key]}
            <div class="collector-row" class:bordered={i < 3}>
              <span class="cdot" style="background: {col.color}"></span>
              <span class="field-label">{col.key}</span>
              <div class="spacer"></div>
              <span class="field-tag">interval</span>
              <input class="field-input sm" value={fmtDur(c.interval)} onchange={(e) => { c.interval = parseDur(e.target.value); }} />
              {#if col.key !== 'nginx'}
                <span class="field-tag">threshold</span>
                <input class="field-input sm" type="number" bind:value={c.threshold} />
              {/if}
              <button class="toggle sm" aria-label="Toggle collector" class:on={c.enabled} onclick={() => c.enabled = !c.enabled}>
                <div class="toggle-thumb"></div>
              </button>
            </div>
          {/each}
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <Send size={14} color="var(--text-label)" />
          <span class="card-key">notifiers</span>
          <span class="card-desc">Notification channels</span>
        </div>
        <div class="card-body no-gap">
          <div class="collector-row">
            <span class="field-label">telegram</span>
            <div class="spacer"></div>
            <span class="field-tag">token</span>
            <input class="field-input" value={maskToken(config.notifiers.telegram.token)} readonly />
            <span class="field-tag">chat_id</span>
            <input class="field-input sm" bind:value={config.notifiers.telegram.chat_id} />
            <button class="toggle sm" aria-label="Toggle telegram" class:on={config.notifiers.telegram.enabled} onclick={() => config.notifiers.telegram.enabled = !config.notifiers.telegram.enabled}>
              <div class="toggle-thumb"></div>
            </button>
          </div>
        </div>
      </div>

    {:else if activeSection === 'collectors'}
      <div class="section-header">
        <span class="section-title">Collectors</span>
        <span class="section-file">live status</span>
      </div>

      {#each [
        { key: 'memory', desc: 'System memory usage', color: 'var(--bar-memory)', icon: Cpu, liveVal: memVal },
        { key: 'cpu', desc: 'CPU utilization', color: 'var(--bar-cpu)', icon: Cpu, liveVal: cpuVal },
        { key: 'disk', desc: 'Disk space usage — path: ' + (config.collectors.disk.path || '/'), color: 'var(--bar-disk)', icon: HardDrive, liveVal: diskVal }
      ] as col}
        {@const c = config.collectors[col.key]}
        <div class="card">
          <div class="card-header">
            <span class="cdot" style="background: {col.color}"></span>
            <span class="card-key">{col.key}</span>
            <span class="card-desc">{col.desc}</span>
            <div class="spacer"></div>
            <button class="toggle" aria-label="Toggle collector" class:on={c.enabled} onclick={() => c.enabled = !c.enabled}>
              <div class="toggle-thumb"></div>
            </button>
          </div>
          <div class="card-body">
            <div class="detail-row">
              <span class="detail-label">Interval</span>
              <span class="field-hint">How often to check and send alert if threshold exceeded</span>
              <div class="spacer"></div>
              <input class="field-input sm" value={fmtDur(c.interval)} onchange={(e) => { c.interval = parseDur(e.target.value); }} />
            </div>
            <div class="detail-row">
              <span class="detail-label">Refresh</span>
              <span class="field-hint">How often to update the dashboard value</span>
              <div class="spacer"></div>
              <input class="field-input sm" value={fmtDur(c.refresh)} onchange={(e) => { c.refresh = parseDur(e.target.value); }} />
            </div>
            <div class="detail-row">
              <span class="detail-label">Threshold</span>
              <span class="field-hint">Alert when usage exceeds this percentage</span>
              <div class="spacer"></div>
              <input class="field-input sm" type="number" bind:value={c.threshold} />
            </div>
            <div class="detail-row">
              <span class="detail-label">Severity</span>
              <span class="field-hint">Alarm severity level</span>
              <div class="spacer"></div>
              <select class="field-input sm" bind:value={c.severity}>
                <option value="info">info</option>
                <option value="warning">warning</option>
                <option value="critical">critical</option>
              </select>
            </div>
            <div class="live-bar-wrap">
              <span class="live-bar-label">Live value</span>
              <div class="live-bar-track">
                <div class="live-bar-fill" style="width: {col.liveVal}%; background: {col.color}"></div>
              </div>
              <span class="live-bar-val" style="color: {col.color}">{col.liveVal.toFixed(1)}%</span>
            </div>
            <div class="test-row">
              <button class="btn-outline" onclick={() => testCollector(col.key)} disabled={testCollectorLoading[col.key]}>
                {testCollectorLoading[col.key] ? 'Testing...' : 'Test'}
              </button>
              {#if testCollectorMsg[col.key]}
                <span class="test-result" class:alarm={testCollectorMsg[col.key]?.startsWith('Alarm')}>{testCollectorMsg[col.key]}</span>
              {/if}
            </div>
          </div>
        </div>
      {/each}

      <div class="card">
        <div class="card-header">
          <span class="cdot" style="background: var(--accent-amber)"></span>
          <span class="card-key">nginx</span>
          <span class="card-desc">Access log error & slow request monitoring</span>
          <div class="spacer"></div>
          <button class="toggle" aria-label="Toggle nginx" class:on={config.collectors.nginx.enabled} onclick={() => config.collectors.nginx.enabled = !config.collectors.nginx.enabled}>
            <div class="toggle-thumb"></div>
          </button>
        </div>
        <div class="card-body">
          <div class="detail-row">
            <span class="detail-label">Access log</span>
            <span class="field-hint">Path to nginx access log file</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.collectors.nginx.access_log} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Log parser</span>
            <span class="field-hint">Looks for <code>"METHOD /path HTTP/x.x" STATUS</code> — works with any nginx format that includes <code>$request</code> and <code>$status</code>. Timings extracted from <code>rt:</code> / <code>request_time:</code> prefixes.</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">Watch statuses</span>
            <span class="field-hint">HTTP status codes/ranges to track (comma-separated)</span>
            <div class="spacer"></div>
            <input class="field-input wide" value={arrToStr(config.collectors.nginx.watch_statuses)} onchange={(e) => { config.collectors.nginx.watch_statuses = strToArr(e.target.value); }} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Interval</span>
            <span class="field-hint">How often to check and send alert</span>
            <div class="spacer"></div>
            <input class="field-input sm" value={fmtDur(config.collectors.nginx.interval)} onchange={(e) => { config.collectors.nginx.interval = parseDur(e.target.value); }} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Refresh</span>
            <span class="field-hint">How often to update dashboard counters</span>
            <div class="spacer"></div>
            <input class="field-input sm" value={fmtDur(config.collectors.nginx.refresh)} onchange={(e) => { config.collectors.nginx.refresh = parseDur(e.target.value); }} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Window</span>
            <span class="field-hint">Time window for counting errors (alerts)</span>
            <div class="spacer"></div>
            <input class="field-input sm" value={fmtDur(config.collectors.nginx.window)} onchange={(e) => { config.collectors.nginx.window = parseDur(e.target.value); }} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Error threshold</span>
            <span class="field-hint">Min errors in window to trigger alert</span>
            <div class="spacer"></div>
            <input class="field-input sm" type="number" bind:value={config.collectors.nginx.threshold} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Slow threshold</span>
            <span class="field-hint">Request time in seconds to consider slow</span>
            <div class="spacer"></div>
            <input class="field-input sm" type="number" bind:value={config.collectors.nginx.slow_threshold} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Slow window</span>
            <span class="field-hint">Time window for counting slow requests</span>
            <div class="spacer"></div>
            <input class="field-input sm" value={fmtDur(config.collectors.nginx.slow_window)} onchange={(e) => { config.collectors.nginx.slow_window = parseDur(e.target.value); }} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Slow count</span>
            <span class="field-hint">Min slow requests to trigger alert</span>
            <div class="spacer"></div>
            <input class="field-input sm" type="number" bind:value={config.collectors.nginx.slow_count} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Min group count</span>
            <span class="field-hint">Min errors per endpoint to include in alert</span>
            <div class="spacer"></div>
            <input class="field-input sm" type="number" bind:value={config.collectors.nginx.min_group_count} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Severity</span>
            <span class="field-hint">Alarm severity level</span>
            <div class="spacer"></div>
            <select class="field-input sm" bind:value={config.collectors.nginx.severity}>
              <option value="info">info</option>
              <option value="warning">warning</option>
              <option value="critical">critical</option>
            </select>
          </div>
          <div class="detail-section">
            <div class="detail-section-header">
              <span class="detail-label">Priority routes</span>
              <span class="field-hint">Routes with lower alert thresholds</span>
              <div class="spacer"></div>
              <button class="btn-outline sm" onclick={addPriorityRoute}><Plus size={12} /><span>Add</span></button>
            </div>
            {#if config.collectors.nginx.priority_routes?.length > 0}
              {#each config.collectors.nginx.priority_routes as route, i}
                <div class="route-row">
                  <select class="field-input xs" bind:value={route.method}>
                    <option>GET</option><option>POST</option><option>PUT</option><option>DELETE</option><option>PATCH</option><option>*</option>
                  </select>
                  <input class="field-input wide" bind:value={route.pattern} placeholder="/api/v1/..." />
                  <span class="field-tag">min</span>
                  <input class="field-input xs" type="number" bind:value={route.min_count} />
                  <span class="field-tag">exclude</span>
                  <input class="field-input" value={arrToStr(route.exclude_statuses)} onchange={(e) => { route.exclude_statuses = strToArr(e.target.value); }} placeholder="499, 408" />
                  <button class="icon-btn" onclick={() => removePriorityRoute(i)}><Trash2 size={12} /></button>
                </div>
              {/each}
            {:else}
              <span class="empty-hint">No priority routes configured</span>
            {/if}
          </div>
          <div class="detail-section">
            <div class="detail-section-header">
              <span class="detail-label">Ignore routes</span>
              <span class="field-hint">Never counted as errors or slow</span>
            </div>
            <div class="ignore-input-row">
              <input
                class="field-input wide"
                bind:value={ignoreInput}
                placeholder="/health  or  GET /metrics  or  * /ping/*"
                onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); addIgnoreRoute(); } }}
              />
              <button class="btn-outline sm" onclick={addIgnoreRoute}><Plus size={12} /><span>Add</span></button>
            </div>
            {#if config.collectors.nginx.ignore_routes?.length > 0}
              <div class="ignore-chips">
                {#each config.collectors.nginx.ignore_routes as route, i}
                  <span class="ignore-chip">
                    <span class="chip-label">{ignoreRouteLabel(route)}</span>
                    <button class="chip-remove" onclick={() => removeIgnoreRoute(i)}>×</button>
                  </span>
                {/each}
              </div>
            {/if}
          </div>
          <div class="live-stats">
            <div class="live-stat">
              <span class="live-stat-label">HTTP errors</span>
              <span class="live-stat-val" style="color: var(--accent-amber)">{nginxLabels.http_errors ?? 0}</span>
            </div>
            <div class="live-stat">
              <span class="live-stat-label">Slow requests</span>
              <span class="live-stat-val" style="color: var(--accent-amber)">{nginxLabels.slow_requests ?? 0}</span>
            </div>
          </div>
          <div class="test-row">
            <button class="btn-outline" onclick={() => testCollector('nginx')} disabled={testCollectorLoading['nginx']}>
              {testCollectorLoading['nginx'] ? 'Testing...' : 'Test'}
            </button>
            {#if testCollectorMsg['nginx']}
              <span class="test-result" class:alarm={testCollectorMsg['nginx']?.startsWith('Alarm')}>{testCollectorMsg['nginx']}</span>
            {/if}
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <span class="cdot" style="background: var(--accent-purple)"></span>
          <span class="card-key">rabbitmq</span>
          <span class="card-desc">RabbitMQ queue depth monitoring</span>
          <div class="spacer"></div>
          <button class="toggle" aria-label="Toggle rabbitmq" class:on={config.collectors.rabbitmq.enabled} onclick={() => config.collectors.rabbitmq.enabled = !config.collectors.rabbitmq.enabled}>
            <div class="toggle-thumb"></div>
          </button>
        </div>
        <div class="card-body">
          <div class="detail-row">
            <span class="detail-label">Management URL</span>
            <span class="field-hint">RabbitMQ management API endpoint</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.collectors.rabbitmq.management_url} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Username</span>
            <span class="field-hint">RabbitMQ management credentials</span>
            <div class="spacer"></div>
            <input class="field-input" bind:value={config.collectors.rabbitmq.username} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Password</span>
            <div class="spacer"></div>
            <input class="field-input" type="password" bind:value={config.collectors.rabbitmq.password} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Interval</span>
            <span class="field-hint">How often to check queue depths</span>
            <div class="spacer"></div>
            <input class="field-input sm" value={fmtDur(config.collectors.rabbitmq.interval)} onchange={(e) => { config.collectors.rabbitmq.interval = parseDur(e.target.value); }} />
          </div>
          <div class="detail-section">
            <div class="detail-section-header">
              <span class="detail-label">Queues</span>
              <span class="field-hint">Monitored queues with thresholds</span>
              <div class="spacer"></div>
              <button class="btn-outline sm" onclick={addRabbitQueue}><Plus size={12} /><span>Add</span></button>
            </div>
            {#if config.collectors.rabbitmq.queues?.length > 0}
              {#each config.collectors.rabbitmq.queues as queue, i}
                <div class="route-row">
                  <input class="field-input" bind:value={queue.name} placeholder="queue name" />
                  <span class="field-tag">max</span>
                  <input class="field-input xs" type="number" bind:value={queue.threshold} />
                  <span class="field-tag">unacked</span>
                  <input class="field-input xs" type="number" bind:value={queue.unacked_threshold} />
                  <select class="field-input xs" bind:value={queue.severity}>
                    <option value="info">info</option>
                    <option value="warning">warning</option>
                    <option value="critical">critical</option>
                  </select>
                  <button class="icon-btn" onclick={() => removeRabbitQueue(i)}><Trash2 size={12} /></button>
                </div>
              {/each}
            {:else}
              <span class="empty-hint">No queues configured</span>
            {/if}
          </div>
          <div class="test-row">
            <button class="btn-outline" onclick={() => testCollector('rabbitmq')} disabled={testCollectorLoading['rabbitmq']}>
              {testCollectorLoading['rabbitmq'] ? 'Testing...' : 'Test'}
            </button>
            {#if testCollectorMsg['rabbitmq']}
              <span class="test-result" class:alarm={testCollectorMsg['rabbitmq']?.startsWith('Alarm')}>{testCollectorMsg['rabbitmq']}</span>
            {/if}
          </div>
        </div>
      </div>

    {:else if activeSection === 'notifiers'}
      <div class="section-header">
        <span class="section-title">Notifiers</span>
        <span class="section-file">notification channels</span>
      </div>

      <div class="card">
        <div class="card-header">
          <Send size={14} color="var(--text-label)" />
          <span class="card-key">Telegram</span>
          <span class="status-chip" class:enabled={config.notifiers.telegram.enabled}>{config.notifiers.telegram.enabled ? 'active' : 'disabled'}</span>
          <div class="spacer"></div>
          <button class="toggle" aria-label="Toggle telegram" class:on={config.notifiers.telegram.enabled} onclick={() => config.notifiers.telegram.enabled = !config.notifiers.telegram.enabled}>
            <div class="toggle-thumb"></div>
          </button>
        </div>
        <div class="card-body">
          <div class="detail-row">
            <span class="detail-label">bot_token</span>
            <span class="field-hint">Telegram Bot API token from @BotFather</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.notifiers.telegram.token} />
          </div>
          <div class="detail-row">
            <span class="detail-label">chat_id</span>
            <span class="field-hint">Target chat/group ID for alerts</span>
            <div class="spacer"></div>
            <input class="field-input" bind:value={config.notifiers.telegram.chat_id} />
          </div>
          <div class="detail-row">
            <span class="detail-label">mentions</span>
            <span class="field-hint">Telegram usernames to mention in alerts (comma-separated)</span>
            <div class="spacer"></div>
            <input class="field-input wide" value={arrToStr(config.notifiers.telegram.mentions)} onchange={(e) => { config.notifiers.telegram.mentions = strToArr(e.target.value); }} placeholder="-" />
          </div>
          <div class="detail-row">
            <button class="btn-accent" onclick={testNotification} disabled={testing}>
              {testing ? 'Sending...' : 'Test Notification'}
            </button>
            {#if testMsg}<span class="test-msg">{testMsg}</span>{/if}
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <Globe size={14} color="var(--text-label)" />
          <span class="card-key">Webhook</span>
          <span class="status-chip" class:enabled={config.notifiers.webhook.enabled}>{config.notifiers.webhook.enabled ? 'active' : 'disabled'}</span>
          <div class="spacer"></div>
          <button class="toggle" aria-label="Toggle webhook" class:on={config.notifiers.webhook.enabled} onclick={() => config.notifiers.webhook.enabled = !config.notifiers.webhook.enabled}>
            <div class="toggle-thumb"></div>
          </button>
        </div>
        <div class="card-body">
          <div class="detail-row">
            <span class="detail-label">url</span>
            <span class="field-hint">Webhook endpoint URL</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.notifiers.webhook.url} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">method</span>
            <span class="field-hint">HTTP method</span>
            <div class="spacer"></div>
            <select class="field-input sm" bind:value={config.notifiers.webhook.method}>
              <option>POST</option>
              <option>PUT</option>
              <option>PATCH</option>
            </select>
          </div>
          <div class="detail-row">
            <span class="detail-label">api_key</span>
            <span class="field-hint">API key value (optional)</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.notifiers.webhook.api_key} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">api_key_header</span>
            <span class="field-hint">Header name for API key (default: Authorization)</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.notifiers.webhook.api_key_header} placeholder="Authorization" />
          </div>
          <div class="detail-row col">
            <span class="detail-label">payload example</span>
            <pre class="json-example">{JSON.stringify({
  server: "prod-server",
  collector: "cpu",
  severity: "critical",
  message: "CPU 91.0% / 90.0%",
  value: 91.0,
  threshold: 90.0,
  timestamp: "2026-04-19T17:00:00Z"
}, null, 2)}</pre>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <Info size={14} color="var(--text-label)" />
          <span class="card-key">Notification log</span>
          <span class="card-desc">Log file for sent notifications</span>
        </div>
        <div class="card-body">
          <div class="detail-row">
            <span class="detail-label">log_path</span>
            <span class="field-hint">Path to notification log file (leave empty to disable)</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.notifiers.log_path} placeholder="-" />
          </div>
        </div>
      </div>

    {:else if activeSection === 'history'}
      <div class="section-header">
        <span class="section-title">History</span>
        <span class="section-file">alarm history storage</span>
      </div>

      <div class="card">
        <div class="card-header">
          <Download size={14} color="var(--text-label)" />
          <span class="card-key">History</span>
          <span class="card-desc">Alarm history log file</span>
          <div class="spacer"></div>
          <button class="toggle" aria-label="Toggle history" class:on={config.history.enabled} onclick={() => config.history.enabled = !config.history.enabled}>
            <div class="toggle-thumb"></div>
          </button>
        </div>
        <div class="card-body">
          <div class="detail-row">
            <span class="detail-label">File path</span>
            <span class="field-hint">Path to alarm history log file</span>
            <div class="spacer"></div>
            <input class="field-input wide" bind:value={config.history.file_path} placeholder="-" />
          </div>
          <div class="detail-row">
            <span class="detail-label">Max size (MB)</span>
            <span class="field-hint">Maximum file size before rotation</span>
            <div class="spacer"></div>
            <input class="field-input sm" type="number" bind:value={config.history.max_size_mb} />
          </div>
          <div class="detail-row">
            <span class="detail-label">Max backups</span>
            <span class="field-hint">Number of rotated backup files to keep</span>
            <div class="spacer"></div>
            <input class="field-input sm" type="number" bind:value={config.history.max_backups} />
          </div>
        </div>
      </div>

    {:else if activeSection === 'alarms'}
      <div class="section-header">
        <span class="section-title">Alarms</span>
        <span class="section-file">{4} rules configured</span>
      </div>

      <div class="card">
        <div class="card-header">
          <TriangleAlert size={14} color="var(--text-label)" />
          <span class="card-key">Alarm Rules</span>
          <div class="spacer"></div>
          <button class="btn-outline" onclick={() => { activeSection = 'collectors'; }}><Plus size={13} /><span>Add Rule</span></button>
        </div>
        <div class="card-body no-gap">
          <div class="rules-head">
            <span class="rh" style="width:120px">NAME</span>
            <span class="rh" style="width:100px">COLLECTOR</span>
            <span class="rh" style="width:80px">THRESHOLD</span>
            <span class="rh" style="width:80px">SEVERITY</span>
            <span class="rh" style="flex:1">DESCRIPTION</span>
          </div>
          {#each [
            { name: 'memory-high', col: 'memory', color: 'var(--bar-memory)', desc: 'Memory usage exceeded threshold', threshFmt: (c) => `> ${c.threshold}%` },
            { name: 'cpu-high', col: 'cpu', color: 'var(--bar-cpu)', desc: 'CPU usage exceeded threshold', threshFmt: (c) => `> ${c.threshold}%` },
            { name: 'disk-high', col: 'disk', color: 'var(--bar-disk)', desc: 'Disk space critically low', threshFmt: (c) => `> ${c.threshold}%` },
            { name: 'nginx-errors', col: 'nginx', color: 'var(--accent-amber)', desc: 'HTTP errors or slow requests detected', threshFmt: (c) => `> ${c.threshold} errs` }
          ] as rule}
            {@const c = config.collectors[rule.col]}
            <div class="rule-row">
              <span class="rc mono" style="width:120px">{rule.name}</span>
              <span class="rc" style="width:100px"><span class="cdot sm" style="background:{rule.color}"></span>{rule.col}</span>
              <span class="rc" style="width:80px"><span class="thresh-pill">{rule.threshFmt(c)}</span></span>
              <span class="rc" style="width:80px"><span class="sev-pill {c.severity}">{c.severity}</span></span>
              <span class="rc desc" style="flex:1">{rule.desc}</span>
              <button class="icon-btn" onclick={() => { config.collectors[rule.col].enabled = false; }}><Trash2 size={13} /></button>
            </div>
          {/each}
        </div>
      </div>

      <div class="card danger">
        <div class="card-header">
          <TriangleAlert size={14} color="var(--accent-red)" />
          <span class="card-key" style="color:var(--accent-red)">Danger Zone</span>
        </div>
        <div class="card-body horizontal">
          <div class="danger-action">
            <span class="danger-desc">Reset all active alarms to initial state</span>
            <button class="btn-danger" onclick={resetAlarms}><RotateCcw size={13} /><span>Reset Alarms</span></button>
          </div>
          <div class="danger-action">
            <span class="danger-desc">Export full configuration as YAML file</span>
            <button class="btn-outline" onclick={exportConfig}><Download size={13} /><span>Export Config</span></button>
          </div>
        </div>
      </div>
    {/if}

    <div class="footer">
      <div class="footer-hint">
        <Info size={12} color="#3A3A5C" />
        <span>Changes are written to {$configPath || 'config.yaml'} on save</span>
      </div>
      <div class="footer-right">
        {#if saveMsg}
          <span class="save-msg">{saveMsg}</span>
        {/if}
        <button class="save-btn" onclick={saveConfig} disabled={saving}>
          <Save size={14} />
          <span>{saving ? 'Saving...' : 'Save changes'}</span>
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .settings-body { flex:1; display:flex; min-height:0; overflow:hidden; }

  .sub-nav {
    width:200px; background:var(--bg-sidebar); border-right:1px solid var(--border);
    padding:20px 12px; display:flex; flex-direction:column; gap:2px; flex-shrink:0;
  }
  .sub-label { font-size:10px; font-weight:600; color:#3A3A5C; letter-spacing:2px; padding:0 10px; }
  .sub-spacer { height:8px; }
  .sub-item {
    display:flex; align-items:center; gap:8px; padding:7px 10px; border:none; background:transparent;
    color:var(--text-nav); font-size:13px; border-radius:var(--radius-sm); cursor:pointer;
    width:100%; text-align:left; transition:background .15s, color .15s;
  }
  .sub-item:hover { background:rgba(30,27,75,.5); color:var(--text-nav-active); }
  .sub-item.active { background:var(--bg-active); color:var(--text-nav-active); font-weight:500; }

  .settings-content {
    flex:1; padding:28px 32px; display:flex; flex-direction:column; gap:24px; overflow-y:auto; min-width:0;
  }
  .loading-msg { color:var(--text-secondary); padding:32px; }

  .section-header { display:flex; justify-content:space-between; align-items:center; }
  .section-title { font-size:18px; font-weight:600; color:var(--text-primary); }
  .section-file { font-size:12px; font-family:var(--font-mono); color:#3A3A5C; }

  .card { background:var(--bg-elevated); border:1px solid var(--border); border-radius:var(--radius); }
  .card.danger { border-color:rgba(239,68,68,.2); }

  .card-header {
    display:flex; align-items:center; gap:8px; padding:14px 20px; border-bottom:1px solid var(--border);
  }
  .card-key { font-size:13px; font-weight:700; font-family:var(--font-mono); color:var(--text-primary); }
  .card-desc { font-size:12px; color:var(--text-label); }

  .card-body { padding:20px; display:flex; flex-direction:column; gap:16px; }
  .card-body.no-gap { gap:0; }
  .card-body.horizontal { flex-direction:row; gap:32px; }

  .spacer { flex:1; }

  .field-row { display:flex; align-items:center; gap:16px; }
  .field-label { font-size:13px; font-weight:500; color:var(--text-primary); }
  .field-sublabel { font-size:12px; color:var(--text-label); }
  .field-tag { font-size:11px; color:var(--text-label); }

  .field-input {
    font-size:13px; color:var(--text-primary); background:var(--bg-input); border:1px solid var(--border);
    padding:5px 12px; border-radius:var(--radius-xs); font-family:var(--font-mono); outline:none;
    transition:border-color .15s; width:160px;
  }
  .field-input:focus { border-color:var(--accent-purple); }
  .field-input.sm { width:100px; text-align:center; }
  .field-input.wide { width:280px; }

  .toggle {
    width:36px; height:20px; border-radius:10px; background:var(--border); position:relative;
    cursor:pointer; transition:background .2s; flex-shrink:0; border:none;
  }
  .toggle.on { background:#6366F1; }
  .toggle.sm { width:32px; height:18px; border-radius:9px; }
  .toggle-thumb {
    width:16px; height:16px; border-radius:50%; background:white; position:absolute; top:2px; left:2px; transition:left .2s;
  }
  .toggle.on .toggle-thumb { left:18px; }
  .toggle.sm .toggle-thumb { width:14px; height:14px; }
  .toggle.sm.on .toggle-thumb { left:16px; }

  .collector-row {
    display:flex; align-items:center; gap:16px; padding:16px 0;
  }
  .collector-row.bordered { border-bottom:1px solid var(--border); }

  .cdot { width:8px; height:8px; border-radius:50%; flex-shrink:0; }
  .cdot.sm { width:6px; height:6px; }

  .detail-row { display:flex; align-items:center; gap:12px; }
  .detail-label { font-size:13px; color:var(--text-secondary); white-space:nowrap; }
  .field-hint { font-size:11px; color:#3A3A5C; }

  .detail-section { display:flex; flex-direction:column; gap:8px; padding-top:8px; border-top:1px solid var(--border); margin-top:4px; }
  .detail-section-header { display:flex; align-items:center; gap:12px; }
  .route-row { display:flex; align-items:center; gap:8px; padding:4px 0; }
  .empty-hint { font-size:12px; color:var(--text-tertiary); padding:4px 0; }
  .detail-row.col { flex-direction:column; align-items:flex-start; gap:8px; }
  .json-example { margin:0; padding:10px 12px; background:var(--bg-tertiary); border-radius:6px; font-size:11px; color:var(--text-secondary); font-family:monospace; line-height:1.6; width:100%; box-sizing:border-box; white-space:pre; overflow-x:auto; }
  .field-input.xs { width:80px; text-align:center; }
  input[type=number]::-webkit-inner-spin-button,
  input[type=number]::-webkit-outer-spin-button { -webkit-appearance:none; margin:0; }
  input[type=number] { -moz-appearance:textfield; }
  .btn-outline.sm { padding:4px 10px; font-size:11px; }

  .ignore-input-row { display:flex; align-items:center; gap:8px; }
  .ignore-chips { display:flex; flex-wrap:wrap; gap:6px; }
  .ignore-chip {
    display:inline-flex; align-items:center; gap:4px; padding:3px 8px 3px 10px;
    background:var(--bg-input); border:1px solid var(--border); border-radius:20px;
    font-size:12px; font-family:var(--font-mono); color:var(--text-secondary);
  }
  .chip-label { line-height:1; }
  .chip-remove {
    background:none; border:none; color:var(--text-tertiary); cursor:pointer;
    font-size:14px; line-height:1; padding:0 2px; display:flex; align-items:center;
  }
  .chip-remove:hover { color:var(--accent-red); }

  .test-row { display:flex; align-items:center; gap:12px; padding-top:8px; border-top:1px solid var(--border); margin-top:4px; }
  .test-result { font-size:12px; color:var(--accent-green); font-family:var(--font-mono); }
  .test-result.alarm { color:var(--accent-amber); }

  .live-stats { display:flex; gap:24px; padding-top:8px; border-top:1px solid var(--border); margin-top:4px; }
  .live-stat { display:flex; align-items:center; gap:8px; }
  .live-stat-label { font-size:12px; color:var(--text-label); }
  .live-stat-val { font-size:14px; font-weight:600; font-family:var(--font-mono); }

  .live-bar-wrap { display:flex; align-items:center; gap:12px; margin-top:4px; }
  .live-bar-label { font-size:11px; color:var(--text-label); white-space:nowrap; }
  .live-bar-track { flex:1; height:6px; background:var(--border); border-radius:3px; overflow:hidden; }
  .live-bar-fill { height:100%; border-radius:3px; transition:width .3s ease; }
  .live-bar-val { font-size:12px; font-weight:600; font-family:var(--font-mono); min-width:50px; text-align:right; }

  .status-chip {
    font-size:11px; padding:2px 8px; border-radius:var(--radius-xs); background:#2a0f0f; color:var(--accent-red);
  }
  .status-chip.enabled { background:#14240f; color:#4ADE80; }

  .btn-accent {
    padding:7px 16px; border-radius:var(--radius-sm); background:#6366F1; border:none;
    color:white; font-size:12px; font-weight:600; cursor:pointer;
  }
  .btn-accent:hover { background:#5558E6; }

  .btn-outline {
    display:flex; align-items:center; gap:6px; padding:6px 14px; border-radius:var(--radius-sm);
    background:transparent; border:1px solid var(--border); color:var(--text-label); font-size:12px; cursor:pointer;
  }
  .btn-outline:hover { color:var(--text-primary); border-color:var(--accent-purple); }

  .btn-danger {
    display:flex; align-items:center; gap:6px; padding:6px 14px; border-radius:var(--radius-sm);
    background:transparent; border:1px solid rgba(239,68,68,.3); color:var(--accent-red); font-size:12px; cursor:pointer;
  }
  .btn-danger:hover { background:rgba(239,68,68,.1); }


  .rules-head {
    display:flex; gap:12px; padding:12px 0; border-bottom:1px solid var(--border);
  }
  .rh { font-size:10px; font-weight:600; color:var(--text-tertiary); letter-spacing:1px; }

  .rule-row {
    display:flex; align-items:center; gap:12px; padding:12px 0; border-bottom:1px solid var(--border);
  }
  .rule-row:last-child { border-bottom:none; }
  .rc { font-size:13px; color:var(--text-primary); display:flex; align-items:center; gap:6px; }
  .rc.mono { font-family:var(--font-mono); }
  .rc.desc { color:var(--text-secondary); font-size:12px; }

  .thresh-pill {
    font-size:12px; font-family:var(--font-mono); padding:2px 8px; border-radius:var(--radius-xs);
    background:var(--bg-input); color:var(--text-primary);
  }
  .sev-pill { font-size:11px; font-weight:600; padding:2px 8px; border-radius:var(--radius-xs); }
  .sev-pill.warning { background:rgba(245,158,11,.1); color:var(--accent-amber); }
  .sev-pill.critical { background:rgba(239,68,68,.1); color:var(--accent-red); }
  .sev-pill.info { background:rgba(16,185,129,.1); color:var(--accent-green); }

  .icon-btn {
    background:none; border:none; color:var(--text-tertiary); cursor:pointer; padding:4px;
  }
  .icon-btn:hover { color:var(--accent-red); }

  .danger-action { display:flex; flex-direction:column; gap:8px; }
  .danger-desc { font-size:12px; color:var(--text-secondary); }

  .footer {
    display:flex; justify-content:space-between; align-items:center; padding-top:16px; margin-top:auto;
  }
  .footer-hint { display:flex; align-items:center; gap:6px; font-size:11px; color:#3A3A5C; }
  .footer-right { display:flex; align-items:center; gap:12px; }
  .save-msg { font-size:12px; color:var(--accent-green); }

  .save-btn {
    display:flex; align-items:center; gap:6px; padding:8px 20px; border-radius:var(--radius-sm);
    background:#6366F1; border:none; color:white; font-size:13px; font-weight:600; cursor:pointer;
  }
  .save-btn:hover { background:#5558E6; }
  .save-btn:disabled { opacity:.6; cursor:not-allowed; }

  .test-msg { font-size:12px; color:var(--accent-green); margin-left:12px; }
</style>
