<script>
  let { title = '', subtitle = '' } = $props();
</script>

<div class="topbar">
  <div class="breadcrumb">
    <span class="page-title">{title}</span>
    {#if subtitle}
      <span class="slash">/</span>
      <span class="page-subtitle">{subtitle}</span>
    {/if}
  </div>
</div>

<style>
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 24px;
    background: var(--bg-topbar);
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }

  .breadcrumb {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .page-title {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .slash {
    font-size: 14px;
    color: var(--text-tertiary);
  }

  .page-subtitle {
    font-size: 14px;
    color: var(--text-label);
  }
</style>
