import os
from datetime import datetime, date
from pathlib import Path
from fastapi import FastAPI, Depends, Query
from fastapi.staticfiles import StaticFiles

from dusk.core.config import load_config
from dusk.core.store import create_store
from dusk.core.store.base import HitStore, HitQuery

app = FastAPI(title="dusk dashboard API")

_store: HitStore | None = None
_config = None


def get_config():
    global _config
    if _config is None:
        config_path = os.environ.get("DUSK_CONFIG", "dusk.yaml")
        _config = load_config(config_path)
    return _config


async def get_store() -> HitStore:
    global _store
    if _store is None:
        cfg = get_config()
        backend = os.environ.get("DUSK_STORE_BACKEND")
        if backend:
            cfg.store.backend = backend
        url = os.environ.get("DUSK_STORE_URL")
        if url:
            cfg.store.url = url
        _store = create_store(cfg)
        await _store.start()
    return _store


@app.on_event("startup")
async def startup():
    await get_store()


@app.on_event("shutdown")
async def shutdown():
    if _store:
        await _store.close()


@app.get("/api/summary")
async def summary(store: HitStore = Depends(get_store), since_days: int = 30):
    return await store.total_summary(since_days=since_days)


@app.get("/api/endpoints")
async def endpoints(
    store: HitStore = Depends(get_store),
    cfg=Depends(get_config),
    since_days: int = 30,
):
    summaries = await store.endpoint_summaries(since_days=since_days)
    summary_map = {s.endpoint_key: s for s in summaries}
    today = date.today()
    result = []

    for ep in cfg.endpoints:
        s = summary_map.get(ep.path)
        days_left = None
        past_sunset = False

        if ep.sunset_at:
            sunset = date.fromisoformat(ep.sunset_at)
            days_left = (sunset - today).days
            past_sunset = days_left < 0

        result.append({
            "path": ep.path,
            "methods": ep.methods,
            "deprecated_at": ep.deprecated_at,
            "sunset_at": ep.sunset_at,
            "successor": ep.successor,
            "days_left": days_left,
            "past_sunset": past_sunset,
            "total_hits": s.total_hits if s else 0,
            "unique_callers": s.unique_callers if s else 0,
            "last_seen": s.last_seen.isoformat() if s and s.last_seen else None,
            "top_callers": s.top_callers if s else [],
        })

    return result


@app.get("/api/hits")
async def hits(
    store: HitStore = Depends(get_store),
    limit: int = Query(20, ge=1, le=500),
    endpoint: str | None = None,
    since_days: int = 30,
):
    query = HitQuery(endpoint_key=endpoint, limit=limit, since_days=since_days)
    raw = await store.recent_hits(query)
    return [
        {
            "ts": h.ts.isoformat(),
            "path": h.path,
            "method": h.method,
            "caller_id": h.caller_id,
            "user_agent": h.user_agent,
            "days_left": h.days_until_sunset,
            "endpoint_key": h.endpoint_key,
            "enforced": h.enforced,
        }
        for h in raw
    ]


@app.get("/api/callers")
async def all_callers(store: HitStore = Depends(get_store), since_days: int = 30):
    top = await store.top_callers(endpoint_key=None, since_days=since_days)
    return {"endpoint_key": None, "top_callers": top}


@app.get("/api/callers/{endpoint_key:path}")
async def callers(endpoint_key: str, store: HitStore = Depends(get_store), since_days: int = 30):
    top = await store.top_callers(endpoint_key=endpoint_key, since_days=since_days)
    return {"endpoint_key": endpoint_key, "top_callers": top}


# Serve compiled Svelte app — must be mounted last
_dist = Path(__file__).resolve().parent.parent / "frontend" / "dist"
if _dist.exists():
    app.mount("/", StaticFiles(directory=str(_dist), html=True), name="static")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="127.0.0.1", port=9001)
