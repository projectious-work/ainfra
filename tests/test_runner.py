"""Subprocess boundary helper tests."""

from ainfra.runner import redact


def test_redacts_every_known_secret() -> None:
    assert redact("before TOKEN after", ["TOKEN"]) == "before [REDACTED] after"


def test_empty_secret_does_not_replace_every_character() -> None:
    assert redact("unchanged", [""]) == "unchanged"
