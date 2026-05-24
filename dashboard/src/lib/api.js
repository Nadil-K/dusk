const BASE = ''

async function get(path) {
  const res = await fetch(BASE + path)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json()
}

export const fetchSummary = (since_days = 30) =>
  get(`/api/summary?since_days=${since_days}`)

export const fetchEndpoints = (since_days = 30) =>
  get(`/api/endpoints?since_days=${since_days}`)

export const fetchHits = ({ limit = 20, endpoint = null, since_days = 30 } = {}) => {
  let url = `/api/hits?limit=${limit}&since_days=${since_days}`
  if (endpoint) url += `&endpoint=${encodeURIComponent(endpoint)}`
  return get(url)
}

export const fetchCallers = (endpoint_key) =>
  get(`/api/callers/${encodeURIComponent(endpoint_key)}`)
