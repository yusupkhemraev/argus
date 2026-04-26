<script>
  import { Search, Download } from 'lucide-svelte';
  import Topbar from '../components/Topbar.svelte';
  import { filteredAlarms, alarmCounts, filter, severityFilter } from '../lib/stores.js';

  const tabs = [
    { id: 'all', label: 'All', color: null },
    { id: 'critical', label: 'Critical', color: 'var(--accent-red)' },
    { id: 'warning', label: 'Warning', color: 'var(--accent-amber)' },
    { id: 'info', label: 'Info', color: 'var(--accent-green)' },
  ];

  let searchText = $state('');

  $effect(() => { filter.set(searchText); });

  function setTab(id) {
    severityFilter.set(id);
  }

  function exportAlarms() {
    const data = JSON.stringify($filteredAlarms, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `argus-alarms-${new Date().toISOString().slice(0,10)}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function severityColor(severity) {
    switch (severity?.toLowerCase()) {
      case 'critical': return 'var(--accent-red)';
      case 'warning': return 'var(--accent-amber)';
      case 'info': return 'var(--accent-green)';
      default: return 'var(--text-secondary)';
    }
  }

  function severityBg(severity) {
    switch (severity?.toLowerCase()) {
      case 'critical': return 'rgba(239, 68, 68, 0.1)';
      case 'warning': return 'rgba(245, 158, 11, 0.1)';
      case 'info': return 'rgba(16, 185, 129, 0.1)';
      default: return 'transparent';
    }
  }

  function formatTime(ts) {
    if (!ts) return '-';
    const d = new Date(ts);
    return d.toLocaleTimeString('en-GB', { hour12: false });
  }
</script>

<Topbar title="Alarms" subtitle="history" />

<div class="content">
  <div class="toolbar">
    <div class="filter-tabs">
      {#each tabs as tab}
        <button
          class="tab"
          class:active={$severityFilter === tab.id}
          onclick={() => setTab(tab.id)}
        >
          {#if tab.color}
            <span class="tab-dot" style="background: {tab.color}"></span>
          {/if}
          <span>{tab.label}</span>
          <span class="tab-badge" class:active-badge={$severityFilter === tab.id}>
            {$alarmCounts[tab.id] ?? 0}
          </span>
        </button>
      {/each}
    </div>

    <div class="right-tools">
      <div class="search-box">
        <Search size={13} color="var(--text-label)" />
        <input
          class="search-input"
          bind:value={searchText}
          placeholder="Filter alarms..."
        />
      </div>
      <button class="export-btn" onclick={exportAlarms}>
        <Download size={13} />
        <span>Export</span>
      </button>
    </div>
  </div>

  <div class="table-card">
    <div class="table-head">
      <span class="th th-time">TIME</span>
      <span class="th th-sev">SEVERITY</span>
      <span class="th th-col">COLLECTOR</span>
      <span class="th th-msg">MESSAGE</span>
      <span class="th th-val">VALUE</span>
      <span class="th th-thresh">THRESHOLD</span>
    </div>
    <div class="table-body">
      {#each $filteredAlarms as alarm, i}
        <div class="table-row" class:highlighted={i === 0}>
          <span class="td td-time">{formatTime(alarm.triggered_at)}</span>
          <span class="td td-sev">
            <span class="sev-dot" style="background: {severityColor(alarm.severity)}"></span>
            <span class="sev-badge" style="background: {severityBg(alarm.severity)}; color: {severityColor(alarm.severity)}">{alarm.severity}</span>
          </span>
          <span class="td td-col">{alarm.collector ?? '-'}</span>
          <span class="td td-msg">{alarm.message}</span>
          <span class="td td-val" style="color: {severityColor(alarm.severity)}">{alarm.value != null ? alarm.value.toFixed(1) : '-'}</span>
          <span class="td td-thresh">{alarm.threshold != null ? alarm.threshold.toFixed(1) : '-'}</span>
        </div>
      {:else}
        <div class="empty-state">no alarms</div>
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
    gap: 16px;
    overflow: hidden;
    min-height: 0;
  }

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-shrink: 0;
    gap: 12px;
  }

  .filter-tabs {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 4px;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 14px;
    border: none;
    background: transparent;
    color: var(--text-label);
    font-size: 13px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: background 0.15s, color 0.15s;
  }

  .tab:hover {
    background: rgba(99, 102, 241, 0.1);
  }

  .tab.active {
    background: #6366F1;
    color: #fff;
    font-weight: 600;
  }

  .tab-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .tab-badge {
    font-size: 11px;
    padding: 1px 6px;
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.05);
  }

  .tab-badge.active-badge {
    background: rgba(255, 255, 255, 0.15);
  }

  .right-tools {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .search-box {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    border-radius: var(--radius-sm);
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    width: 240px;
  }

  .search-input {
    background: none;
    border: none;
    outline: none;
    color: var(--text-primary);
    font-size: 13px;
    width: 100%;
  }

  .search-input::placeholder {
    color: #3A3A5C;
  }

  .export-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 14px;
    border-radius: var(--radius-sm);
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    color: var(--text-label);
    font-size: 13px;
    transition: color 0.15s;
  }

  .export-btn:hover {
    color: var(--text-primary);
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
    color: #3A3A5C;
    letter-spacing: 1px;
  }

  .th-time { width: 110px; flex-shrink: 0; }
  .th-sev { width: 110px; flex-shrink: 0; }
  .th-col { width: 120px; flex-shrink: 0; }
  .th-msg { flex: 1; min-width: 0; }
  .th-val { width: 90px; flex-shrink: 0; }
  .th-thresh { width: 90px; flex-shrink: 0; }

  .table-body {
    flex: 1;
    overflow-y: auto;
  }

  .table-row {
    display: flex;
    align-items: center;
    padding: 12px 20px;
    border-bottom: 1px solid var(--border-row);
    transition: background 0.1s;
  }

  .table-row:hover {
    background: rgba(30, 27, 75, 0.15);
  }

  .table-row.highlighted {
    background: var(--bg-input);
  }

  .td {
    font-size: 12px;
    font-family: var(--font-mono);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .td-time { width: 110px; flex-shrink: 0; color: var(--text-label); }
  .td-sev {
    width: 110px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .td-col { width: 120px; flex-shrink: 0; color: #8888AA; }
  .td-msg { flex: 1; min-width: 0; color: var(--text-primary); }
  .td-val { width: 90px; flex-shrink: 0; font-weight: 600; }
  .td-thresh { width: 90px; flex-shrink: 0; color: var(--text-label); }

  .sev-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .sev-badge {
    font-size: 11px;
    font-weight: 600;
    font-family: var(--font-sans);
    padding: 2px 8px;
    border-radius: var(--radius-xs);
  }

  .empty-state {
    padding: 32px 20px;
    text-align: center;
    color: var(--text-secondary);
    font-size: 13px;
  }
</style>
