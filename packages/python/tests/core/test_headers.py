import pytest
from dusk.core.models import EndpointConfig
from dusk.core.headers import HeaderBuilder


def make_ep(sunset_at=None, successor=None, migration_doc=None):
    return EndpointConfig(path="/api/v1/users", methods=["GET"], deprecated_at="2025-01-01",
                          sunset_at=sunset_at, successor=successor, migration_doc=migration_doc)


def test_deprecation_header_always_present():
    headers = HeaderBuilder().build(make_ep())
    assert headers["Deprecation"] == "@1735689600"


def test_sunset_header_when_set():
    assert HeaderBuilder().build(make_ep(sunset_at="2026-01-01"))["Sunset"] == "Thu, 01 Jan 2026 00:00:00 GMT"


def test_no_sunset_header_when_none():
    assert "Sunset" not in HeaderBuilder().build(make_ep())


def test_link_header_with_successor():
    headers = HeaderBuilder().build(make_ep(successor="/api/v2/users"))
    assert "Link" in headers and "successor-version" in headers["Link"]


def test_days_until_sunset_future():
    assert HeaderBuilder().days_until_sunset(make_ep(sunset_at="2099-12-31")) > 0


def test_days_until_sunset_none():
    assert HeaderBuilder().days_until_sunset(make_ep()) is None
