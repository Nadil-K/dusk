from datetime import date
from dusk.core.models import EndpointConfig


class HeaderBuilder:
    def build(self, endpoint: EndpointConfig) -> dict[str, str]:
        headers: dict[str, str] = {}

        headers["Deprecation"] = f'@"{endpoint.deprecated_at}"'

        if endpoint.sunset_at:
            headers["Sunset"] = endpoint.sunset_at

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
