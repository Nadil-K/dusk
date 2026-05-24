import asyncio
import sys
from datetime import date
import click

from core.config import load_config
from core.store import create_store


@click.group()
def cli():
    """dusk — RFC 8594 API deprecation manager."""
    pass


@cli.command()
@click.option("--config", default="dusk.yaml", show_default=True)
def status(config):
    """Show deprecation status for all configured endpoints."""
    try:
        cfg = load_config(config)
    except FileNotFoundError:
        click.echo(f"Config file not found: {config}", err=True)
        sys.exit(1)

    today = date.today()
    click.echo(f"{'Endpoint':<40} {'Methods':<15} {'Status':<20} {'Sunset'}")
    click.echo("-" * 95)

    for ep in cfg.endpoints:
        methods = ",".join(ep.methods)
        if ep.sunset_at:
            sunset = date.fromisoformat(ep.sunset_at)
            days_left = (sunset - today).days
            if days_left < 0:
                status_str = click.style("PAST SUNSET", fg="red", bold=True)
                sunset_str = f"{ep.sunset_at} ({abs(days_left)}d ago)"
            elif days_left < 30:
                status_str = click.style(f"{days_left}d left", fg="yellow")
                sunset_str = ep.sunset_at
            else:
                status_str = click.style(f"{days_left}d left", fg="green")
                sunset_str = ep.sunset_at
        else:
            status_str = click.style("soft", fg="blue")
            sunset_str = "none"

        click.echo(f"{ep.path:<40} {methods:<15} {status_str:<20} {sunset_str}")


@cli.command()
@click.option("--config", default="dusk.yaml", show_default=True)
@click.option("--since-days", default=30, show_default=True)
def report(config, since_days):
    """Print a hit report from the store."""
    cfg = load_config(config)
    store = create_store(cfg)

    async def _run():
        from core.store.base import HitQuery
        await store.start()
        try:
            summary = await store.total_summary(since_days=since_days)
            click.echo(f"\nHit summary (last {since_days} days):")
            click.echo(f"  Total hits:              {summary['total_hits']:,}")
            click.echo(f"  Unique callers:          {summary['unique_callers']:,}")
            click.echo(f"  Endpoints with traffic:  {summary['endpoints_with_traffic']}")
            click.echo(f"  Past sunset with traffic:{summary['past_sunset_with_traffic']}")

            summaries = await store.endpoint_summaries(since_days=since_days)
            if summaries:
                click.echo(f"\nPer-endpoint breakdown:")
                for s in summaries:
                    last = s.last_seen.strftime("%Y-%m-%d %H:%M") if s.last_seen else "never"
                    click.echo(
                        f"  {s.endpoint_key:<40} hits={s.total_hits:<8} "
                        f"callers={s.unique_callers:<6} last={last}"
                    )
        finally:
            await store.close()

    asyncio.run(_run())


@cli.command()
@click.option("--config", default="dusk.yaml", show_default=True)
def check(config):
    """Exit non-zero if any past-sunset endpoints have recent traffic (CI-friendly)."""
    cfg = load_config(config)
    store = create_store(cfg)

    async def _run() -> int:
        await store.start()
        try:
            summary = await store.total_summary()
            count = summary.get("past_sunset_with_traffic", 0)
            if count:
                click.echo(
                    click.style(
                        f"FAIL: {count} past-sunset endpoint(s) still receiving traffic.",
                        fg="red",
                        bold=True,
                    )
                )
                return 1
            click.echo(click.style("OK: no past-sunset endpoints with traffic.", fg="green"))
            return 0
        finally:
            await store.close()

    code = asyncio.run(_run())
    sys.exit(code)


@cli.command()
@click.option("--store", "store_override", default=None, help="Path to SQLite DB or Redis URL")
@click.option("--port", default=9001, show_default=True)
@click.option("--config", default="dusk.yaml", show_default=True)
def dashboard(store_override, port, config):
    """Launch the dusk dashboard server."""
    import os
    import uvicorn

    os.environ["DUSK_CONFIG"] = config
    if store_override:
        if store_override.startswith("redis://"):
            os.environ["DUSK_STORE_BACKEND"] = "redis"
            os.environ["DUSK_STORE_URL"] = store_override
        else:
            os.environ["DUSK_STORE_BACKEND"] = "sqlite"
            os.environ["DUSK_STORE_PATH"] = store_override

    click.echo(f"Starting dusk dashboard on http://localhost:{port}")
    uvicorn.run(
        "dashboard_server.server:app",
        host="0.0.0.0",
        port=port,
        reload=False,
    )
