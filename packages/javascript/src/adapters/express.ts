import type { Request, Response, NextFunction } from 'express'
import { loadConfig } from '../core/config.js'
import { RouteMatcher } from '../core/matcher.js'
import { buildHeaders } from '../core/headers.js'
import { checkSunset } from '../core/enforcer.js'
import type { DuskConfig } from '../core/models.js'

export type DuskMiddleware = (req: Request, res: Response, next: NextFunction) => void

export function createDusk(configPath = 'dusk.yaml'): DuskMiddleware {
  const cfg: DuskConfig = loadConfig(configPath)
  const matcher = new RouteMatcher(cfg.endpoints)

  return function duskMiddleware(req, res, next) {
    const endpoint = matcher.match(req.path, req.method)
    if (!endpoint) return next()

    const headers = buildHeaders(endpoint)
    Object.entries(headers).forEach(([k, v]) => res.setHeader(k, v))

    const { enforce, body } = checkSunset(endpoint)
    if (enforce) {
      return res.status(410).json(body)
    }

    next()
  }
}
