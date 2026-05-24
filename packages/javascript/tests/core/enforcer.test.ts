import { describe, it, expect } from 'vitest'
import { checkSunset } from '../../src/core/enforcer'
import type { EndpointConfig } from '../../src/core/models'

function ep(sunset_at: string | null, successor?: string): EndpointConfig {
  return { path: '/api/v1/users', methods: ['GET'], deprecated_at: '2025-01-01', sunset_at, successor }
}

describe('checkSunset', () => {
  it('does not enforce when no sunset_at', () => {
    const r = checkSunset(ep(null))
    expect(r.enforce).toBe(false)
    expect(r.body).toBeNull()
  })

  it('does not enforce for a future sunset', () => {
    expect(checkSunset(ep('2099-12-31')).enforce).toBe(false)
  })

  it('enforces for a past sunset', () => {
    const r = checkSunset(ep('2000-01-01'))
    expect(r.enforce).toBe(true)
    expect(r.body?.error).toBe('Gone')
  })

  it('includes successor in 410 body', () => {
    const r = checkSunset(ep('2000-01-01', '/api/v2/users'))
    expect(r.body?.successor).toBe('/api/v2/users')
  })

  it('includes the sunset date in the message', () => {
    const r = checkSunset(ep('2000-01-01'))
    expect(r.body?.message).toContain('2000-01-01')
  })
})
