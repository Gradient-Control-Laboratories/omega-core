"""omega_core.Session's answers checked against closed forms: where a ball's surface is and which way it faces, a
field's extremes and the proofs about them, the box it fills and its volume. Standard library and pytest only, so it
runs against a bare install."""

from __future__ import annotations

import math
from pathlib import Path

import pytest

import omega_core

# A sphere of radius r about the origin; r is a slider, the first name declared.
BALL = "@scene(40)\n@range(5, 15) r = 10.0;\nball = length(p) - r;\nreturn ball;\n"
CUBE = (-20, -20, -20, 20, 20, 20)


@pytest.fixture
def gsl():
    with omega_core.Session() as session:
        yield session


@pytest.fixture
def ball(gsl, tmp_path: Path):
    (tmp_path / "ball.gsl").write_text(BALL)
    return gsl.document(tmp_path / "ball.gsl")


def test_a_point_moved_onto_the_surface_lands_on_it_facing_out(ball) -> None:
    starts = [(20, 5, 3), (0.5, 0.5, 0.5), (-30, 0, 0)]
    for start, sample in zip(starts, ball.surface(starts)):
        out = tuple(c / math.hypot(*start) for c in start)
        assert sample.converged
        assert sample.point == pytest.approx(tuple(10 * c for c in out), abs=1e-3)
        assert sample.normal == pytest.approx(out, abs=1e-3)


def test_an_extremum_brackets_the_true_min_and_max_within_its_tolerance(ball) -> None:
    found = ball.extremum(region=CUBE, mode="both", tolerance=1e-4)
    corner = 20 * math.sqrt(3) - 10
    assert found.min_conditional_bound <= -10 <= found.min
    assert found.min - found.min_conditional_bound <= 1e-4
    assert found.max <= corner <= found.max_conditional_bound + 1e-5
    assert found.max_conditional_bound - found.max <= 1e-4


def test_an_extremum_bound_holds_once_told_the_fields_lipschitz_constant(ball) -> None:
    # 3 * ball's gradient is 3 long: assuming 1, the bound is not a bound; told 3, it is.
    assert ball.extremum("3 * ball", region=CUBE).min_conditional_bound > -30
    assert ball.extremum("3 * ball", region=CUBE, lipschitz=3).min_conditional_bound <= -30 + 1e-5


def test_a_true_claim_is_proven_and_a_false_one_shown_where_it_fails(ball) -> None:
    assert ball.certify(region=CUBE, threshold=-10).proven
    false = ball.certify(region=CUBE, threshold=-9.9)
    assert false.verdict == "disproven"
    assert false.counter_upper < -9.9
    assert all(abs(c) <= 0.1 for c in false.counter_box)  # the counterexample is at the centre


def test_bounds_on_a_fine_grid_find_the_ball_to_a_cell(ball) -> None:
    found = ball.bounds(region=CUBE, resolution=128)
    assert found.found
    for k in range(3):
        cell = found.cell_size[k]
        assert found.bbox_min[k] == pytest.approx(-10, abs=cell)
        assert found.bbox_max[k] == pytest.approx(10, abs=cell)


@pytest.mark.parametrize("name, text, exact", [
    ("ball", "ball = length(p) - 10;\nreturn ball;", 4 / 3 * math.pi * 1000),
    ("shell", "skin = abs(length(p) - 9) - 1;\nreturn skin;", 4 / 3 * math.pi * (1000 - 512)),
    ("sharp box", "blk = boxCentered([0, 0, 0], [20, 20, 20]);\nreturn blk;", 8000.0),
    ("two returns", "a = length(p - [-12, 0, 0]) - 5;\nb = length(p - [12, 0, 0]) - 5;\nreturn a, b;",
     2 * 4 / 3 * math.pi * 125),
])
def test_a_volume_brackets_the_exact_one(gsl, name: str, text: str, exact: float) -> None:
    found = gsl.code("@scene(40)\n" + text + "\n").volume(region=CUBE, relative_tolerance=0.05)
    assert found.lower_mm3 <= exact <= found.upper_mm3, name


def test_a_value_that_is_not_finite_is_none(ball) -> None:
    assert ball.eval("1 / ball", points=[(10, 0, 0)])[0].values == (None,)


def test_a_document_is_read_again_for_each_question(gsl, tmp_path: Path) -> None:
    path = tmp_path / "part.gsl"
    path.write_text("@scene(40)\nball = length(p) - 10;\nreturn ball;\n")
    part = gsl.document(path)
    assert part.eval("ball", points=[(20, 0, 0)])[0].values == pytest.approx((10,))
    path.write_text("@scene(40)\nball = length(p) - 12;\nreturn ball;\n")
    assert part.eval("ball", points=[(20, 0, 0)])[0].values == pytest.approx((8,))


def test_an_unknown_op_is_refused_and_the_session_goes_on(gsl, ball) -> None:
    with pytest.raises(omega_core.EvalError, match="one of eval"):
        gsl.request({"id": 1, "op": "spin", "code": "@scene(40)\nq = length(p) - 3;\nreturn q;\n"})
    assert ball.eval("ball", points=[(20, 0, 0)])[0].values == pytest.approx((10,))
