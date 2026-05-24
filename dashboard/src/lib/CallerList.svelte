<script>
  import { fetchCallers } from './api.js'

  export let endpointKey = null

  let callers = []
  let loading = false

  $: if (endpointKey) {
    loading = true
    fetchCallers(endpointKey)
      .then(data => { callers = data.top_callers; loading = false })
      .catch(() => { loading = false })
  }
</script>

{#if endpointKey}
  <div class="caller-list">
    <h3>Top callers — <code>{endpointKey}</code></h3>
    {#if loading}
      <p class="muted">Loading…</p>
    {:else if callers.length === 0}
      <p class="muted">No callers recorded.</p>
    {:else}
      <ol>
        {#each callers as [id, count]}
          <li>
            <span class="id">{id}</span>
            <span class="count">{count.toLocaleString()} hits</span>
          </li>
        {/each}
      </ol>
    {/if}
  </div>
{/if}

<style>
  .caller-list { padding: 1rem 0; }
  h3 { color: #cdd6f4; font-size: 0.95rem; margin-bottom: 0.75rem; }
  code { color: #89b4fa; }
  ol { list-style: decimal inside; padding: 0; margin: 0; }
  li {
    display: flex;
    justify-content: space-between;
    padding: 0.4rem 0.75rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }
  li:nth-child(odd) { background: #1e1e2e; }
  .id    { color: #cdd6f4; font-family: monospace; }
  .count { color: #a6adc8; }
  .muted { color: #585b70; font-size: 0.85rem; }
</style>
