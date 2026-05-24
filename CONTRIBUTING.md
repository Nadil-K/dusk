# Contributing to dusk

Thank you for your interest in contributing! dusk is a deprecated API usage monitoring middleware and any help improving it is appreciated.

## Getting started

1. Fork the repository and clone your fork.
2. Create a branch: `git checkout -b your-feature-name`
3. Make your changes (see development setup below).
4. Push and open a pull request against `main`.

## Development setup

The repository is a monorepo with three packages. Set up only the one(s) you are working on.

### Python (`packages/python`)

```bash
cd packages/python
python3 -m venv venv
source venv/bin/activate
pip install -e ".[dev]"
pytest
```

### JavaScript (`packages/javascript`)

```bash
cd packages/javascript
npm install
npm test
```

### Go (`packages/go`)

```bash
cd packages/go
go test ./...
```

## Project structure

```
packages/
  python/       # dusk-deprecation — FastAPI/Starlette middleware
  javascript/   # dusk-js — Express middleware
  go/           # Go module — Chi middleware
examples/
  fastapi/      # runnable FastAPI example
  express/      # runnable Express example
  chi/          # runnable Chi example
```

## Adding a new framework adapter

New adapters are very welcome — see the roadmap in [README.md](README.md).

Each adapter follows the same pattern:
1. Load config via the core `LoadConfig` / `load_config` function.
2. Match incoming requests against the endpoint list using the core route matcher.
3. Inject RFC 8594 headers (`Deprecation`, `Sunset`, `Link`) on matches.
4. Return `410 Gone` if `sunset_at` is in the past.
5. Record the hit asynchronously via the store.

Look at an existing adapter for reference:
- Python: `packages/python/src/dusk/adapters/fastapi/`
- JavaScript: `packages/javascript/src/adapters/express.ts`
- Go: `packages/go/adapters/chi/middleware.go`

## Pull request guidelines

- Keep PRs focused — one feature or fix per PR.
- Add or update tests to cover your changes.
- Run the full test suite for the package(s) you touched before opening the PR.
- Follow the existing code style — no new linters or formatters will be introduced without discussion.

## Reporting bugs

Open an issue and include:
- Which package (Python / JavaScript / Go) and version.
- Minimal reproduction steps.
- Expected vs actual behaviour.

## Contact

For questions or discussion: **contact@nadil.me**
