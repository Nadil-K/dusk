import asyncio
import pytest
import tempfile
import os
from datetime import datetime, timezone

from core.store.sqlite import SQLiteHitStore
from core.store.base import HitEvent, HitQuery


def make_hit(path="/api/v1/users", caller="key-abc", days_left=100):
    return HitEvent(
        ts=datetime.now(timezone.utc),
        path=path,
        method="GET",
        caller_id=caller,
        user_agent="test-agent/1.0",
        days_until_sunset=days_left,
        endpoint_key=path,
    )


@pytest.fixture
def store(tmp_path):
    db_path = str(tmp_path / "test.db")
    return SQLiteHitStore(path=db_path)


@pytest.mark.asyncio
async def test_record_and_recent_hits(store):
    await store.record(make_hit())
    hits = await store.recent_hits(HitQuery(limit=10))
    assert len(hits) == 1
    assert hits[0].path == "/api/v1/users"


@pytest.mark.asyncio
async def test_endpoint_summaries(store):
    await store.record(make_hit("/api/v1/users", caller="key-a"))
    await store.record(make_hit("/api/v1/users", caller="key-b"))
    await store.record(make_hit("/api/v1/orders/1", caller="key-a"))

    summaries = await store.endpoint_summaries()
    assert len(summaries) == 2
    users = next(s for s in summaries if s.endpoint_key == "/api/v1/users")
    assert users.total_hits == 2
    assert users.unique_callers == 2


@pytest.mark.asyncio
async def test_total_summary(store):
    await store.record(make_hit("/api/v1/users", caller="key-a"))
    await store.record(make_hit("/api/v1/orders/1", caller="key-b", days_left=-5))

    summary = await store.total_summary()
    assert summary["total_hits"] == 2
    assert summary["endpoints_with_traffic"] == 2
    assert summary["past_sunset_with_traffic"] == 1


@pytest.mark.asyncio
async def test_filter_by_endpoint(store):
    await store.record(make_hit("/api/v1/users"))
    await store.record(make_hit("/api/v1/orders/1"))

    hits = await store.recent_hits(HitQuery(endpoint_key="/api/v1/users"))
    assert all(h.endpoint_key == "/api/v1/users" for h in hits)
    assert len(hits) == 1
