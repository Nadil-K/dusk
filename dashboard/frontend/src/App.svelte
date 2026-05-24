<script lang="ts">
  import { loading, error } from './stores/data.js'
  import Summary from './lib/Summary.svelte'
  import EndpointTable from './lib/EndpointTable.svelte'
  import HitTail from './lib/HitTail.svelte'
  import CallerList from './lib/CallerList.svelte'

  let selectedEndpoint: string | null = null
</script>

<main>
  <header>
    <h1>dusk</h1>
    <span class="subtitle">API deprecation manager</span>
  </header>

  {#if $error}
    <div class="error-banner">Failed to load data: {$error}</div>
  {/if}

  {#if $loading}
    <div class="loading">Loading…</div>
  {:else}
    <Summary />

    <section>
      <h2>Deprecated endpoints</h2>
      <EndpointTable />
    </section>

    <div class="lower">
      <section class="hits">
        <h2>Recent hits</h2>
        <HitTail />
      </section>
      <section class="callers">
        <h2>Top callers</h2>
        <CallerList endpointKey={selectedEndpoint} />
        {#if !selectedEndpoint}
          <p class="muted">Select an endpoint above to see callers.</p>
        {/if}
      </section>
    </div>
  {/if}
</main>

<style>
  :global(*, *::before, *::after) { box-sizing: border-box; }
  :global(body) { margin: 0; background: #11111b; color: #cdd6f4; font-family: 'Inter', system-ui, sans-serif; font-size: 15px; }
  main { max-width: 1200px; margin: 0 auto; padding: 2rem 1.5rem; }
  header { display: flex; align-items: baseline; gap: 0.75rem; margin-bottom: 2rem; border-bottom: 1px solid #313244; padding-bottom: 1rem; }
  h1 { margin: 0; font-size: 1.5rem; color: #89b4fa; letter-spacing: -0.02em; }
  .subtitle { color: #585b70; font-size: 0.85rem; }
  h2 { font-size: 0.9rem; text-transform: uppercase; letter-spacing: 0.05em; color: #a6adc8; margin: 0 0 0.75rem; }
  section { margin-bottom: 2rem; }
  .lower { display: grid; grid-template-columns: 1fr 320px; gap: 2rem; }
  .error-banner { background: #3d1e1e; color: #f38ba8; border: 1px solid #f38ba8; border-radius: 6px; padding: 0.75rem 1rem; margin-bottom: 1rem; font-size: 0.875rem; }
  .loading { color: #585b70; text-align: center; padding: 3rem; }
  .muted { color: #585b70; font-size: 0.85rem; }
</style>
