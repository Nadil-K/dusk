import { describe, it, expect } from 'vitest'
import { RouteMatcher } from '../../src/core/matcher'
import type { EndpointConfig } from '../../src/core/models'

function ep(path: string, methods = ['GET']): EndpointConfig {
  return { path, methods, deprecated_at: '2025-01-01', sunset_at: '2026-01-01' }
}

describe('RouteMatcher', () => {
  it('matches exact path', () => {
    expect(new RouteMatcher([ep('/api/v1/users')]).match('/api/v1/users', 'GET')).not.toBeNull()
  })

  it('returns null for wrong path', () => {
    expect(new RouteMatcher([ep('/api/v1/users')]).match('/api/v2/users', 'GET')).toBeNull()
  })

  it('returns null for wrong method', () => {
    expect(new RouteMatcher([ep('/api/v1/users', ['POST'])]).match('/api/v1/users', 'GET')).toBeNull()
  })

  it('matches path with param', () => {
    const m = new RouteMatcher([ep('/api/v1/orders/{id}')])
    expect(m.match('/api/v1/orders/123', 'GET')).not.toBeNull()
    expect(m.match('/api/v1/orders/abc-xyz', 'GET')).not.toBeNull()
  })

  it('does not match extra path segments', () => {
    expect(new RouteMatcher([ep('/api/v1/orders/{id}')]).match('/api/v1/orders/123/items', 'GET')).toBeNull()
  })

  it('is case-insensitive on method', () => {
    expect(new RouteMatcher([ep('/api/v1/users')]).match('/api/v1/users', 'get')).not.toBeNull()
  })

  it('returns the matched endpoint config', () => {
    const config = ep('/api/v1/users')
    const result = new RouteMatcher([config]).match('/api/v1/users', 'GET')
    expect(result?.path).toBe('/api/v1/users')
  })
})
