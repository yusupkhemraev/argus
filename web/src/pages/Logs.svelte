<script>
  import { X, Inbox } from 'lucide-svelte';
  import Topbar from '../components/Topbar.svelte';
  import { realtimeLogs } from '../lib/stores.js';

  let apiLogs = $state([]);
  let statusFilter = $state('all');
  let loading = $state(true);

  const tabs = [
    { id: 'all', label: 'All' },
    { id: 'sent', label: 'Sent' },
    { id: 'error', label: 'Error' },
  ];

  async function fetchLogs() {
    loading = true;
    try {
      const params = new URLSearchParams({ limit: '100' });
      if (statusFilter !== 'all') params.set('status', statusFilter);
      const res = await fetch(`/api/logs?${params}`);
      apiLogs = await res.json();
    } catch {
      apiLogs = [];
    }
    loading = false;
  }

  $effect(() => {
    statusFilter;
    fetchLogs();
  });

  let allLogs = $derived.by(() => {
    const rt = $realtimeLogs ?? [];
    const api = apiLogs ?? [];
    const seen = new Set();
    const merged = [];
    for (const entry of [...rt, ...api]) {
      const key = `${entry.ts}-${entry.notifier}-${entry.alarm_id}`;
      if (!seen.has(key)) {
        seen.add(key);
        merged.push(entry);
      }
    }
    return merged;
  });

  let filteredLogs = $derived.by(() => {
    if (statusFilter === 'all') return allLogs;
    return allLogs.filter(l => l.status === statusFilter);
  });

  let counts = $derived.by(() => {
    const all = allLogs.length;
    const sent = allLogs.filter(l => l.status === 'sent').length;
    const error = allLogs.filter(l => l.status === 'error').length;
    return { all, sent, error };
  });

  function formatTime(ts) {
    if (!ts) return '-';
    const d = new Date(ts);
    return d.toLocaleTimeString('en-GB', { hour12: false });
  }

  function channelColor(name) {
    switch (name?.toLowerCase()) {
      case 'telegram': return { color: '#818CF8', bg: '#1a1f3a' };
      case 'slack': return { color: '#38BDF8', bg: '#0f1e2a' };
      case 'webhook': return { color: '#34D399', bg: '#0d1e17' };
      default: return { color: '#6B7280', bg: '#1a1a2e' };
    }
  }

  function statusStyle(status) {
    if (status === 'sent') return { color: '#4ADE80', bg: '#14240f' };
    return { color: '#EF4444', bg: '#2a0f0f' };
  }
</script>

<Topbar title="Alarms" subtitle="logs" />

<div class="content">
  <div class="filter-row">
    <div class="pills">
      {#each tabs as tab}
        <button
          class="pill"
          class:active={statusFilter === tab.id}
          onclick={() => { statusFilter = tab.id; }}
        >
          <span>{tab.label}</span>
          <span class="pill-badge" class:active-badge={statusFilter === tab.id}>
            {counts[tab.id] ?? 0}
          </span>
        </button>
      {/each}
    </div>
    {#if statusFilter !== 'all'}
      <button class="clear-btn" onclick={() => { statusFilter = 'all'; }}>
        <X size={13} />
        <span>Clear filters</span>
      </button>
    {/if}
  </div>

  <div class="table-card">
    <div class="table-head">
      <span class="th th-time">TIME</span>
      <span class="th th-chan">CHANNEL</span>
      <span class="th th-msg">MESSAGE</span>
      <span class="th th-status">STATUS</span>
      <span class="th th-lat">LATENCY</span>
    </div>
    <div class="table-body">
      {#each filteredLogs as log}
        {@const ch = channelColor(log.notifier)}
        {@const st = statusStyle(log.status)}
        <div class="table-row">
          <span class="td td-time">{formatTime(log.ts)}</span>
          <span class="td td-chan">
            <span class="chan-badge" style="background: {ch.bg}; color: {ch.color}">{log.notifier ?? '-'}</span>
          </span>
          <span class="td td-msg">{log.alarm_id ?? '-'}</span>
          <span class="td td-status">
            <span class="status-badge" style="background: {st.bg}; color: {st.color}">{log.status}</span>
          </span>
          <span class="td td-lat" class:error-lat={log.error}>{log.error || '-'}</span>
        </div>
      {:else}
        {#if !loading}
          <div class="empty-card">
            <Inbox size={28} color="#374151" />
            <span class="empty-title">No errors found</span>
            <span class="empty-desc">All dispatch notifications are delivering successfully</span>
          </div>
        {/if}
      {/each}
    </div>
  </div>
</div>

<style>
  .content {
    flex: 1;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 20px;
    overflow: hidden;
    min-height: 0;
  }

  .filter-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-shrink: 0;
  }

  .pills {
    display: flex;
    gap: 4px;
  }

  .pill {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 5px 12px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .pill:hover {
    border-color: #2D2B55;
    background: rgba(30, 27, 75, 0.3);
  }

  .pill.active {
    background: #312E81;
    border-color: #4F46E5;
    color: #C7D2FE;
    font-weight: 500;
  }

  .pill-badge {
    font-size: 11px;
    padding: 1px 6px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.05);
  }

  .pill-badge.active-badge {
    background: #4F46E5;
  }

  .clear-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 5px 12px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-secondary);
    font-size: 13px;
    cursor: pointer;
    opacity: 0.6;
    transition: opacity 0.15s;
  }

  .clear-btn:hover {
    opacity: 1;
  }

  .table-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .table-head {
    display: flex;
    padding: 11px 20px;
    background: #0D0D14;
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .th {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-tertiary);
    letter-spacing: 1px;
  }

  .th-time { width: 130px; flex-shrink: 0; }
  .th-chan { width: 130px; flex-shrink: 0; }
  .th-msg { flex: 1; min-width: 0; }
  .th-status { width: 90px; flex-shrink: 0; }
  .th-lat { width: 80px; flex-shrink: 0; }

  .table-body {
    flex: 1;
    overflow-y: auto;
  }

  .table-row {
    display: flex;
    align-items: center;
    padding: 13px 20px;
    border-bottom: 1px solid var(--border-row);
    transition: background 0.1s;
  }

  .table-row:hover {
    background: rgba(30, 27, 75, 0.15);
  }

  .td {
    font-size: 12px;
    font-family: var(--font-mono);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .td-time { width: 130px; flex-shrink: 0; color: var(--text-secondary); }
  .td-chan { width: 130px; flex-shrink: 0; }
  .td-msg { flex: 1; min-width: 0; color: #D1D5DB; font-family: var(--font-sans); font-size: 13px; }
  .td-status { width: 90px; flex-shrink: 0; }
  .td-lat { width: 80px; flex-shrink: 0; color: var(--text-secondary); }
  .td-lat.error-lat { color: rgba(239, 68, 68, 0.5); }

  .chan-badge, .status-badge {
    font-size: 12px;
    font-family: var(--font-mono);
    padding: 2px 8px;
    border-radius: var(--radius-xs);
  }

  .status-badge {
    font-weight: 500;
  }

  .empty-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 48px 32px;
  }

  .empty-title {
    font-size: 14px;
    font-weight: 500;
    color: var(--text-tertiary);
  }

  .empty-desc {
    font-size: 12px;
    color: #374151;
  }
</style>
