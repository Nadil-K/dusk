import { EndpointConfig } from './models.js'

export interface EnforcerResult {
  enforce: boolean
  body: Record<string, string> | null
}

export function checkSunset(endpoint: EndpointConfig): EnforcerResult {
  if (!endpoint.sunset_at) return { enforce: false, body: null }

  const isPast = new Date() > new Date(endpoint.sunset_at)
  if (!isPast) return { enforce: false, body: null }

  const body: Record<string, string> = {
    error: 'Gone',
    message: `${endpoint.path} was sunset on ${endpoint.sunset_at} and is no longer available.`,
  }
  if (endpoint.successor) body.successor = endpoint.successor
  if (endpoint.migration_doc) body.migration_doc = endpoint.migration_doc

  return { enforce: true, body }
}
