from datetime import date, datetime, timezone
from dusk.core.models import EndpointConfig


def _to_unix(iso_date: str) -> int:
    d = date.fromisoformat(iso_date)
    return int(datetime(d.year, d.month, d.day, tzinfo=timezone.utc).timestamp())


def _to_http_date(iso_date: str) -> str:
    d = date.fromisoformat(iso_date)
    return datetime(d.year, d.month, d.day, tzinfo=timezone.utc).strftime("%a, %d %b %Y %H:%M:%S GMT")


class HeaderBuilder:
    def build(self, endpoint: EndpointConfig) -> dict[str, str]:
        headers: dict[str, str] = {}

        headers["Deprecation"] = f"@{_to_unix(endpoint.deprecated_at)}"

        if endpoint.sunset_at:
            headers["Sunset"] = _to_http_date(endpoint.sunset_at)

        links: list[str] = []
        if endpoint.successor:
            links.append(f'<{endpoint.successor}>; rel="successor-version"')
        if endpoint.migration_doc:
            links.append(f'<{endpoint.migration_doc}>; rel="deprecation"')

        if links:
            headers["Link"] = ", ".join(links)

        return headers

    def days_until_sunset(self, endpoint: EndpointConfig) -> int | None:
        if not endpoint.sunset_at:
            return None
        sunset = date.fromisoformat(endpoint.sunset_at)
        return (sunset - date.today()).days
