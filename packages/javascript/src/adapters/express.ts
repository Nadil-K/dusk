import type { Request, Response, NextFunction } from 'express'
import { loadConfig } from '../core/config'
import { RouteMatcher } from '../core/matcher'
import { buildHeaders, daysUntilSunset } from '../core/headers'
import { checkSunset } from '../core/enforcer'
import { createStore } from '../core/store'
import type { AsyncWriteBuffer } from '../core/store/buffer'
import type { HitEvent } from '../core/store/base'
import type { DuskConfig } from '../core/models'

export type DuskMiddleware = (req: Request, res: Response, next: NextFunction) => void

function resolveCallerId(req: Request, cfg: DuskConfig): string | null {
  for (const rule of cfg.log.identify_by) {
    if (rule === 'ip') {
      return req.ip ?? req.socket.remoteAddress ?? null
    }
    const val = req.headers[rule.header.toLowerCase()]
    if (val) return Array.isArray(val) ? val[0] : val
  }
  return null
}

export function createDusk(configPath = 'dusk.yaml'): DuskMiddleware {
  const cfg: DuskConfig = loadConfig(configPath)
  const matcher = new RouteMatcher(cfg.endpoints)
  const store: AsyncWriteBuffer = createStore(cfg)

  return function duskMiddleware(req, res, next) {
    const endpoint = matcher.match(req.path, req.method)
    if (!endpoint) return next()

    const headers = buildHeaders(endpoint)
    Object.entries(headers).forEach(([k, v]) => res.setHeader(k, v))

    const { enforce, body } = checkSunset(endpoint)
    const days_left = daysUntilSunset(endpoint)
    const caller_id = resolveCallerId(req, cfg)

    const hit: HitEvent = {
      ts: new Date(),
      path: req.path,
      method: req.method,
      caller_id,
      user_agent: req.headers['user-agent'] ?? null,
      days_until_sunset: days_left,
      endpoint_key: endpoint.path,
      enforced: enforce,
    }

    // Log regardless of enforcement — sunset traffic is what operators need to see.
    void store.record(hit)

    if (enforce) {
      return res.status(410).json(body)
    }

    next()
  }
}
