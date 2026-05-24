import { writable } from 'svelte/store'
import { fetchSummary, fetchEndpoints, fetchHits } from '../lib/api.js'

export const summary    = writable(null)
export const endpoints  = writable([])
export const recentHits = writable([])
export const loading    = writable(true)
export const error      = writable(null)

async function poll() {
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
    error.set(err.message)
    loading.set(false)
  }
}

poll()
setInterval(poll, 10_000)
