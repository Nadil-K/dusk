import { DuskConfig } from '../models'
import { HitStore } from './base'
import { AsyncWriteBuffer } from './buffer'
import { SQLiteHitStore } from './sqlite'
import { RedisHitStore } from './redis'

export { HitStore } from './base'
export { AsyncWriteBuffer } from './buffer'
export { SQLiteHitStore } from './sqlite'
export { RedisHitStore } from './redis'
export type { HitEvent, HitQuery, EndpointSummary, TotalSummary } from './base'

export function createStore(config: DuskConfig): AsyncWriteBuffer {
  let backend: HitStore

  if (config.store.backend === 'sqlite') {
    backend = new SQLiteHitStore(config.store.path ?? '.dusk/hits.db')
  } else if (config.store.backend === 'redis') {
    backend = new RedisHitStore({
      url: config.store.url,
      keyPrefix: config.store.key_prefix,
      ttlDays: config.store.ttl_days,
    })
  } else {
    throw new Error(`Unknown store backend: ${(config.store as { backend: string }).backend}`)
  }

  const buffer = new AsyncWriteBuffer(backend)
  buffer.start()
  return buffer
}
