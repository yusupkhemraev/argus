<script>
  import { onMount } from 'svelte';
  import { connect, disconnect } from './lib/sse.js';
  import { currentPage } from './lib/router.js';
  import Sidebar from './components/Sidebar.svelte';
  import Dashboard from './pages/Dashboard.svelte';
  import Alarms from './pages/Alarms.svelte';
  import Logs from './pages/Logs.svelte';
  import Settings from './pages/Settings.svelte';

  onMount(() => {
    connect();
    return () => disconnect();
  });
</script>

<div class="app-shell">
  <Sidebar />
  <div class="main-slot">
    {#if $currentPage === 'alarms'}
      <Alarms />
    {:else if $currentPage === 'logs'}
      <Logs />
    {:else if $currentPage === 'settings'}
      <Settings />
    {:else}
      <Dashboard />
    {/if}
  </div>
</div>

<style>
  .app-shell {
    display: flex;
    height: 100%;
    overflow: hidden;
  }

  .main-slot {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
    overflow: hidden;
  }
</style>
