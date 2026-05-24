from datetime import date
from dusk.core.models import EndpointConfig


class SunsetEnforcer:
    def check(self, endpoint: EndpointConfig) -> tuple[bool, dict | None]:
        """Returns (should_enforce_410, response_body_or_None)."""
        if not endpoint.sunset_at:
            return False, None

        sunset = date.fromisoformat(endpoint.sunset_at)
        if date.today() <= sunset:
            return False, None

        body: dict = {
            "error": "Gone",
            "message": f"{endpoint.path} was sunset on {endpoint.sunset_at} and is no longer available.",
        }
        if endpoint.successor:
            body["successor"] = endpoint.successor
        if endpoint.migration_doc:
            body["migration_doc"] = endpoint.migration_doc

        return True, body
