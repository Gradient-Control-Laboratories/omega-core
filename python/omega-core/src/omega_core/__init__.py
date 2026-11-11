"""omega-core's ``gsl`` program, packaged as a wheel: ``omega-core``, what text-to-cad's ``cad[omega]`` installs.

This package ships omega-core's ``gsl`` command and the ``compile`` program it runs for ``gsl compile`` (GSL
implicit models: parse, check, mesh to STL), built for one platform per wheel and linked so they load only the
system's libraries. What it adds is a way to find them from Python::

    import omega_core

    omega_core.executable()              # Path to gsl
    omega_core.executable("compile")     # Path to compile, which `gsl compile` runs
    omega_core.version()                 # "0.1.0": what `gsl --version` reports

and a way to ask a document's field questions, through ``gsl eval`` (:class:`Session`)::

    with omega_core.Session() as gsl:
        part = gsl.document("part.gsl")
        part.eval(["body", "body.d(skin)"], points=[(40, 0, 9)], defs="skin = 1.6;")
        part.certify("body", region=[...], threshold=0, over_sliders=True).verdict

The ``gsl`` command this package installs runs the bundled program with its arguments as given.
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
    """The omega-core release bundled ("0.1.0"), without running anything."""
    from importlib.metadata import version as installed

    return installed("omega-core")
