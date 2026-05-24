from core.store.base import HitStore
from core.store.buffer import AsyncWriteBuffer


def create_store(config) -> HitStore:
    backend_name = config.store.backend

    if backend_name == "sqlite":
        from core.store.sqlite import SQLiteHitStore
        backend = SQLiteHitStore(path=config.store.path)

    elif backend_name == "redis":
        from core.store.redis import RedisHitStore
        backend = RedisHitStore(
            url=config.store.url,
            key_prefix=config.store.key_prefix or "dusk",
            ttl_days=config.store.ttl_days or 90,
        )

    else:
        raise ValueError(f"Unknown store backend: {backend_name!r}")

    return AsyncWriteBuffer(backend, flush_interval=2.0)
