export interface HitEvent {
  ts: Date
  path: string
  method: string
  caller_id: string | null
  user_agent: string | null
  days_until_sunset: number | null
  endpoint_key: string
  enforced: boolean
}

export interface HitQuery {
  endpoint_key?: string
  caller_id?: string
  since_days?: number
  limit?: number
}

export interface EndpointSummary {
  endpoint_key: string
  total_hits: number
  unique_callers: number
  last_seen: Date | null
  top_callers: [string, number][]
}

export interface TotalSummary {
  total_hits: number
  unique_callers: number
  endpoints_with_traffic: number
  past_sunset_with_traffic: number
}

export abstract class HitStore {
  abstract record(hit: HitEvent): Promise<void>
  abstract recentHits(query?: HitQuery): Promise<HitEvent[]>
  abstract endpointSummaries(since_days?: number): Promise<EndpointSummary[]>
  abstract totalSummary(since_days?: number): Promise<TotalSummary>
  async close(): Promise<void> {}
}
