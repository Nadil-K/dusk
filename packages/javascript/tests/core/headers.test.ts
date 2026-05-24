import { describe, it, expect } from 'vitest'
import { buildHeaders, daysUntilSunset } from '../../src/core/headers'
import type { EndpointConfig } from '../../src/core/models'

function ep(overrides: Partial<EndpointConfig> = {}): EndpointConfig {
  return {
    path: '/api/v1/users',
    methods: ['GET'],
    deprecated_at: '2025-01-01',
    sunset_at: null,
    ...overrides,
  }
}

describe('buildHeaders', () => {
  it('always includes Deprecation header', () => {
    const h = buildHeaders(ep())
    expect(h.Deprecation).toContain('2025-01-01')
  })

  it('includes Sunset header when sunset_at is set', () => {
    expect(buildHeaders(ep({ sunset_at: '2026-01-01' })).Sunset).toBe('2026-01-01')
  })

  it('omits Sunset header when sunset_at is null', () => {
    expect(buildHeaders(ep()).Sunset).toBeUndefined()
  })

  it('includes Link header with successor-version rel', () => {
    const h = buildHeaders(ep({ successor: '/api/v2/users' }))
    expect(h.Link).toContain('successor-version')
    expect(h.Link).toContain('/api/v2/users')
  })

  it('includes Link header with deprecation rel for migration_doc', () => {
    const h = buildHeaders(ep({ migration_doc: 'https://docs.example.com' }))
    expect(h.Link).toContain('rel="deprecation"')
  })

  it('combines multiple Link entries', () => {
    const h = buildHeaders(ep({ successor: '/api/v2/users', migration_doc: 'https://docs.example.com' }))
    expect(h.Link?.split(', ').length).toBe(2)
  })
})

describe('daysUntilSunset', () => {
  it('returns null when no sunset_at', () => {
    expect(daysUntilSunset(ep())).toBeNull()
  })

  it('returns positive number for future sunset', () => {
    expect(daysUntilSunset(ep({ sunset_at: '2099-12-31' }))).toBeGreaterThan(0)
  })

  it('returns negative number for past sunset', () => {
    expect(daysUntilSunset(ep({ sunset_at: '2000-01-01' }))).toBeLessThan(0)
  })
})
