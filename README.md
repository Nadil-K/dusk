# dusk

**Deprecated API usage monitoring middleware** for Python, JavaScript, and Go.

Track who is still calling your deprecated endpoints, automatically inject `Deprecation` ([RFC 9745](https://www.rfc-editor.org/rfc/rfc9745)), `Sunset` ([RFC 8594](https://www.rfc-editor.org/rfc/rfc8594)), and `Link` headers, and know with confidence when it is safe to remove an endpoint.

[![Python](https://github.com/Nadil-K/dusk/actions/workflows/ci-python.yml/badge.svg)](https://github.com/Nadil-K/dusk/actions/workflows/ci-python.yml)
[![JavaScript](https://github.com/Nadil-K/dusk/actions/workflows/ci-javascript.yml/badge.svg)](https://github.com/Nadil-K/dusk/actions/workflows/ci-javascript.yml)
[![Go](https://github.com/Nadil-K/dusk/actions/workflows/ci-go.yml/badge.svg)](https://github.com/Nadil-K/dusk/actions/workflows/ci-go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## How it works

1. You declare deprecated endpoints in a `dusk.yaml` config file.
2. dusk middleware intercepts every matching request and logs the hit — who called it, when, and from where.
3. `Deprecation` (RFC 9745), `Sunset` (RFC 8594), and `Link` headers are injected into the response automatically.
4. Once the `sunset_at` date passes, dusk returns `410 Gone` so clients know the endpoint is gone.
5. Use the CLI or dashboard to see which callers are still hitting deprecated endpoints before you remove them.

---

## Packages

| Language | Framework | Package | Install |
|----------|-----------|---------|---------|
| Python | FastAPI / Starlette | [`dusk-deprecation`](packages/python) | `pip install dusk-deprecation[fastapi]` |
| JavaScript | Express | [`dusk-js`](packages/javascript) | `npm install dusk-js` |
| Go | Chi | [`github.com/Nadil-K/dusk/packages/go`](packages/go) | `go get github.com/Nadil-K/dusk/packages/go` |

---

## Quick start

### Python (FastAPI)

```bash
pip install dusk-deprecation[fastapi]
```

```python
from fastapi import FastAPI
from dusk.adapters.fastapi import DuskMiddleware

app = FastAPI()
app.add_middleware(DuskMiddleware, config_path="dusk.yaml")
```

### JavaScript (Express)

```bash
npm install dusk-js
```

```js
const express = require('express')
const { createDusk } = require('dusk-js')

const app = express()
app.use(createDusk('dusk.yaml'))
```

### Go (Chi)

```bash
go get github.com/Nadil-K/dusk/packages/go
```

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    dusk "github.com/Nadil-K/dusk/packages/go/adapters/chi"
)

func main() {
    r := chi.NewRouter()
    r.Use(dusk.New("dusk.yaml"))
    http.ListenAndServe(":8080", r)
}
```

---

## Configuration

Create a `dusk.yaml` in your project root:

```yaml
dusk:
  version: 1

  store:
    backend: sqlite          # or redis (see Storage section)
    path: .dusk/hits.db

  log:
    identify_by:
      - header: X-API-Key    # use API key as caller identity when present
      - header: Authorization
      - ip                   # fall back to IP address

  endpoints:
    - path: /api/v1/users
      methods: [GET, POST]
      deprecated_at: "2025-01-01"
      sunset_at: "2026-01-01"
      successor: /api/v2/users
      migration_doc: https://docs.example.com/migration/users-v1-v2
      note: "v1 lacks pagination and field filtering"

    - path: /api/v1/orders/{id}
      methods: [GET]
      deprecated_at: "2025-03-01"
      sunset_at: "2026-03-01"
      successor: /api/v2/orders/{id}
```

### Fields

| Field | Required | Description |
|-------|----------|-------------|
| `path` | yes | Endpoint path. Supports `{param}` placeholders. |
| `methods` | no | HTTP methods to match. Defaults to `["GET"]`. |
| `deprecated_at` | yes | ISO date when the endpoint was deprecated. Sets the `Deprecation` header. |
| `sunset_at` | no | ISO date when the endpoint will be removed. Sets the `Sunset` header. After this date, dusk returns `410 Gone`. |
| `successor` | no | Path of the replacement endpoint. Included in the `Link` header. |
| `migration_doc` | no | URL to migration documentation. Included in the `Link` header. |
| `note` | no | Human-readable reason for deprecation. |

---

## Storage

dusk supports two storage backends.

### SQLite (default — single instance)

```yaml
store:
  backend: sqlite
  path: .dusk/hits.db   # relative to dusk.yaml
  ttl_days: 90          # prune hits older than this (default: 90)
```

### Redis (production — multi-instance)

```yaml
store:
  backend: redis
  url: redis://localhost:6379
  key_prefix: dusk      # optional namespace prefix
  ttl_days: 90
```

Use Redis when running multiple instances of your application so hits from all instances are aggregated in one place.

---

## CLI (Python)

The `dusk` CLI is available after installing the Python package:

```bash
dusk status          # show all endpoint statuses and sunset countdowns
dusk report          # usage summary — top callers per deprecated endpoint
dusk check           # CI-friendly: exits 1 if past-sunset endpoints still have traffic
dusk dashboard       # launch the web dashboard on :9001
```

---

## Response headers

For every request matched to a deprecated endpoint, dusk injects:

```
Deprecation: @1735689600
Sunset: Thu, 01 Jan 2026 00:00:00 GMT
Link: </api/v2/users>; rel="successor-version",
      <https://docs.example.com/migration/users-v1-v2>; rel="deprecation"
```

| Header | RFC | Format | Description |
|--------|-----|--------|-------------|
| `Deprecation` | [RFC 9745](https://www.rfc-editor.org/rfc/rfc9745) | `@<unix-timestamp>` | Unix timestamp of when the endpoint was deprecated. |
| `Sunset` | [RFC 8594](https://www.rfc-editor.org/rfc/rfc8594) | HTTP-date (`Day, DD Mon YYYY HH:MM:SS GMT`) | Date after which the endpoint will be removed. |
| `Link` | RFC 8594 | URI + rel type | Points to the successor endpoint (`successor-version`) and/or migration docs (`deprecation`). |

After the sunset date, the response is:

```
HTTP/1.1 410 Gone
```

---

## Future improvements

- Additional framework adapters (Flask and Django for Python, Fastify and NestJS for JavaScript, Gin and Echo for Go are on the roadmap — contributions welcome)
- Community-contributed adapters for other languages and frameworks are welcome — PHP (Laravel, Symfony), Ruby (Rails, Sinatra), and Java (Spring Boot) are good starting points
- Dashboard UI improvements
- Webhook / alerting when a past-sunset endpoint receives traffic

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## License

MIT — see [LICENSE](LICENSE).
