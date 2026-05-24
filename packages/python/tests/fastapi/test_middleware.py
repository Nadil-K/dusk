import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient
from dusk.adapters.fastapi.middleware import DuskMiddleware

DUSK_YAML = """
dusk:
  version: 1
  store:
    backend: sqlite
    path: {db_path}
  log:
    identify_by:
      - header: X-API-Key
  endpoints:
    - path: /api/v1/users
      methods: [GET, POST]
      deprecated_at: "2025-01-01"
      sunset_at: "2099-01-01"
      successor: /api/v2/users
    - path: /api/v1/gone
      methods: [GET]
      deprecated_at: "2020-01-01"
      sunset_at: "2020-06-01"
"""


def _build_app(config_file: str) -> tuple[FastAPI, DuskMiddleware]:
    """
    Returns (inner_app, middleware_instance).
    We wrap the FastAPI app directly so we hold a reference to the exact
    DuskMiddleware instance that handles requests (unlike add_middleware(),
    which creates an internal copy we can't reach).
    """
    app = FastAPI()

    @app.get("/api/v1/users")
    async def users(): return {"users": []}

    @app.get("/api/v1/gone")
    async def gone(): return {}

    @app.get("/api/v2/users")
    async def users_v2(): return {"users": []}

    mw = DuskMiddleware(app, config_path=config_file)
    return app, mw


@pytest.fixture
def config_file(tmp_path):
    db_path = str(tmp_path / "hits.db").replace("\\", "/")
    config = tmp_path / "dusk.yaml"
    config.write_text(DUSK_YAML.format(db_path=db_path))
    return str(config)


@pytest.fixture
def mw_client(config_file):
    _, mw = _build_app(config_file)
    with TestClient(mw, raise_server_exceptions=True) as client:
        yield mw, client


def test_deprecated_endpoint_gets_headers(mw_client):
    _, client = mw_client
    res = client.get("/api/v1/users")
    assert res.status_code == 200
    assert res.headers["Deprecation"] == "@1735689600"
    assert res.headers["Sunset"] == "Thu, 01 Jan 2099 00:00:00 GMT"
    assert "successor-version" in res.headers.get("Link", "")


def test_non_deprecated_no_headers(mw_client):
    _, client = mw_client
    res = client.get("/api/v2/users")
    assert res.status_code == 200
    assert "Deprecation" not in res.headers


def test_past_sunset_returns_410(mw_client):
    _, client = mw_client
    res = client.get("/api/v1/gone")
    assert res.status_code == 410
    assert res.json()["error"] == "Gone"


def test_sunset_hit_is_logged(mw_client):
    """410 responses must be enqueued in the store (enforced=True)."""
    mw, client = mw_client
    client.get("/api/v1/gone")

    # AsyncWriteBuffer enqueues immediately; the 2-second background flush
    # to SQLite hasn't run yet in tests, so check the in-memory queue.
    queue = list(mw.store._queue)
    enforced_hits = [h for h in queue if h.enforced and h.endpoint_key == "/api/v1/gone"]
    assert len(enforced_hits) >= 1, "410 hit was not enqueued in the store"
