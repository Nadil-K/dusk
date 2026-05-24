import { HitStore, HitEvent, HitQuery, EndpointSummary, TotalSummary } from './base'

export interface RedisHitStoreOptions {
  url?: string
  keyPrefix?: string
  ttlDays?: number
}

export class RedisHitStore extends HitStore {
  private redis: import('ioredis').Redis | null = null
  private readonly prefix: string
  private readonly ttl: number

  constructor({ url = 'redis://localhost:6379', keyPrefix = 'dusk', ttlDays = 90 }: RedisHitStoreOptions = {}) {
    super()
    this.prefix = keyPrefix
    this.ttl = ttlDays * 86_400

    // Dynamic require so ioredis is optional
    // eslint-disable-next-line @typescript-eslint/no-require-imports
    const { Redis } = require('ioredis') as typeof import('ioredis')
    this.redis = new Redis(url)
  }

  private r(): import('ioredis').Redis {
    if (!this.redis) throw new Error('Redis client not initialized')
    return this.redis
  }

  async record(hit: HitEvent): Promise<void> {
    const r = this.r()
    const pipe = r.pipeline()

    pipe.xadd(
      `${this.prefix}:hits`,
      'MAXLEN', '~', '100000',
      '*',
      'ts', hit.ts.toISOString(),
      'path', hit.path,
      'method', hit.method,
      'caller', hit.caller_id ?? '',
      'ua', hit.user_agent ?? '',
      'ep', hit.endpoint_key,
      'days_left', hit.days_until_sunset != null ? String(hit.days_until_sunset) : '',
      'enforced', hit.enforced ? '1' : '0',
    )

    const counterKey = `${this.prefix}:count:${hit.endpoint_key}`
    pipe.incr(counterKey)
    pipe.expire(counterKey, this.ttl)

    const callersKey = `${this.prefix}:callers:${hit.endpoint_key}`
    pipe.zincrby(callersKey, 1, hit.caller_id ?? 'anonymous')
    pipe.expire(callersKey, this.ttl)

    pipe.sadd(`${this.prefix}:endpoints`, hit.endpoint_key)
    pipe.expire(`${this.prefix}:endpoints`, this.ttl)

    await pipe.exec()
  }

  async recentHits(query: HitQuery = {}): Promise<HitEvent[]> {
    const entries = await this.r().xrevrange(`${this.prefix}:hits`, '+', '-', 'COUNT', query.limit ?? 100)
    let hits: HitEvent[] = entries.map(([, fields]) => {
      const f = Object.fromEntries(
        (fields as string[]).reduce<[string, string][]>((acc, v, i, arr) => {
          if (i % 2 === 0) acc.push([v, arr[i + 1]])
          return acc
        }, [])
      )
      return {
        ts: new Date(f.ts),
        path: f.path,
        method: f.method,
        caller_id: f.caller || null,
        user_agent: f.ua || null,
        days_until_sunset: f.days_left ? parseInt(f.days_left, 10) : null,
        endpoint_key: f.ep,
        enforced: f.enforced === '1',
      }
    })

    if (query.endpoint_key) hits = hits.filter((h) => h.endpoint_key === query.endpoint_key)
    if (query.caller_id)    hits = hits.filter((h) => h.caller_id === query.caller_id)
    return hits
  }

  async endpointSummaries(_since_days = 30): Promise<EndpointSummary[]> {
    const r = this.r()
    const endpoints = await r.smembers(`${this.prefix}:endpoints`)
    const summaries: EndpointSummary[] = []

    for (const ep of endpoints) {
      const total = parseInt((await r.get(`${this.prefix}:count:${ep}`)) ?? '0', 10)
      const topRaw = await r.zrevrange(`${this.prefix}:callers:${ep}`, 0, 9, 'WITHSCORES')
      const top_callers: [string, number][] = []
      for (let i = 0; i < topRaw.length; i += 2) {
        top_callers.push([topRaw[i], parseInt(topRaw[i + 1], 10)])
      }
      const unique_callers = await r.zcard(`${this.prefix}:callers:${ep}`)
      summaries.push({ endpoint_key: ep, total_hits: total, unique_callers, last_seen: null, top_callers })
    }

    return summaries.sort((a, b) => b.total_hits - a.total_hits)
  }

  async totalSummary(_since_days = 30): Promise<TotalSummary> {
    const r = this.r()
    const endpoints = await r.smembers(`${this.prefix}:endpoints`)

    let total_hits = 0
    for (const ep of endpoints) {
      total_hits += parseInt((await r.get(`${this.prefix}:count:${ep}`)) ?? '0', 10)
    }

    let unique_callers = 0
    if (endpoints.length > 0) {
      const tmpKey = `${this.prefix}:_tmp_callers`
      unique_callers = await r.zunionstore(tmpKey, endpoints.length, ...endpoints.map((ep) => `${this.prefix}:callers:${ep}`))
      await r.del(tmpKey)
    }

    return {
      total_hits,
      unique_callers,
      endpoints_with_traffic: endpoints.length,
      past_sunset_with_traffic: 0,
    }
  }

  async close(): Promise<void> {
    await this.redis?.quit()
    this.redis = null
  }
}
