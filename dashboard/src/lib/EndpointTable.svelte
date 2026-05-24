<script>
  import { endpoints } from '../stores/data.js'

  function statusClass(ep) {
    if (ep.past_sunset) return 'badge-danger'
    if (ep.days_left !== null && ep.days_left < 30) return 'badge-warning'
    if (!ep.sunset_at) return 'badge-soft'
    return 'badge-active'
  }

  function statusLabel(ep) {
    if (ep.past_sunset) return 'Past sunset'
    if (!ep.sunset_at) return 'Soft'
    return `${ep.days_left}d left`
  }
</script>

<table>
  <thead>
    <tr>
      <th>Endpoint</th>
      <th>Methods</th>
      <th>Status</th>
      <th>Hits (30d)</th>
      <th>Callers</th>
      <th>Last seen</th>
    </tr>
  </thead>
  <tbody>
    {#each $endpoints as ep (ep.path)}
      <tr class={ep.past_sunset && ep.total_hits > 0 ? 'row-urgent' : ''}>
        <td><code>{ep.path}</code></td>
        <td class="methods">{ep.methods.join(', ')}</td>
        <td><span class="badge {statusClass(ep)}">{statusLabel(ep)}</span></td>
        <td>{ep.total_hits.toLocaleString()}</td>
        <td>{ep.unique_callers}</td>
        <td>{ep.last_seen ? new Date(ep.last_seen).toLocaleString() : '—'}</td>
      </tr>
    {/each}
  </tbody>
</table>

<style>
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.9rem;
  }
  th {
    text-align: left;
    padding: 0.6rem 1rem;
    border-bottom: 1px solid #313244;
    color: #a6adc8;
    font-weight: 600;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  td {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid #1e1e2e;
    color: #cdd6f4;
  }
  tr:hover td { background: #181825; }
  tr.row-urgent td { background: #2d1b22; }
  code {
    font-family: 'Cascadia Code', 'Fira Code', monospace;
    color: #89b4fa;
    font-size: 0.85rem;
  }
  .methods { color: #a6adc8; font-size: 0.8rem; }
  .badge {
    display: inline-block;
    padding: 0.2rem 0.6rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
  }
  .badge-active  { background: #1e3a2f; color: #a6e3a1; }
  .badge-warning { background: #3d2f1e; color: #fab387; }
  .badge-danger  { background: #3d1e1e; color: #f38ba8; }
  .badge-soft    { background: #1e2a3d; color: #89b4fa; }
</style>
