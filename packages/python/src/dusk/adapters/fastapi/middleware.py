from datetime import datetime, timezone
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import JSONResponse, Response

from dusk.core.config import load_config
from dusk.core.matcher import RouteMatcher
from dusk.core.headers import HeaderBuilder
from dusk.core.enforcer import SunsetEnforcer
from dusk.core.store import create_store
from dusk.core.store.base import HitEvent


class DuskMiddleware(BaseHTTPMiddleware):
    def __init__(self, app, config_path: str = "dusk.yaml"):
        super().__init__(app)
        cfg = load_config(config_path)
        self._cfg = cfg
        self.matcher = RouteMatcher(cfg.endpoints)
        self.builder = HeaderBuilder()
        self.enforcer = SunsetEnforcer()
        self.store = create_store(cfg)
        self._store_started = False

    async def _ensure_store_started(self) -> None:
        if not self._store_started:
            await self.store.start()
            self._store_started = True

    def _resolve_caller(self, request: Request) -> str | None:
        for rule in self._cfg.log.identify_by:
            if isinstance(rule, dict) and "header" in rule:
                value = request.headers.get(rule["header"])
                if value:
                    return value
            elif rule == "ip" and request.client:
                return request.client.host
        return None

    async def dispatch(self, request: Request, call_next) -> Response:
        await self._ensure_store_started()

        endpoint = self.matcher.match(request.url.path, request.method)
        if endpoint is None:
            return await call_next(request)

        should_enforce, gone_body = self.enforcer.check(endpoint)
        headers = self.builder.build(endpoint)
        days_left = self.builder.days_until_sunset(endpoint)
        caller_id = self._resolve_caller(request)

        hit = HitEvent(
            ts=datetime.now(timezone.utc),
            path=request.url.path,
            method=request.method,
            caller_id=caller_id,
            user_agent=request.headers.get("user-agent"),
            days_until_sunset=days_left,
            endpoint_key=endpoint.path,
            enforced=should_enforce,
        )

        # Log the hit regardless of whether we enforce 410 — sunset traffic
        # is exactly what operators need to see in the dashboard.
        await self.store.record(hit)

        if should_enforce:
            return JSONResponse(gone_body, status_code=410, headers=headers)

        response = await call_next(request)
        for k, v in headers.items():
            response.headers[k] = v

        return response
