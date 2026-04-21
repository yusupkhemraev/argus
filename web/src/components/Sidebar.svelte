<script>
  import { LayoutDashboard, Bell, ScrollText, Settings, Monitor, Timer } from 'lucide-svelte';
  import { currentPage, navigate } from '../lib/router.js';
  import { connected, startTime, serverName } from '../lib/stores.js';

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard, path: '/' },
    { id: 'alarms', label: 'Alarms', icon: Bell, path: '/alarms' },
    { id: 'logs', label: 'Logs', icon: ScrollText, path: '/logs' },
    { id: 'settings', label: 'Settings', icon: Settings, path: '/settings' },
  ];

  let now = $state(Date.now());

  $effect(() => {
    const interval = setInterval(() => { now = Date.now(); }, 1000);
    return () => clearInterval(interval);
  });

  let uptime = $derived(() => {
    const seconds = Math.floor((now - $startTime) / 1000);
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  });
</script>

<aside class="sidebar">
  <div class="logo-area">
    <span class="logo-prompt">{'>'}</span>
    <span class="logo-name">argus</span>
  </div>

  <nav class="nav-section">
    {#each navItems as item}
      <button
        class="nav-item"
        class:active={$currentPage === item.id}
        onclick={() => navigate(item.path)}
        disabled={item.disabled}
      >
        <item.icon size={16} />
        <span>{item.label}</span>
      </button>
    {/each}
  </nav>

  <div class="spacer"></div>

  <div class="divider"></div>

  <div class="bottom-section">
    <div class="bottom-row">
      <Monitor size={13} color="var(--text-tertiary)" />
      <span class="bottom-text">{$serverName || 'unknown'}</span>
    </div>
    <div class="bottom-row">
      <span class="status-dot" class:connected={$connected} class:disconnected={!$connected}></span>
      <span class="status-text" class:connected={$connected}>{$connected ? 'Connected' : 'Disconnected'}</span>
    </div>
    <div class="bottom-row">
      <Timer size={13} color="var(--text-tertiary)" />
      <span class="bottom-text">Uptime: {uptime()}</span>
    </div>
  </div>
</aside>

<style>
  .sidebar {
    width: 240px;
    background: var(--bg-sidebar);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    height: 100%;
    flex-shrink: 0;
    overflow-y: auto;
  }

  .logo-area {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 24px 20px 20px 20px;
  }

  .logo-prompt {
    color: var(--accent-green-bright);
    font-family: var(--font-logo);
    font-size: 15px;
    font-weight: 700;
  }

  .logo-name {
    color: var(--text-primary);
    font-family: var(--font-logo);
    font-size: 15px;
    font-weight: 700;
  }

  .nav-section {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 0 8px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border: none;
    background: transparent;
    color: var(--text-nav);
    font-size: 13px;
    font-weight: 400;
    border-radius: var(--radius-sm);
    width: 100%;
    text-align: left;
    transition: background 0.15s, color 0.15s;
  }

  .nav-item:not(:disabled):hover {
    background: rgba(30, 27, 75, 0.5);
    color: var(--text-nav-active);
  }

  .nav-item.active {
    background: var(--bg-active);
    color: var(--text-nav-active);
    font-weight: 500;
  }

  .nav-item:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .spacer {
    flex: 1;
  }

  .divider {
    height: 1px;
    background: var(--border);
    margin: 0;
  }

  .bottom-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px 20px;
  }

  .bottom-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .bottom-text {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .status-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
  }

  .status-dot.connected {
    background: var(--accent-green-bright);
  }

  .status-dot.disconnected {
    background: var(--accent-red);
  }

  .status-text {
    font-size: 12px;
    color: var(--accent-red);
  }

  .status-text.connected {
    color: var(--accent-green-bright);
  }
</style>
