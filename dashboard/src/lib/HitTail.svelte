<script>
  import { recentHits } from '../stores/data.js'
</script>

<div class="tail">
  {#each $recentHits as hit (hit.ts + hit.path)}
    <div class="hit-row" class:past-sunset={hit.days_left !== null && hit.days_left <= 0}>
      <span class="ts">{new Date(hit.ts).toLocaleTimeString()}</span>
      <span class="method method-{hit.method.toLowerCase()}">{hit.method}</span>
      <span class="path">{hit.path}</span>
      <span class="caller">{hit.caller_id ?? 'anonymous'}</span>
      {#if hit.days_left !== null}
        <span class="days" class:urgent={hit.days_left < 0}>
          {hit.days_left < 0
            ? `${Math.abs(hit.days_left)}d past sunset`
            : `${hit.days_left}d left`}
        </span>
      {/if}
    </div>
  {:else}
    <div class="empty">No hits recorded yet.</div>
  {/each}
</div>

<style>
  .tail {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.85rem;
    font-family: 'Cascadia Code', 'Fira Code', monospace;
  }
  .hit-row {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.4rem 0.75rem;
    border-radius: 4px;
    background: #1e1e2e;
  }
  .hit-row.past-sunset { background: #2d1b22; }
  .ts    { color: #585b70; min-width: 80px; }
  .method { font-weight: 700; min-width: 48px; }
  .method-get    { color: #89b4fa; }
  .method-post   { color: #a6e3a1; }
  .method-put    { color: #fab387; }
  .method-patch  { color: #f9e2af; }
  .method-delete { color: #f38ba8; }
  .path   { color: #cdd6f4; flex: 1; }
  .caller { color: #a6adc8; min-width: 140px; overflow: hidden; text-overflow: ellipsis; }
  .days   { color: #a6e3a1; font-size: 0.75rem; }
  .days.urgent { color: #f38ba8; font-weight: 700; }
  .empty  { color: #585b70; padding: 1rem; text-align: center; }
</style>
