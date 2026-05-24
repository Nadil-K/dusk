import pytest
import tempfile
import os
from pathlib import Path
from fastapi import FastAPI
from fastapi.testclient import TestClient


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


@pytest.fixture
def config_file(tmp_path):
    db_path = str(tmp_path / "hits.db").replace("\\", "/")
    cfg = DUSK_YAML.format(db_path=db_path)
    config = tmp_path / "dusk.yaml"
    config.write_text(cfg)
    return str(config)


@pytest.fixture
def client(config_file):
    app = FastAPI()

    @app.get("/api/v1/users")
    async def users():
        return {"users": []}

    @app.get("/api/v1/gone")
    async def gone():
        return {}

    @app.get("/api/v2/users")
    async def users_v2():
        return {"users": []}

    from adapters.fastapi.middleware import DuskMiddleware
    app.add_middleware(DuskMiddleware, config_path=config_file)
    return TestClient(app, raise_server_exceptions=True)


def test_deprecated_endpoint_gets_headers(client):
    res = client.get("/api/v1/users")
    assert res.status_code == 200
    assert "Deprecation" in res.headers
    assert "Sunset" in res.headers
    assert "successor-version" in res.headers.get("Link", "")


def test_non_deprecated_no_headers(client):
    res = client.get("/api/v2/users")
    assert res.status_code == 200
    assert "Deprecation" not in res.headers


def test_past_sunset_returns_410(client):
    res = client.get("/api/v1/gone")
    assert res.status_code == 410
    assert res.json()["error"] == "Gone"
