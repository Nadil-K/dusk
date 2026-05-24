import asyncio
from collections import deque
from dusk.core.store.base import HitStore, HitEvent, HitQuery, EndpointSummary


class AsyncWriteBuffer(HitStore):
    """
    Wraps any HitStore backend. record() is non-blocking — enqueues the hit
    and returns immediately. A background worker drains the queue in batches
    every flush_interval seconds. Drops oldest on overflow (maxlen deque).
    """

    def __init__(
        self,
        backend: HitStore,
        flush_interval: float = 2.0,
        max_buffer: int = 10_000,
        batch_size: int = 100,
    ):
        self._backend = backend
        self._flush_interval = flush_interval
        self._queue: deque[HitEvent] = deque(maxlen=max_buffer)
        self._batch_size = batch_size
        self._task: asyncio.Task | None = None

    async def start(self) -> None:
        self._task = asyncio.create_task(self._flush_worker())

    async def record(self, hit: HitEvent) -> None:
        self._queue.append(hit)

    async def _flush_worker(self) -> None:
        while True:
            await asyncio.sleep(self._flush_interval)
            batch: list[HitEvent] = []
            while self._queue and len(batch) < self._batch_size:
                batch.append(self._queue.popleft())
            if batch:
                for hit in batch:
                    await self._backend.record(hit)

    async def recent_hits(self, query: HitQuery) -> list[HitEvent]:
        return await self._backend.recent_hits(query)

    async def endpoint_summaries(self, since_days: int = 30) -> list[EndpointSummary]:
        return await self._backend.endpoint_summaries(since_days)

    async def total_summary(self, since_days: int = 30) -> dict:
        return await self._backend.total_summary(since_days)

    async def close(self) -> None:
        if self._task:
            self._task.cancel()
        await self._backend.close()
