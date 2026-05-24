import aiosqlite
from datetime import datetime
from pathlib import Path
from dusk.core.store.base import HitStore, HitEvent, HitQuery, EndpointSummary


class SQLiteHitStore(HitStore):
    """
    Default backend. Zero infrastructure — just a file.
    Good for single-instance, local dev, small teams.
    Not suitable for multi-instance deployments (use Redis instead).
    """

    def __init__(self, path: str = ".dusk/hits.db"):
        self._path = path
        self._db: aiosqlite.Connection | None = None

    async def _connect(self) -> None:
        Path(self._path).parent.mkdir(parents=True, exist_ok=True)
        self._db = await aiosqlite.connect(self._path)
        self._db.row_factory = aiosqlite.Row
        await self._db.execute("PRAGMA journal_mode=WAL")
        await self._db.execute("""
            CREATE TABLE IF NOT EXISTS hits (
                ts            TEXT NOT NULL,
                path          TEXT NOT NULL,
                method        TEXT NOT NULL,
                caller_id     TEXT,
                user_agent    TEXT,
                days_left     INTEGER,
                endpoint_key  TEXT NOT NULL,
                enforced      INTEGER NOT NULL DEFAULT 0
            )
        """)
        await self._db.execute(
            "CREATE INDEX IF NOT EXISTS idx_endpoint ON hits(endpoint_key)"
        )
        await self._db.execute(
            "CREATE INDEX IF NOT EXISTS idx_ts ON hits(ts)"
        )
        await self._db.commit()

    async def _ensure_connected(self) -> None:
        if not self._db:
            await self._connect()

    async def record(self, hit: HitEvent) -> None:
        await self._ensure_connected()
        await self._db.execute(
            "INSERT INTO hits VALUES (?,?,?,?,?,?,?,?)",
            (
                hit.ts.isoformat(),
                hit.path,
                hit.method,
                hit.caller_id,
                hit.user_agent,
                hit.days_until_sunset,
                hit.endpoint_key,
                1 if hit.enforced else 0,
            ),
        )
        await self._db.commit()

    async def recent_hits(self, query: HitQuery) -> list[HitEvent]:
        await self._ensure_connected()
        sql = """
            SELECT ts, path, method, caller_id, user_agent, days_left, endpoint_key, enforced
            FROM hits
            WHERE ts >= datetime('now', ? || ' days')
        """
        params: list = [f"-{query.since_days}"]

        if query.endpoint_key:
            sql += " AND endpoint_key = ?"
            params.append(query.endpoint_key)
        if query.caller_id:
            sql += " AND caller_id = ?"
            params.append(query.caller_id)

        sql += " ORDER BY ts DESC LIMIT ?"
        params.append(query.limit)

        rows = await self._db.execute_fetchall(sql, params)
        return [
            HitEvent(
                ts=datetime.fromisoformat(r["ts"]),
                path=r["path"],
                method=r["method"],
                caller_id=r["caller_id"],
                user_agent=r["user_agent"],
                days_until_sunset=r["days_left"],
                endpoint_key=r["endpoint_key"],
                enforced=bool(r["enforced"]),
            )
            for r in rows
        ]

    async def endpoint_summaries(self, since_days: int = 30) -> list[EndpointSummary]:
        await self._ensure_connected()
        rows = await self._db.execute_fetchall(
            """
            SELECT endpoint_key,
                   COUNT(*)                  AS total_hits,
                   COUNT(DISTINCT caller_id) AS unique_callers,
                   MAX(ts)                   AS last_seen
            FROM hits
            WHERE ts >= datetime('now', ? || ' days')
            GROUP BY endpoint_key
            ORDER BY total_hits DESC
            """,
            (f"-{since_days}",),
        )

        summaries: list[EndpointSummary] = []
        for row in rows:
            callers = await self._db.execute_fetchall(
                """
                SELECT caller_id, COUNT(*) AS cnt
                FROM hits
                WHERE endpoint_key = ?
                  AND ts >= datetime('now', ? || ' days')
                  AND caller_id IS NOT NULL
                GROUP BY caller_id
                ORDER BY cnt DESC
                LIMIT 10
                """,
                (row["endpoint_key"], f"-{since_days}"),
            )
            summaries.append(
                EndpointSummary(
                    endpoint_key=row["endpoint_key"],
                    total_hits=row["total_hits"],
                    unique_callers=row["unique_callers"],
                    last_seen=datetime.fromisoformat(row["last_seen"]) if row["last_seen"] else None,
                    top_callers=[(c["caller_id"], c["cnt"]) for c in callers],
                )
            )
        return summaries

    async def total_summary(self, since_days: int = 30) -> dict:
        await self._ensure_connected()
        row = (
            await self._db.execute_fetchall(
                """
                SELECT COUNT(*)                    AS total_hits,
                       COUNT(DISTINCT caller_id)   AS unique_callers,
                       COUNT(DISTINCT endpoint_key) AS endpoints_with_traffic
                FROM hits
                WHERE ts >= datetime('now', ? || ' days')
                """,
                (f"-{since_days}",),
            )
        )[0]

        past_sunset = (
            await self._db.execute_fetchall(
                """
                SELECT COUNT(DISTINCT endpoint_key)
                FROM hits
                WHERE ts >= datetime('now', ? || ' days')
                  AND days_left < 0
                """,
                (f"-{since_days}",),
            )
        )[0][0]

        return {
            "total_hits": row["total_hits"],
            "unique_callers": row["unique_callers"],
            "endpoints_with_traffic": row["endpoints_with_traffic"],
            "past_sunset_with_traffic": past_sunset,
        }

    async def top_callers(
        self, endpoint_key: str | None = None, since_days: int = 30, limit: int = 10
    ) -> list[tuple[str, int]]:
        await self._ensure_connected()
        sql = """
            SELECT caller_id, COUNT(*) AS cnt
            FROM hits
            WHERE ts >= datetime('now', ? || ' days')
              AND caller_id IS NOT NULL
        """
        params: list = [f"-{since_days}"]
        if endpoint_key is not None:
            sql += " AND endpoint_key = ?"
            params.append(endpoint_key)
        sql += " GROUP BY caller_id ORDER BY cnt DESC LIMIT ?"
        params.append(limit)
        rows = await self._db.execute_fetchall(sql, params)
        return [(r["caller_id"], r["cnt"]) for r in rows]

    async def close(self) -> None:
        if self._db:
            await self._db.close()
            self._db = None
