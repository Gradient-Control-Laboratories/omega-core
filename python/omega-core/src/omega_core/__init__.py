"""omega-core's GSL toolchain, packaged as a wheel: the ``gsl`` command, which compiles, meshes, renders and
queries implicit (signed-distance) models written in GSL, and a client for querying a model's field from Python.

The programs are built for one platform per wheel and load only the system's libraries. The ``gsl`` command this
package installs runs the bundled program with its arguments as given.

Querying a model (:class:`Session`, :class:`Document`; ``help(omega_core.Document)`` lists every question and what
it returns)::

    import omega_core

    with omega_core.Session() as gsl:
        part = gsl.document("part.gsl")
        part.eval(["body", "body.d(skin)"], points=[(40, 0, 9)], defs="skin = 1.6;")
        part.certify("body", region=(-60, -60, 0, 60, 60, 50), threshold=0, over_sliders=True).verdict
        part.render("cut.png", "max(body, p.y)", views="iso,top", light=True)

Finding the program: :func:`executable` is the path to ``gsl``, and :func:`version` the release (``"0.1.1"``).
"""

from __future__ import annotations

import sys
from pathlib import Path

from omega_core._eval import Document, EvalError, Session

__all__ = ["Document", "EvalError", "Session", "executable", "version"]

_BIN = Path(__file__).resolve().parent / "_bin"
_PROGRAMS = ("gsl", "compile")


def executable(name: str = "gsl") -> Path:
    """The bundled program ``name`` (``gsl`` or ``compile``). ``FileNotFoundError`` for an install that has none (a
    source checkout, or a wheel for another platform)."""
    if name not in _PROGRAMS:
        raise ValueError(f"omega-core ships {' and '.join(_PROGRAMS)}, not {name!r}")
    path = _BIN / (name + ".exe" if sys.platform == "win32" else name)
    if not path.is_file():
        raise FileNotFoundError(f"this install of the omega-core package has no {name} program at {path}: install a "
                                "wheel built for this platform")
    return path


def version() -> str:
    """The omega-core release bundled ("0.1.1"), without running anything."""
    from importlib.metadata import version as installed

    return installed("omega-core")
