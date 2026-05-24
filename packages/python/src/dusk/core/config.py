import yaml
from pathlib import Path
from dusk.core.models import DeprecationConfig, StoreConfig, LogConfig, EndpointConfig


def load_config(path: str = "dusk.yaml") -> DeprecationConfig:
    config_dir = Path(path).resolve().parent
    raw = yaml.safe_load(Path(path).read_text())
    dusk = raw["dusk"]

    store_raw = dusk.get("store", {})
    sqlite_path = store_raw.get("path", ".dusk/hits.db")
    backend = store_raw.get("backend", "sqlite")
    if backend == "sqlite" and not Path(sqlite_path).is_absolute():
        sqlite_path = str(config_dir / sqlite_path)
    store = StoreConfig(
        backend=backend,
        path=sqlite_path,
        url=store_raw.get("url"),
        key_prefix=store_raw.get("key_prefix", "dusk"),
        ttl_days=store_raw.get("ttl_days", 90),
    )

    log_raw = dusk.get("log", {})
    log = LogConfig(identify_by=log_raw.get("identify_by", []))

    endpoints = [
        EndpointConfig(
            path=ep["path"],
            methods=[m.upper() for m in ep.get("methods", ["GET"])],
            deprecated_at=ep["deprecated_at"],
            sunset_at=ep.get("sunset_at"),
            successor=ep.get("successor"),
            migration_doc=ep.get("migration_doc"),
            note=ep.get("note"),
        )
        for ep in dusk.get("endpoints", [])
    ]

    return DeprecationConfig(
        version=dusk.get("version", 1),
        store=store,
        log=log,
        endpoints=endpoints,
    )
