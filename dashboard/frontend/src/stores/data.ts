import { writable } from 'svelte/store'
import { fetchSummary, fetchEndpoints, fetchHits } from '../lib/api.js'
import type { Summary, EndpointRow, HitRow } from '../lib/api.js'

export const summary          = writable<Summary | null>(null)
export const endpoints        = writable<EndpointRow[]>([])
export const recentHits       = writable<HitRow[]>([])
export const loading          = writable<boolean>(true)
export const error            = writable<string | null>(null)
export const selectedEndpoint = writable<string | null>(null)

async function poll(): Promise<void> {
  try {
    const [s, e, h] = await Promise.all([
      fetchSummary(),
      fetchEndpoints(),
      fetchHits({ limit: 20 }),
    ])
    summary.set(s)
    endpoints.set(e)
    recentHits.set(h)
    loading.set(false)
    error.set(null)
  } catch (err) {
    error.set(err instanceof Error ? err.message : 'Unknown error')
    loading.set(false)
  }
}

poll()
setInterval(poll, 10_000)
