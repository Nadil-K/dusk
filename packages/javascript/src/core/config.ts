import fs from 'fs'
import path from 'path'
import yaml from 'js-yaml'
import { DuskConfig, EndpointConfig } from './models'

export function loadConfig(configPath = 'dusk.yaml'): DuskConfig {
  const raw = yaml.load(fs.readFileSync(configPath, 'utf8')) as { dusk: Record<string, unknown> }
  const dusk = raw.dusk
  const configDir = path.dirname(path.resolve(configPath))

  const storeRaw = (dusk.store as Record<string, unknown>) ?? {}
  const logRaw = (dusk.log as Record<string, unknown>) ?? {}
  const endpointsRaw = (dusk.endpoints as Record<string, unknown>[]) ?? []

  const endpoints: EndpointConfig[] = endpointsRaw.map((ep) => ({
    path: ep.path as string,
    methods: ((ep.methods as string[]) ?? ['GET']).map((m) => m.toUpperCase()),
    deprecated_at: ep.deprecated_at as string,
    sunset_at: (ep.sunset_at as string | null) ?? null,
    successor: ep.successor as string | undefined,
    migration_doc: ep.migration_doc as string | undefined,
    note: ep.note as string | undefined,
  }))

  const backend = (storeRaw.backend as 'sqlite' | 'redis') ?? 'sqlite'
  let sqlitePath = (storeRaw.path as string) ?? '.dusk/hits.db'
  if (backend === 'sqlite' && !path.isAbsolute(sqlitePath)) {
    sqlitePath = path.join(configDir, sqlitePath)
  }

  return {
    version: (dusk.version as number) ?? 1,
    store: {
      backend,
      path: sqlitePath,
      url: storeRaw.url as string | undefined,
      key_prefix: (storeRaw.key_prefix as string) ?? 'dusk',
      ttl_days: (storeRaw.ttl_days as number) ?? 90,
    },
    log: {
      identify_by: (logRaw.identify_by as Array<{ header: string } | 'ip'>) ?? [],
    },
    endpoints,
  }
}
