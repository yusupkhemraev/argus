<script>
  let { label = '', value = '', icon: Icon = null, color = '#fff', subtitle = '', barPercent = -1, warning = false } = $props();
</script>

<div class="card">
  <div class="card-top">
    <span class="card-label">{label}</span>
    {#if Icon}
      <Icon size={15} color={color} />
    {/if}
  </div>
  <span class="card-value" style="color: {warning ? color : 'var(--text-primary)'}">{value}</span>
  {#if barPercent >= 0}
    <div class="bar-bg">
      <div class="bar-fill" style="width: {Math.min(100, barPercent)}%; background: {color}"></div>
    </div>
  {/if}
  {#if subtitle}
    <div class="card-sub" class:warn={warning}>
      {#if warning}
        <Icon size={11} color={color} />
      {/if}
      <span style="color: {warning ? color : 'var(--text-label)'}">{subtitle}</span>
    </div>
  {/if}
</div>

<style>
  .card {
    flex: 1;
    min-width: 0;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    height: 168px;
  }

  .card-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .card-label {
    font-size: 12px;
    font-weight: 500;
    color: var(--text-label);
  }

  .card-value {
    font-size: 32px;
    font-weight: 700;
    letter-spacing: -1px;
    line-height: 1;
  }

  .bar-bg {
    height: 4px;
    background: var(--border);
    border-radius: 2px;
    overflow: hidden;
  }

  .bar-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.3s ease;
  }

  .card-sub {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 11px;
    color: var(--text-label);
    font-family: var(--font-mono);
  }

  .card-sub.warn {
    color: var(--accent-amber);
  }
</style>
