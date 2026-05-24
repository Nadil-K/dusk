<script lang="ts">
  import { loading, error, selectedEndpoint } from './stores/data.js'
  import Summary from './lib/Summary.svelte'
  import EndpointTable from './lib/EndpointTable.svelte'
  import HitTail from './lib/HitTail.svelte'
  import CallerList from './lib/CallerList.svelte'
</script>

<main>
  <header>
    <div class="brand">
      <h1>dusk</h1>
      <span class="subtitle">Deprecated API usage monitoring</span>
    </div>
    <a class="github" href="https://github.com/Nadil-K/dusk" target="_blank" rel="noopener">
      <svg viewBox="0 0 16 16" width="22" height="22" fill="currentColor" aria-hidden="true">
        <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38
          0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13
          -.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66
          .07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15
          -.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27s1.36.09
          2 .27c1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82
          2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01
          2.2 0 .21.15.46.55.38A8.01 8.01 0 0 0 16 8c0-4.42-3.58-8-8-8z"/>
      </svg>
      GitHub
    </a>
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
        <CallerList endpointKey={$selectedEndpoint} />
      </section>
    </div>
  {/if}
</main>

<style>
  :global(*, *::before, *::after) { box-sizing: border-box; }
  :global(body) { margin: 0; background: #11111b; color: #cdd6f4; font-family: 'Inter', system-ui, sans-serif; font-size: 15px; }
  main { max-width: 1200px; margin: 0 auto; padding: 2rem 1.5rem; }
  header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 2rem; border-bottom: 1px solid #313244; padding-bottom: 1rem; }
  .brand { display: flex; align-items: baseline; gap: 0.75rem; }
  h1 { margin: 0; font-size: 1.5rem; color: #89b4fa; letter-spacing: -0.02em; }
  .subtitle { color: #585b70; font-size: 0.85rem; }
  .github { display: flex; align-items: center; gap: 0.5rem; color: #a6adc8; font-size: 0.9rem; text-decoration: none; transition: color 0.15s; }
  .github:hover { color: #cdd6f4; }
  h2 { font-size: 0.9rem; text-transform: uppercase; letter-spacing: 0.05em; color: #a6adc8; margin: 0 0 0.75rem; }
  section { margin-bottom: 2rem; }
  .lower { display: grid; grid-template-columns: 1fr 320px; gap: 2rem; }
  .error-banner { background: #3d1e1e; color: #f38ba8; border: 1px solid #f38ba8; border-radius: 6px; padding: 0.75rem 1rem; margin-bottom: 1rem; font-size: 0.875rem; }
  .loading { color: #585b70; text-align: center; padding: 3rem; }
</style>
