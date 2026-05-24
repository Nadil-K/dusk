import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { mkdtempSync, rmSync, writeFileSync } from 'fs'
import { join } from 'path'
import { tmpdir } from 'os'
import express from 'express'
import request from 'supertest'
import { createDusk } from '../../src/adapters/express'

const DUSK_YAML = (dbPath: string) => `
dusk:
  version: 1
  store:
    backend: sqlite
    path: ${dbPath}
  log:
    identify_by:
      - header: X-API-Key
  endpoints:
    - path: /api/v1/users
      methods: [GET, POST]
      deprecated_at: "2025-01-01"
      sunset_at: "2099-01-01"
      successor: /api/v2/users
    - path: /api/v1/gone
      methods: [GET]
      deprecated_at: "2020-01-01"
      sunset_at: "2020-06-01"
    - path: /api/v1/orders/{id}
      methods: [GET]
      deprecated_at: "2025-03-01"
      sunset_at: "2099-06-01"
`

let tmpDir: string
let configPath: string
let app: express.Application

beforeEach(() => {
  tmpDir = mkdtempSync(join(tmpdir(), 'dusk-express-test-'))
  const dbPath = join(tmpDir, 'hits.db').replace(/\\/g, '/')
  configPath = join(tmpDir, 'dusk.yaml')
  writeFileSync(configPath, DUSK_YAML(dbPath))

  app = express()
  app.use(createDusk(configPath))

  app.get('/api/v1/users', (_req, res) => res.json({ users: [] }))
  app.get('/api/v1/gone', (_req, res) => res.json({}))
  app.get('/api/v1/orders/:id', (req, res) => res.json({ id: req.params.id }))
  app.get('/api/v2/users', (_req, res) => res.json({ users: [] }))
})

afterEach(() => {
  rmSync(tmpDir, { recursive: true })
})

describe('createDusk (Express middleware)', () => {
  it('adds Deprecation and Sunset headers to deprecated endpoints', async () => {
    const res = await request(app).get('/api/v1/users')
    expect(res.status).toBe(200)
    expect(res.headers['deprecation']).toBe('@1735689600')
    expect(res.headers['sunset']).toBe('Thu, 01 Jan 2099 00:00:00 GMT')
  })

  it('adds Link header with successor-version', async () => {
    const res = await request(app).get('/api/v1/users')
    expect(res.headers['link']).toContain('successor-version')
    expect(res.headers['link']).toContain('/api/v2/users')
  })

  it('returns 410 Gone for past-sunset endpoints', async () => {
    const res = await request(app).get('/api/v1/gone')
    expect(res.status).toBe(410)
    expect(res.body.error).toBe('Gone')
    expect(res.body.message).toContain('2020-06-01')
  })

  it('does not add headers to non-deprecated endpoints', async () => {
    const res = await request(app).get('/api/v2/users')
    expect(res.status).toBe(200)
    expect(res.headers['deprecation']).toBeUndefined()
  })

  it('matches path parameters', async () => {
    const res = await request(app).get('/api/v1/orders/123')
    expect(res.status).toBe(200)
    expect(res.headers['deprecation']).toBeTruthy()
  })

  it('passes through requests not matching any endpoint', async () => {
    const res = await request(app).get('/health')
    expect(res.status).toBe(404) // Express default 404 — no dusk interference
  })
})
