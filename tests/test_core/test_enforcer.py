import pytest
from core.models import EndpointConfig
from core.enforcer import SunsetEnforcer


def make_ep(sunset_at, successor=None):
    return EndpointConfig(
        path="/api/v1/users",
        methods=["GET"],
        deprecated_at="2025-01-01",
        sunset_at=sunset_at,
        successor=successor,
    )


def test_no_sunset_never_enforces():
    enforcer = SunsetEnforcer()
    enforce, body = enforcer.check(make_ep(None))
    assert enforce is False
    assert body is None


def test_future_sunset_not_enforced():
    enforcer = SunsetEnforcer()
    enforce, body = enforcer.check(make_ep("2099-12-31"))
    assert enforce is False


def test_past_sunset_enforced():
    enforcer = SunsetEnforcer()
    enforce, body = enforcer.check(make_ep("2000-01-01"))
    assert enforce is True
    assert body["error"] == "Gone"


def test_past_sunset_body_includes_successor():
    enforcer = SunsetEnforcer()
    _, body = enforcer.check(make_ep("2000-01-01", successor="/api/v2/users"))
    assert body["successor"] == "/api/v2/users"
