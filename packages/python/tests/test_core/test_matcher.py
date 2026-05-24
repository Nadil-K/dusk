import pytest
from dusk.core.models import EndpointConfig
from dusk.core.matcher import RouteMatcher


def make_ep(path, methods=None):
    return EndpointConfig(path=path, methods=methods or ["GET"], deprecated_at="2025-01-01", sunset_at="2026-01-01")


def test_exact_match():
    assert RouteMatcher([make_ep("/api/v1/users")]).match("/api/v1/users", "GET") is not None


def test_no_match_wrong_path():
    assert RouteMatcher([make_ep("/api/v1/users")]).match("/api/v2/users", "GET") is None


def test_no_match_wrong_method():
    assert RouteMatcher([make_ep("/api/v1/users", ["POST"])]).match("/api/v1/users", "GET") is None


def test_param_match():
    matcher = RouteMatcher([make_ep("/api/v1/orders/{id}")])
    assert matcher.match("/api/v1/orders/123", "GET") is not None
    assert matcher.match("/api/v1/orders/abc-xyz", "GET") is not None


def test_param_no_match_extra_segments():
    assert RouteMatcher([make_ep("/api/v1/orders/{id}")]).match("/api/v1/orders/123/items", "GET") is None


def test_method_case_insensitive():
    assert RouteMatcher([make_ep("/api/v1/users", ["GET"])]).match("/api/v1/users", "get") is not None
