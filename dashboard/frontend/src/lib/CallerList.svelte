<script lang="ts">
  import { fetchAllCallers, fetchCallers } from './api.js'
  import { selectedEndpoint } from '../stores/data.js'

  export let endpointKey: string | null = null

  let callers: [string, number][] = []
  let loading = false

  async function load(key: string | null) {
    loading = true
    try {
      const data = key ? await fetchCallers(key) : await fetchAllCallers()
      callers = data.top_callers
    } catch {
      callers = []
    } finally {
      loading = false
    }
  }

  $: load(endpointKey)
</script>

<div class="caller-list">
  {#if endpointKey}
    <p class="label">Filtered by <code>{endpointKey}</code> <button class="clear" on:click={() => selectedEndpoint.set(null)}>Clear</button></p>
  {:else}
    <p class="label">All endpoints</p>
  {/if}

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

<style>
  .caller-list { padding: 0.5rem 0; }
  .label { display: flex; align-items: center; gap: 0.4rem; font-size: 0.8rem; color: #a6adc8; margin: 0 0 0.75rem; }
  code { color: #89b4fa; }
  .clear { background: #313244; border: none; color: #a6adc8; font-size: 0.7rem; cursor: pointer; padding: 0.15rem 0.45rem; border-radius: 4px; line-height: 1; transition: background 0.15s, color 0.15s; }
  .clear:hover { background: #45475a; color: #cdd6f4; }
  ol { list-style: decimal inside; padding: 0; margin: 0; }
  li { display: flex; justify-content: space-between; padding: 0.4rem 0.75rem; border-radius: 4px; font-size: 0.85rem; }
  li:nth-child(odd) { background: #1e1e2e; }
  .id    { color: #cdd6f4; font-family: monospace; }
  .count { color: #a6adc8; }
  .muted { color: #585b70; font-size: 0.85rem; }
</style>
