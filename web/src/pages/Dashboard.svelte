<script>
  import { TriangleAlert, Cpu, MemoryStick, HardDrive, Globe, ListFilter, Download } from 'lucide-svelte';
  import Topbar from '../components/Topbar.svelte';
  import StatCard from '../components/StatCard.svelte';
  import { metrics, alarms, alarmCounts, filteredAlarms, filter } from '../lib/stores.js';

  let memoryValue = $derived($metrics.memory?.value ?? 0);
  let cpuValue = $derived($metrics.cpu?.value ?? 0);
  let diskValue = $derived($metrics.disk?.value ?? 0);
  let nginxErrors = $derived(parseInt($metrics.nginx?.labels?.http_errors ?? '0'));
  let nginxSlow = $derived(parseInt($metrics.nginx?.labels?.slow_requests ?? '0'));

  let memLabels = $derived($metrics.memory?.labels ?? {});
  let diskLabels = $derived($metrics.disk?.labels ?? {});

  let memSubtitle = $derived(() => {
    const used = memLabels.used_gb;
    const total = memLabels.total_gb;
    const free = memLabels.free_gb;
    if (used && total) return `${used} / ${total} GB  ·  ${free ?? '?'} GB free`;
    return '';
  });

  let diskSubtitle = $derived(() => {
    const path = diskLabels.path ?? '/';
    const used = diskLabels.used_gb;
    const total = diskLabels.total_gb;
    const free = diskLabels.free_gb;
    if (used && total) return `${path} · ${used} / ${total} GB · ${free ?? '?'} GB free`;
    return '';
  });

  let activeAlarmCount = $derived($alarmCounts.critical + $alarmCounts.warning);
  let alarmSubtitle = $derived(() => {
    const c = $alarmCounts;
    const parts = [];
    if (c.critical > 0) parts.push(`${c.critical} critical`);
    if (c.warning > 0) parts.push(`${c.warning} warning`);
    return parts.join(' · ') || 'no active alarms';
  });

  let recentAlarms = $derived($filteredAlarms.slice(0, 20));

  let filterText = $state('');
  let filtering = $state(false);

  function startFilter() { filtering = true; }
  function clearFilter() { filtering = false; filterText = ''; filter.set(''); }
  function handleFilterKey(e) { if (e.key === 'Escape') clearFilter(); }

  $effect(() => { filter.set(filterText); });

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

<Topbar title="Dashboard" subtitle="overview" />

<div class="content">
  <div class="stat-cards">
    <StatCard
      label="Active Alarms"
      value={activeAlarmCount.toString()}
      icon={TriangleAlert}
      color="var(--accent-red)"
      subtitle={alarmSubtitle()}
      warning={activeAlarmCount > 0}
    />
    <StatCard
      label="CPU Usage"
      value="{cpuValue.toFixed(1)}%"
      icon={Cpu}
      color="var(--bar-cpu)"
      barPercent={cpuValue}
    />
    <StatCard
      label="Memory"
      value="{memoryValue.toFixed(1)}%"
      icon={MemoryStick}
      color="var(--bar-memory)"
      barPercent={memoryValue}
      subtitle={memSubtitle()}
    />
    <StatCard
      label="Disk Usage"
      value="{diskValue.toFixed(1)}%"
      icon={HardDrive}
      color="var(--bar-disk)"
      barPercent={diskValue}
      subtitle={diskSubtitle()}
      warning={diskValue >= 75}
    />
    <StatCard
      label="Nginx"
      value="{nginxErrors} errs"
      icon={Globe}
      color="var(--accent-amber)"
      subtitle="{nginxSlow} slow requests"
      warning={nginxErrors > 0}
    />
  </div>

  <div class="table-section">
    <div class="table-topbar">
      <div class="table-title-group">
        <span class="table-title">Recent Alarms</span>
        <span class="count-badge">{recentAlarms.length}</span>
      </div>
      <div class="table-actions">
        {#if filtering}
          <div class="filter-input-wrap">
            <ListFilter size={13} color="var(--text-label)" />
            <input
              class="filter-input"
              bind:value={filterText}
              onkeydown={handleFilterKey}
              placeholder="type to filter..."
              autofocus
            />
          </div>
        {:else}
          <button class="action-btn" onclick={startFilter}>
            <ListFilter size={13} />
            <span>Filter</span>
          </button>
        {/if}
        <button class="action-btn" onclick={exportAlarms}>
          <Download size={13} />
          <span>Export</span>
        </button>
      </div>
    </div>

    <div class="table-card">
      <div class="table-head">
        <span class="th th-time">TIME</span>
        <span class="th th-sev">SEVERITY</span>
        <span class="th th-msg">MESSAGE</span>
        <span class="th th-src">COLLECTOR</span>
        <span class="th th-val">VALUE</span>
        <span class="th th-thresh">THRESHOLD</span>
      </div>
      <div class="table-body">
        {#each recentAlarms as alarm}
          <div class="table-row">
            <span class="td td-time">{formatTime(alarm.triggered_at)}</span>
            <span class="td td-sev">
              <span class="sev-dot" style="background: {severityColor(alarm.severity)}"></span>
              <span class="sev-badge" style="background: {severityBg(alarm.severity)}; color: {severityColor(alarm.severity)}">{alarm.severity}</span>
            </span>
            <span class="td td-msg">{alarm.message}</span>
            <span class="td td-src">{alarm.collector ?? '-'}</span>
            <span class="td td-val" style="color: {severityColor(alarm.severity)}">{alarm.value != null ? alarm.value.toFixed(1) : '-'}</span>
            <span class="td td-thresh">{alarm.threshold != null ? alarm.threshold.toFixed(1) : '-'}</span>
          </div>
        {:else}
          <div class="empty-state">no alarms</div>
        {/each}
      </div>
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

  .stat-cards {
    display: flex;
    gap: 16px;
    flex-shrink: 0;
  }

  .table-section {
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex: 1;
    min-height: 0;
  }

  .table-topbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-shrink: 0;
  }

  .table-title-group {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .table-title {
    font-size: 15px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .count-badge {
    font-size: 11px;
    font-weight: 600;
    color: var(--accent-red);
    background: rgba(239, 68, 68, 0.1);
    padding: 2px 8px;
    border-radius: var(--radius-xs);
  }

  .table-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 7px 12px;
    border-radius: var(--radius-sm);
    background: var(--bg-input);
    border: 1px solid var(--border);
    color: var(--text-label);
    font-size: 12px;
    transition: color 0.15s;
  }

  .action-btn:hover {
    color: var(--text-primary);
  }

  .filter-input-wrap {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 7px 12px;
    border-radius: var(--radius-sm);
    background: var(--bg-input);
    border: 1px solid var(--accent-purple);
  }

  .filter-input {
    background: none;
    border: none;
    outline: none;
    color: var(--text-primary);
    font-size: 12px;
    width: 150px;
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
    padding: 12px 20px;
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
  .th-msg { flex: 1; min-width: 0; }
  .th-src { width: 130px; flex-shrink: 0; }
  .th-val { width: 90px; flex-shrink: 0; }
  .th-thresh { width: 90px; flex-shrink: 0; }

  .table-body {
    flex: 1;
    overflow-y: auto;
  }

  .table-row {
    display: flex;
    align-items: center;
    padding: 14px 20px;
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

  .td-time { width: 110px; flex-shrink: 0; color: var(--text-label); }
  .td-sev {
    width: 110px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .td-msg { flex: 1; min-width: 0; color: var(--text-primary); }
  .td-src { width: 130px; flex-shrink: 0; color: #8888AA; }
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
