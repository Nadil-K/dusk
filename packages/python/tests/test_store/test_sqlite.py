import pytest
from datetime import datetime, timezone

from dusk.core.store.sqlite import SQLiteHitStore
from dusk.core.store.base import HitEvent, HitQuery


def make_hit(path="/api/v1/users", caller="key-abc", days_left=100, enforced=False):
    return HitEvent(
        ts=datetime.now(timezone.utc), path=path, method="GET",
        caller_id=caller, user_agent="test/1.0",
        days_until_sunset=days_left, endpoint_key=path, enforced=enforced,
    )


@pytest.fixture
def store(tmp_path):
    return SQLiteHitStore(path=str(tmp_path / "test.db"))


@pytest.mark.asyncio
async def test_record_and_recent_hits(store):
    await store.record(make_hit())
    hits = await store.recent_hits(HitQuery(limit=10))
    assert len(hits) == 1 and hits[0].path == "/api/v1/users"


@pytest.mark.asyncio
async def test_enforced_hit_is_recorded(store):
    await store.record(make_hit(enforced=True))
    hits = await store.recent_hits(HitQuery())
    assert hits[0].enforced is True


@pytest.mark.asyncio
async def test_endpoint_summaries(store):
    await store.record(make_hit("/api/v1/users", caller="key-a"))
    await store.record(make_hit("/api/v1/users", caller="key-b"))
    await store.record(make_hit("/api/v1/orders/1", caller="key-a"))

    summaries = await store.endpoint_summaries()
    users = next(s for s in summaries if s.endpoint_key == "/api/v1/users")
    assert users.total_hits == 2 and users.unique_callers == 2


@pytest.mark.asyncio
async def test_total_summary(store):
    await store.record(make_hit("/api/v1/users", caller="key-a"))
    await store.record(make_hit("/api/v1/orders/1", caller="key-b", days_left=-5))

    s = await store.total_summary()
    assert s["total_hits"] == 2
    assert s["endpoints_with_traffic"] == 2
    assert s["past_sunset_with_traffic"] == 1


@pytest.mark.asyncio
async def test_filter_by_endpoint(store):
    await store.record(make_hit("/api/v1/users"))
    await store.record(make_hit("/api/v1/orders/1"))

    hits = await store.recent_hits(HitQuery(endpoint_key="/api/v1/users"))
    assert len(hits) == 1 and all(h.endpoint_key == "/api/v1/users" for h in hits)
