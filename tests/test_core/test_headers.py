import pytest
from core.models import EndpointConfig
from core.headers import HeaderBuilder


def make_ep(sunset_at=None, successor=None, migration_doc=None):
    return EndpointConfig(
        path="/api/v1/users",
        methods=["GET"],
        deprecated_at="2025-01-01",
        sunset_at=sunset_at,
        successor=successor,
        migration_doc=migration_doc,
    )


def test_deprecation_header_always_present():
    builder = HeaderBuilder()
    headers = builder.build(make_ep())
    assert "Deprecation" in headers
    assert "2025-01-01" in headers["Deprecation"]


def test_sunset_header_when_set():
    builder = HeaderBuilder()
    headers = builder.build(make_ep(sunset_at="2026-01-01"))
    assert headers["Sunset"] == "2026-01-01"


def test_no_sunset_header_when_none():
    builder = HeaderBuilder()
    headers = builder.build(make_ep())
    assert "Sunset" not in headers


def test_link_header_with_successor():
    builder = HeaderBuilder()
    headers = builder.build(make_ep(successor="/api/v2/users"))
    assert "Link" in headers
    assert "successor-version" in headers["Link"]


def test_days_until_sunset_future():
    builder = HeaderBuilder()
    days = builder.days_until_sunset(make_ep(sunset_at="2099-12-31"))
    assert days > 0


def test_days_until_sunset_none():
    builder = HeaderBuilder()
    assert builder.days_until_sunset(make_ep()) is None
