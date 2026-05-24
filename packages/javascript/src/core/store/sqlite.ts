import fs from 'fs'
import path from 'path'
import { HitStore, HitEvent, HitQuery, EndpointSummary, TotalSummary } from './base'

// Dynamic import so the package installs without better-sqlite3 if not needed
// eslint-disable-next-line @typescript-eslint/no-require-imports
type BetterSqlite3 = typeof import('better-sqlite3')
type Database = InstanceType<BetterSqlite3>

export class SQLiteHitStore extends HitStore {
  private db: Database | null = null

  constructor(private readonly dbPath: string = '.dusk/hits.db') {
    super()
  }

  private getDb(): Database {
    if (!this.db) {
      // eslint-disable-next-line @typescript-eslint/no-require-imports
      const Database = require('better-sqlite3') as BetterSqlite3
      fs.mkdirSync(path.dirname(this.dbPath), { recursive: true })
      const db = new Database(this.dbPath)
      db.pragma('journal_mode = WAL')
      db.exec(`
        CREATE TABLE IF NOT EXISTS hits (
          ts            TEXT NOT NULL,
          path          TEXT NOT NULL,
          method        TEXT NOT NULL,
          caller_id     TEXT,
          user_agent    TEXT,
          days_left     INTEGER,
          endpoint_key  TEXT NOT NULL,
          enforced      INTEGER NOT NULL DEFAULT 0
        );
        CREATE INDEX IF NOT EXISTS idx_endpoint ON hits(endpoint_key);
        CREATE INDEX IF NOT EXISTS idx_ts ON hits(ts);
      `)
      this.db = db
    }
    return this.db!
  }

  async record(hit: HitEvent): Promise<void> {
    this.getDb()
      .prepare('INSERT INTO hits VALUES (?,?,?,?,?,?,?,?)')
      .run(
        hit.ts.toISOString(),
        hit.path,
        hit.method,
        hit.caller_id,
        hit.user_agent,
        hit.days_until_sunset,
        hit.endpoint_key,
        hit.enforced ? 1 : 0,
      )
  }

  async recentHits(query: HitQuery = {}): Promise<HitEvent[]> {
    const { endpoint_key, caller_id, since_days = 30, limit = 100 } = query
    const since = new Date(Date.now() - since_days * 86_400_000).toISOString()

    let sql = `
      SELECT ts, path, method, caller_id, user_agent, days_left, endpoint_key, enforced
      FROM hits
      WHERE ts >= ?
    `
    const params: unknown[] = [since]

    if (endpoint_key) { sql += ' AND endpoint_key = ?'; params.push(endpoint_key) }
    if (caller_id)    { sql += ' AND caller_id = ?';    params.push(caller_id) }

    sql += ' ORDER BY ts DESC LIMIT ?'
    params.push(limit)

    const rows = this.getDb().prepare(sql).all(...params) as Array<Record<string, unknown>>
    return rows.map((r) => ({
      ts: new Date(r.ts as string),
      path: r.path as string,
      method: r.method as string,
      caller_id: (r.caller_id as string | null) ?? null,
      user_agent: (r.user_agent as string | null) ?? null,
      days_until_sunset: r.days_left as number | null,
      endpoint_key: r.endpoint_key as string,
      enforced: (r.enforced as number) === 1,
    }))
  }

  async endpointSummaries(since_days = 30): Promise<EndpointSummary[]> {
    const since = new Date(Date.now() - since_days * 86_400_000).toISOString()
    const db = this.getDb()

    const rows = db.prepare(`
      SELECT endpoint_key,
             COUNT(*)                    AS total_hits,
             COUNT(DISTINCT caller_id)   AS unique_callers,
             MAX(ts)                     AS last_seen
      FROM hits
      WHERE ts >= ?
      GROUP BY endpoint_key
      ORDER BY total_hits DESC
    `).all(since) as Array<Record<string, unknown>>

    return rows.map((row) => {
      const callers = db.prepare(`
        SELECT caller_id, COUNT(*) AS cnt
        FROM hits
        WHERE endpoint_key = ? AND ts >= ? AND caller_id IS NOT NULL
        GROUP BY caller_id
        ORDER BY cnt DESC
        LIMIT 10
      `).all(row.endpoint_key, since) as Array<Record<string, unknown>>

      return {
        endpoint_key: row.endpoint_key as string,
        total_hits: row.total_hits as number,
        unique_callers: row.unique_callers as number,
        last_seen: row.last_seen ? new Date(row.last_seen as string) : null,
        top_callers: callers.map((c) => [c.caller_id as string, c.cnt as number]),
      }
    })
  }

  async totalSummary(since_days = 30): Promise<TotalSummary> {
    const since = new Date(Date.now() - since_days * 86_400_000).toISOString()
    const db = this.getDb()

    const row = db.prepare(`
      SELECT COUNT(*)                      AS total_hits,
             COUNT(DISTINCT caller_id)     AS unique_callers,
             COUNT(DISTINCT endpoint_key)  AS endpoints_with_traffic
      FROM hits WHERE ts >= ?
    `).get(since) as Record<string, number>

    const past = db.prepare(`
      SELECT COUNT(DISTINCT endpoint_key)
      FROM hits WHERE ts >= ? AND days_left < 0
    `).pluck().get(since) as number

    return {
      total_hits: row.total_hits,
      unique_callers: row.unique_callers,
      endpoints_with_traffic: row.endpoints_with_traffic,
      past_sunset_with_traffic: past,
    }
  }

  async close(): Promise<void> {
    if (this.db) {
      this.db.close()
      this.db = null
    }
  }
}
