import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { mkdtempSync, rmSync } from 'fs'
import { join } from 'path'
import { tmpdir } from 'os'
import { SQLiteHitStore } from '../../../src/core/store/sqlite'
import type { HitEvent } from '../../../src/core/store/base'

function makeHit(overrides: Partial<HitEvent> = {}): HitEvent {
  return {
    ts: new Date(),
    path: '/api/v1/users',
    method: 'GET',
    caller_id: 'key-abc',
    user_agent: 'test/1.0',
    days_until_sunset: 100,
    endpoint_key: '/api/v1/users',
    enforced: false,
    ...overrides,
  }
}

let tmpDir: string
let store: SQLiteHitStore

beforeEach(() => {
  tmpDir = mkdtempSync(join(tmpdir(), 'dusk-test-'))
  store = new SQLiteHitStore(join(tmpDir, 'test.db'))
})

afterEach(async () => {
  await store.close()
  rmSync(tmpDir, { recursive: true })
})

describe('SQLiteHitStore', () => {
  it('records and retrieves a hit', async () => {
    await store.record(makeHit())
    const hits = await store.recentHits({ limit: 10 })
    expect(hits).toHaveLength(1)
    expect(hits[0].path).toBe('/api/v1/users')
  })

  it('persists enforced=true for 410 hits', async () => {
    await store.record(makeHit({ enforced: true }))
    const hits = await store.recentHits()
    expect(hits[0].enforced).toBe(true)
  })

  it('filters hits by endpoint_key', async () => {
    await store.record(makeHit({ path: '/api/v1/users', endpoint_key: '/api/v1/users' }))
    await store.record(makeHit({ path: '/api/v1/orders/1', endpoint_key: '/api/v1/orders/{id}' }))
    const hits = await store.recentHits({ endpoint_key: '/api/v1/users' })
    expect(hits).toHaveLength(1)
    expect(hits[0].endpoint_key).toBe('/api/v1/users')
  })

  it('returns endpoint summaries with correct counts', async () => {
    await store.record(makeHit({ caller_id: 'key-a' }))
    await store.record(makeHit({ caller_id: 'key-b' }))
    await store.record(makeHit({ path: '/api/v1/orders/1', endpoint_key: '/api/v1/orders/{id}', caller_id: 'key-a' }))

    const summaries = await store.endpointSummaries()
    const users = summaries.find((s) => s.endpoint_key === '/api/v1/users')!
    expect(users.total_hits).toBe(2)
    expect(users.unique_callers).toBe(2)
  })

  it('returns total summary with past_sunset_with_traffic', async () => {
    await store.record(makeHit({ caller_id: 'key-a', days_until_sunset: 10 }))
    await store.record(makeHit({
      path: '/api/v1/gone', endpoint_key: '/api/v1/gone',
      caller_id: 'key-b', days_until_sunset: -5, enforced: true,
    }))

    const s = await store.totalSummary()
    expect(s.total_hits).toBe(2)
    expect(s.endpoints_with_traffic).toBe(2)
    expect(s.past_sunset_with_traffic).toBe(1)
  })

  it('orders recent hits newest first', async () => {
    const now = Date.now()
    await store.record(makeHit({ ts: new Date(now - 10_000) }))  // 10s ago
    await store.record(makeHit({ ts: new Date(now - 5_000) }))   // 5s ago (newer)
    const hits = await store.recentHits({ limit: 10 })
    expect(hits).toHaveLength(2)
    expect(hits[0].ts >= hits[1].ts).toBe(true)
  })
})
