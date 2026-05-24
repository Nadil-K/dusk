import { EndpointConfig } from './models'

export interface DeprecationHeaders {
  Deprecation: string
  Sunset?: string
  Link?: string
}

function toUnix(isoDate: string): number {
  return Math.floor(new Date(isoDate + 'T00:00:00Z').getTime() / 1000)
}

function toHttpDate(isoDate: string): string {
  return new Date(isoDate + 'T00:00:00Z').toUTCString()
}

export function buildHeaders(endpoint: EndpointConfig): DeprecationHeaders {
  const headers: DeprecationHeaders = {
    Deprecation: `@${toUnix(endpoint.deprecated_at)}`,
  }

  if (endpoint.sunset_at) {
    headers.Sunset = toHttpDate(endpoint.sunset_at)
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
  return Math.floor((sunset - Date.now()) / 86_400_000)
}
