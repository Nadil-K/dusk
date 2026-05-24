import { EndpointConfig } from './models.js'

export interface DeprecationHeaders {
  Deprecation: string
  Sunset?: string
  Link?: string
}

export function buildHeaders(endpoint: EndpointConfig): DeprecationHeaders {
  const headers: DeprecationHeaders = {
    Deprecation: `@"${endpoint.deprecated_at}"`,
  }

  if (endpoint.sunset_at) {
    headers.Sunset = endpoint.sunset_at
  }

  const links: string[] = []
  if (endpoint.successor) {
    links.push(`<${endpoint.successor}>; rel="successor-version"`)
  }
  if (endpoint.migration_doc) {
    links.push(`<${endpoint.migration_doc}>; rel="deprecation"`)
  }
  if (links.length) {
    headers.Link = links.join(', ')
  }

  return headers
}

export function daysUntilSunset(endpoint: EndpointConfig): number | null {
  if (!endpoint.sunset_at) return null
  const sunset = new Date(endpoint.sunset_at).getTime()
  const now = Date.now()
  return Math.floor((sunset - now) / 86_400_000)
}
