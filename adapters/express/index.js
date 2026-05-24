const fs = require('fs')
const path = require('path')
const yaml = require('js-yaml')

function loadConfig(configPath = 'dusk.yaml') {
  const raw = yaml.load(fs.readFileSync(configPath, 'utf8'))
  return raw.dusk
}

function pathToRegex(routePath) {
  const escaped = routePath.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, (c) =>
    c === '{' || c === '}' ? '' : '\\' + c
  )
  const pattern = routePath.replace(/\{[^}]+\}/g, '[^/]+')
  const escaped2 = pattern.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, (c) =>
    c === '[' || c === ']' || c === '^' || c === '+' || c === '/' ? c : '\\' + c
  )
  return new RegExp(`^${escaped2}$`)
}

function buildHeaders(endpoint) {
  const headers = {}
  headers['Deprecation'] = `@"${endpoint.deprecated_at}"`
  if (endpoint.sunset_at) headers['Sunset'] = endpoint.sunset_at
  const links = []
  if (endpoint.successor) links.push(`<${endpoint.successor}>; rel="successor-version"`)
  if (endpoint.migration_doc) links.push(`<${endpoint.migration_doc}>; rel="deprecation"`)
  if (links.length) headers['Link'] = links.join(', ')
  return headers
}

function daysUntilSunset(endpoint) {
  if (!endpoint.sunset_at) return null
  const sunset = new Date(endpoint.sunset_at)
  const now = new Date()
  return Math.floor((sunset - now) / 86_400_000)
}

function isPastSunset(endpoint) {
  if (!endpoint.sunset_at) return false
  return new Date() > new Date(endpoint.sunset_at)
}

function createDusk(configPath = 'dusk.yaml') {
  const cfg = loadConfig(configPath)

  const routes = (cfg.endpoints || []).map((ep) => ({
    pattern: pathToRegex(ep.path),
    methods: (ep.methods || ['GET']).map((m) => m.toUpperCase()),
    endpoint: ep,
  }))

  function match(reqPath, method) {
    for (const route of routes) {
      if (route.pattern.test(reqPath) && route.methods.includes(method.toUpperCase())) {
        return route.endpoint
      }
    }
    return null
  }

  return function duskMiddleware(req, res, next) {
    const endpoint = match(req.path, req.method)
    if (!endpoint) return next()

    const headers = buildHeaders(endpoint)
    Object.entries(headers).forEach(([k, v]) => res.set(k, v))

    if (isPastSunset(endpoint)) {
      const body = {
        error: 'Gone',
        message: `${endpoint.path} was sunset on ${endpoint.sunset_at} and is no longer available.`,
      }
      if (endpoint.successor) body.successor = endpoint.successor
      if (endpoint.migration_doc) body.migration_doc = endpoint.migration_doc
      return res.status(410).json(body)
    }

    next()
  }
}

module.exports = { createDusk }
