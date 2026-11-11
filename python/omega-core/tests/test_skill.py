"""omega_core.skill: the GSL reference the bundled gsl prints for a model. Standard library and pytest only."""

from __future__ import annotations

import pytest

import omega_core


def test_the_index_names_the_packages_and_how_to_read_on() -> None:
    text = omega_core.skill()
    assert "### Package Index" in text
    assert "- **`material/bulk/wood`**" in text
    assert "gsl skill PACKAGE..." in text
    assert len(text) < len(omega_core.skill(everything=True)) / 4


def test_a_package_by_prefix_or_path() -> None:
    assert omega_core.skill("wood") == omega_core.skill("material/bulk/wood")
    assert "### `material/bulk/wood`" in omega_core.skill("wood")


def test_an_unknown_package_is_a_value_error() -> None:
    with pytest.raises(ValueError, match="no library"):
        omega_core.skill("nope")
