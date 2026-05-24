export interface Summary {
  total_hits: number
  unique_callers: number
  endpoints_with_traffic: number
  past_sunset_with_traffic: number
}

export interface EndpointRow {
  path: string
  methods: string[]
  deprecated_at: string
  sunset_at: string | null
  successor: string | null
  days_left: number | null
  past_sunset: boolean
  total_hits: number
  unique_callers: number
  last_seen: string | null
  top_callers: [string, number][]
}

export interface HitRow {
  ts: string
  path: string
  method: string
  caller_id: string | null
  user_agent: string | null
  days_left: number | null
  endpoint_key: string
  enforced: boolean
}

export interface CallerData {
  endpoint_key: string
  top_callers: [string, number][]
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(path)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json() as Promise<T>
}

export const fetchSummary = (since_days = 30): Promise<Summary> =>
  get(`/api/summary?since_days=${since_days}`)

export const fetchEndpoints = (since_days = 30): Promise<EndpointRow[]> =>
  get(`/api/endpoints?since_days=${since_days}`)

export const fetchHits = ({
  limit = 20,
  endpoint = null,
  since_days = 30,
}: { limit?: number; endpoint?: string | null; since_days?: number } = {}): Promise<HitRow[]> => {
  let url = `/api/hits?limit=${limit}&since_days=${since_days}`
  if (endpoint) url += `&endpoint=${encodeURIComponent(endpoint)}`
  return get(url)
}

export const fetchCallers = (endpoint_key: string): Promise<CallerData> =>
  get(`/api/callers/${encodeURIComponent(endpoint_key)}`)
