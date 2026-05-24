import json
import redis.asyncio as aioredis
from datetime import datetime, timedelta, timezone
from core.store.base import HitStore, HitEvent, HitQuery, EndpointSummary


class RedisHitStore(HitStore):
    """
    Production backend. Multi-instance safe — all pods write to one Redis.
    Uses Redis Streams for ordered hit events (live tail).
    Uses Redis Sorted Sets for fast caller ranking.
    Uses plain counters for hit totals.
    Auto-expires entries after ttl_days.
    """

    def __init__(
        self,
        url: str = "redis://localhost:6379",
        key_prefix: str = "dusk",
        ttl_days: int = 90,
    ):
        self._redis = aioredis.from_url(url, decode_responses=True)
        self._prefix = key_prefix
        self._ttl = ttl_days * 86_400

    def _stream_key(self) -> str:
        return f"{self._prefix}:hits"

    def _counter_key(self, endpoint_key: str) -> str:
        return f"{self._prefix}:count:{endpoint_key}"

    def _callers_key(self, endpoint_key: str) -> str:
        return f"{self._prefix}:callers:{endpoint_key}"

    def _endpoints_key(self) -> str:
        return f"{self._prefix}:endpoints"

    async def record(self, hit: HitEvent) -> None:
        pipe = self._redis.pipeline()

        pipe.xadd(
            self._stream_key(),
            {
                "ts": hit.ts.isoformat(),
                "path": hit.path,
                "method": hit.method,
                "caller": hit.caller_id or "",
                "ua": hit.user_agent or "",
                "ep": hit.endpoint_key,
                "days_left": str(hit.days_until_sunset) if hit.days_until_sunset is not None else "",
            },
            maxlen=100_000,
            approximate=True,
        )

        counter_key = self._counter_key(hit.endpoint_key)
        pipe.incr(counter_key)
        pipe.expire(counter_key, self._ttl)

        callers_key = self._callers_key(hit.endpoint_key)
        pipe.zincrby(callers_key, 1, hit.caller_id or "anonymous")
        pipe.expire(callers_key, self._ttl)

        pipe.sadd(self._endpoints_key(), hit.endpoint_key)
        pipe.expire(self._endpoints_key(), self._ttl)

        await pipe.execute()

    def _deserialize(self, entry: tuple) -> HitEvent:
        _, fields = entry
        days_left_raw = fields.get("days_left", "")
        return HitEvent(
            ts=datetime.fromisoformat(fields["ts"]),
            path=fields["path"],
            method=fields["method"],
            caller_id=fields["caller"] or None,
            user_agent=fields["ua"] or None,
            days_until_sunset=int(days_left_raw) if days_left_raw else None,
            endpoint_key=fields["ep"],
        )

    async def recent_hits(self, query: HitQuery) -> list[HitEvent]:
        entries = await self._redis.xrevrange(
            self._stream_key(), count=query.limit
        )
        hits = [self._deserialize(e) for e in entries]

        if query.endpoint_key:
            hits = [h for h in hits if h.endpoint_key == query.endpoint_key]
        if query.caller_id:
            hits = [h for h in hits if h.caller_id == query.caller_id]

        return hits

    async def endpoint_summaries(self, since_days: int = 30) -> list[EndpointSummary]:
        endpoints = await self._redis.smembers(self._endpoints_key())
        summaries: list[EndpointSummary] = []

        for ep_key in endpoints:
            total = int(await self._redis.get(self._counter_key(ep_key)) or 0)
            top_raw = await self._redis.zrevrange(
                self._callers_key(ep_key), 0, 9, withscores=True
            )
            top_callers = [(caller, int(score)) for caller, score in top_raw]
            unique_callers = await self._redis.zcard(self._callers_key(ep_key))

            summaries.append(
                EndpointSummary(
                    endpoint_key=ep_key,
                    total_hits=total,
                    unique_callers=unique_callers,
                    last_seen=None,
                    top_callers=top_callers,
                )
            )

        return sorted(summaries, key=lambda s: s.total_hits, reverse=True)

    async def total_summary(self, since_days: int = 30) -> dict:
        endpoints = await self._redis.smembers(self._endpoints_key())
        total_hits = 0
        past_sunset_with_traffic = 0

        pipe = self._redis.pipeline()
        for ep in endpoints:
            pipe.get(self._counter_key(ep))
        counts = await pipe.execute()

        for ep, count in zip(endpoints, counts):
            c = int(count or 0)
            total_hits += c

        # unique callers: union of all caller sorted sets
        unique_callers = 0
        if endpoints:
            all_caller_keys = [self._callers_key(ep) for ep in endpoints]
            unique_callers = await self._redis.zunionstore(
                f"{self._prefix}:_tmp_callers", all_caller_keys
            )
            await self._redis.delete(f"{self._prefix}:_tmp_callers")

        return {
            "total_hits": total_hits,
            "unique_callers": unique_callers,
            "endpoints_with_traffic": len(endpoints),
            "past_sunset_with_traffic": past_sunset_with_traffic,
        }

    async def close(self) -> None:
        await self._redis.aclose()
