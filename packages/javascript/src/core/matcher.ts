import { EndpointConfig } from './models'

const PARAM_RE = /\{[^}]+\}/g

function pathToRegex(path: string): RegExp {
  const parts = path.split(PARAM_RE)
  const escaped = parts.map((p) => p.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
  return new RegExp(`^${escaped.join('[^/]+')}$`)
}

export class RouteMatcher {
  private readonly routes: Array<{ pattern: RegExp; endpoint: EndpointConfig }>

  constructor(endpoints: EndpointConfig[]) {
    this.routes = endpoints.map((ep) => ({
      pattern: pathToRegex(ep.path),
      endpoint: ep,
    }))
  }

  match(path: string, method: string): EndpointConfig | null {
    const upper = method.toUpperCase()
    for (const { pattern, endpoint } of this.routes) {
      if (pattern.test(path) && endpoint.methods.includes(upper)) {
        return endpoint
      }
    }
    return null
  }
}
