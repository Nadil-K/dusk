# dusk

RFC 8594-compliant API deprecation middleware with pluggable storage and an optional dashboard.

## Quick start

```bash
pip install dusk-deprecation[fastapi]
```

```python
from fastapi import FastAPI
from adapters.fastapi import DuskMiddleware

app = FastAPI()
app.add_middleware(DuskMiddleware, config_path="dusk.yaml")
```

Add a `dusk.yaml` to your project root and dusk will inject `Deprecation`, `Sunset`, and `Link` headers on matched endpoints, and return `410 Gone` once a sunset date passes.

## Configuration

See `dusk.yaml` for a full example. Switch from SQLite to Redis with one line:

```yaml
store:
  backend: redis
  url: redis://localhost:6379
```

## CLI

```bash
dusk status          # show all endpoint statuses
dusk report          # hit summary from the store
dusk check           # CI-friendly: exit 1 if past-sunset endpoints have traffic
dusk dashboard       # launch the web dashboard on :9001
```

## License

MIT
