import re
from dusk.core.models import EndpointConfig


_PARAM_RE = re.compile(r"\{[^}]+\}")


def _path_to_regex(path: str) -> re.Pattern:
    parts = _PARAM_RE.split(path)
    escaped = [re.escape(p) for p in parts]
    pattern = "[^/]+".join(escaped)
    return re.compile(f"^{pattern}$")


class RouteMatcher:
    def __init__(self, endpoints: list[EndpointConfig]):
        self._routes = [
            (_path_to_regex(ep.path), ep)
            for ep in endpoints
        ]

    def match(self, path: str, method: str) -> EndpointConfig | None:
        method = method.upper()
        for pattern, ep in self._routes:
            if pattern.match(path) and method in ep.methods:
                return ep
        return None
