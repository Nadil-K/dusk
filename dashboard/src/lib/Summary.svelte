<script>
  import { summary } from '../stores/data.js'
</script>

{#if $summary}
  <div class="cards">
    <div class="card">
      <span class="label">Endpoints with traffic</span>
      <span class="value">{$summary.endpoints_with_traffic}</span>
    </div>
    <div class="card {$summary.past_sunset_with_traffic > 0 ? 'warn' : ''}">
      <span class="label">Past sunset</span>
      <span class="value">{$summary.past_sunset_with_traffic}</span>
    </div>
    <div class="card">
      <span class="label">Hits (30d)</span>
      <span class="value">{$summary.total_hits.toLocaleString()}</span>
    </div>
    <div class="card">
      <span class="label">Unique callers (30d)</span>
      <span class="value">{$summary.unique_callers.toLocaleString()}</span>
    </div>
  </div>
{/if}

<style>
  .cards {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1rem;
    margin-bottom: 2rem;
  }
  .card {
    background: #1e1e2e;
    border: 1px solid #313244;
    border-radius: 8px;
    padding: 1.25rem 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .card.warn {
    border-color: #f38ba8;
    background: #2d1b22;
  }
  .label {
    font-size: 0.8rem;
    color: #a6adc8;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .value {
    font-size: 2rem;
    font-weight: 700;
    color: #cdd6f4;
  }
</style>
