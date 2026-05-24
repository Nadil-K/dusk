from abc import ABC, abstractmethod
from dataclasses import dataclass
from datetime import datetime
from typing import Optional


@dataclass
class HitEvent:
    ts: datetime
    path: str
    method: str
    caller_id: Optional[str]
    user_agent: Optional[str]
    days_until_sunset: Optional[int]
    endpoint_key: str


@dataclass
class HitQuery:
    endpoint_key: Optional[str] = None
    caller_id: Optional[str] = None
    since_days: int = 30
    limit: int = 100


@dataclass
class EndpointSummary:
    endpoint_key: str
    total_hits: int
    unique_callers: int
    last_seen: Optional[datetime]
    top_callers: list[tuple[str, int]]


class HitStore(ABC):

    @abstractmethod
    async def record(self, hit: HitEvent) -> None:
        ...

    @abstractmethod
    async def recent_hits(self, query: HitQuery) -> list[HitEvent]:
        ...

    @abstractmethod
    async def endpoint_summaries(self, since_days: int = 30) -> list[EndpointSummary]:
        ...

    @abstractmethod
    async def total_summary(self, since_days: int = 30) -> dict:
        ...

    async def close(self) -> None:
        pass
