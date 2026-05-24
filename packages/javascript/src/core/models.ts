export interface EndpointConfig {
  path: string
  methods: string[]
  deprecated_at: string
  sunset_at: string | null
  successor?: string
  migration_doc?: string
  note?: string
}

export interface StoreConfig {
  backend: 'sqlite' | 'redis'
  path?: string
  url?: string
  key_prefix?: string
  ttl_days?: number
}

export interface LogConfig {
  identify_by: Array<{ header: string } | 'ip'>
}

export interface DuskConfig {
  version: number
  store: StoreConfig
  log: LogConfig
  endpoints: EndpointConfig[]
}
