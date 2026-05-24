from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional


@dataclass
class EndpointConfig:
    path: str
    methods: list[str]
    deprecated_at: str
    sunset_at: Optional[str]
    successor: Optional[str] = None
    migration_doc: Optional[str] = None
    note: Optional[str] = None


@dataclass
class StoreConfig:
    backend: str = "sqlite"
    path: str = ".dusk/hits.db"
    url: Optional[str] = None
    key_prefix: str = "dusk"
    ttl_days: int = 90


@dataclass
class LogConfig:
    identify_by: list[dict] = field(default_factory=list)


@dataclass
class DeprecationConfig:
    version: int
    store: StoreConfig
    log: LogConfig
    endpoints: list[EndpointConfig]
