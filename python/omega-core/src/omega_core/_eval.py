"""Questions asked of a GSL document's field, answered by ``gsl eval --stdio``: one program per :class:`Session`, kept
for as many questions as are asked, so a question asked again of an unchanged document is not assembled again.

    with omega_core.Session() as gsl:
        part = gsl.document("part.gsl")
        part.eval(["body", "body.dp"], points=[(0, 0, 5)])["body"].values[0]

Values are float32, as the evaluator computes them. A value that is not finite is None.
"""

from __future__ import annotations

import json
import os
import subprocess
import tempfile
from dataclasses import dataclass
from typing import Any, Dict, Iterable, Iterator, List, Mapping, Optional, Sequence, Tuple, Union

__all__ = ["Bounds", "Certificate", "Description", "Document", "EvalError", "Extremum", "Field", "Fields", "Picture",
           "Session", "SurfaceSample", "Volume"]

Vec3 = Tuple[float, float, float]
Region = Sequence[float]


class EvalError(RuntimeError):
    """A question gsl eval could not answer, in its words (a GSL error reads as the compiler's)."""


def _vec(v: Optional[Sequence[Any]]) -> Optional[Tuple[Any, ...]]:
    return None if v is None else tuple(v)


def _ranges(m: Optional[Mapping[str, Sequence[float]]]) -> Optional[Dict[str, Tuple[float, float]]]:
    return None if m is None else {name: (r[0], r[1]) for name, r in m.items()}


@dataclass(frozen=True)
class Field:
    """One expression evaluated at each point. ``values[i]`` and ``gradients[i]`` are its value and spatial gradient
    ``(dx, dy, dz)`` at point ``i``: a float and a 3-tuple for a scalar, a tuple of them per component for a vector or
    matrix (``type`` is the GSL type, ``"num"`` for a scalar)."""

    expr: Optional[str]
    type: str
    values: Tuple[Any, ...]
    gradients: Tuple[Any, ...]

    @classmethod
    def _of(cls, group: Mapping[str, Any]) -> "Field":
        values = group["values"]  # [component][point] on the wire
        gradients = group["gradients"]
        points = len(values[0]) if values else 0
        if len(values) == 1:
            by_point = tuple(values[0])
            grads = tuple(tuple(g) for g in gradients[0])
        else:
            by_point = tuple(tuple(component[i] for component in values) for i in range(points))
            grads = tuple(tuple(tuple(component[i]) for component in gradients) for i in range(points))
        return cls(group.get("expr"), group["type"], by_point, grads)


class Fields(Sequence[Field]):
    """The fields an eval asked for, in order, also found by their expression (``fields["body"]``). A call that returns
    several values gives several fields for one expression: then they are found by their place only."""

    def __init__(self, fields: Sequence[Field]):
        self._fields = tuple(fields)

    def __getitem__(self, key: Union[int, str, slice]) -> Any:
        if isinstance(key, str):
            for f in self._fields:
                if f.expr == key:
                    return f
            if any(f.expr is None for f in self._fields):
                raise KeyError(f"{key!r}: an expression returned several values, so the fields are found by place")
            raise KeyError(key)
        return self._fields[key]

    def __len__(self) -> int:
        return len(self._fields)

    def __iter__(self) -> Iterator[Field]:
        return iter(self._fields)

    def __repr__(self) -> str:
        return f"Fields({list(self._fields)!r})"


@dataclass(frozen=True)
class Extremum:
    """A global minimum and/or maximum. ``min``/``max`` are true field values at ``min_at``/``max_at``. The
    ``conditional_bound`` (lower for a minimum, upper for a maximum) holds only if ``|grad expr|`` never exceeds
    ``assumed_lipschitz`` in the region: to claim a bound, certify it."""

    expr: str
    mode: str
    assumed_lipschitz: float
    tolerance: float
    evals: int
    min: Optional[float] = None
    min_at: Optional[Vec3] = None
    min_conditional_bound: Optional[float] = None
    min_converged: Optional[bool] = None
    max: Optional[float] = None
    max_at: Optional[Vec3] = None
    max_conditional_bound: Optional[float] = None
    max_converged: Optional[bool] = None


@dataclass(frozen=True)
class Certificate:
    """A sound proof, by interval arithmetic, of ``expr >= threshold`` over the region (and every value of the
    parameters in ``params``). ``verdict`` is ``"proven"``, ``"disproven"`` (``counter_box``, and the parameter values
    where it fails for every value, ``counter_params``) or ``"indeterminate"`` (``reason``; ``worst_box``)."""

    expr: str
    verdict: str
    threshold: float
    tolerance: float
    evals: int
    region_lower: Optional[float]
    region_upper: Optional[float]
    params: Optional[Dict[str, Tuple[float, float]]] = None
    reason: Optional[str] = None
    counter_box: Optional[Tuple[float, ...]] = None
    counter_upper: Optional[float] = None
    counter_params: Optional[Dict[str, Tuple[float, float]]] = None
    worst_box: Optional[Tuple[float, ...]] = None
    worst_lower: Optional[float] = None
    worst_params: Optional[Dict[str, Tuple[float, float]]] = None

    @property
    def proven(self) -> bool:
        return self.verdict == "proven"


@dataclass(frozen=True)
class Bounds:
    """The box of the grid cells where ``expr`` is below zero; it can fall short of the surface by up to a cell
    (``cell_size``), so grow it before relying on it as an enclosure. ``min_value`` at ``min_at`` is the least sample,
    reported even when nothing was found."""

    expr: str
    found: bool
    bbox_min: Vec3
    bbox_max: Vec3
    cell_size: Vec3
    resolution: int
    min_value: Optional[float]
    min_at: Vec3


@dataclass(frozen=True)
class SurfaceSample:
    point: Vec3
    normal: Vec3
    value: Optional[float]
    converged: bool


@dataclass(frozen=True)
class Volume:
    """The volume where ``expr`` is below zero inside the region, bracketed, in cubic millimetres."""

    expr: str
    lower_mm3: float
    upper_mm3: float
    evals: int


@dataclass(frozen=True)
class Picture:
    """A PNG written at ``path``: the whole sheet's ``width`` and ``height`` in pixels (views tiled two across)."""

    path: str
    width: int
    height: int


@dataclass(frozen=True)
class Description:
    sliders: Dict[str, Tuple[float, float]]
    names: Tuple[str, ...]
    returns: Tuple[str, ...]


class Session:
    """One ``gsl eval --stdio`` program, asked questions until :meth:`close` (or the ``with`` block ends).
    ``source_paths`` are its ``-I`` roots for package resolution (default: each document's folder); ``program`` is
    the ``gsl`` to run (default: this package's)."""

    def __init__(self, source_paths: Iterable[Union[str, "os.PathLike[str]"]] = (),
                 program: Optional[Union[str, "os.PathLike[str]"]] = None):
        if program is None:
            from omega_core import executable

            program = executable()
        args = [os.fspath(program), "eval", "--stdio"]
        for path in source_paths:
            args += ["-I", os.fspath(path)]
        self._said = tempfile.TemporaryFile()  # stderr, read when the program stops; a pipe could fill and stall it
        self._process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=self._said,
                                         text=True, encoding="utf-8", bufsize=1)
        self._next = 0

    def __enter__(self) -> "Session":
        return self

    def __exit__(self, *_exc: Any) -> None:
        self.close()

    def close(self) -> None:
        if self._process.poll() is None:
            assert self._process.stdin is not None
            self._process.stdin.close()
            try:
                self._process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self._process.kill()
                self._process.wait()
        if self._process.stdout is not None:
            self._process.stdout.close()
        self._said.close()

    def document(self, path: Union[str, "os.PathLike[str]"]) -> "Document":
        """The document at ``path``, read again for each question, so an edit between questions is seen."""
        return Document(self, {"source": os.path.abspath(os.fspath(path))})

    def code(self, text: str) -> "Document":
        """A document given as its text."""
        return Document(self, {"code": text})

    def request(self, request: Mapping[str, Any]) -> Dict[str, Any]:
        """Send one request as is and return its response, raising :class:`EvalError` for one that failed."""
        self._next += 1
        ident = self._next
        line = json.dumps({**request, "id": ident}, allow_nan=False)
        stdin, stdout = self._process.stdin, self._process.stdout
        assert stdin is not None and stdout is not None
        try:
            stdin.write(line + "\n")
            stdin.flush()
            answer = stdout.readline()
        except (BrokenPipeError, OSError):
            answer = ""
        if not answer:
            raise EvalError(f"gsl eval stopped: {self._stderr() or 'it said nothing'}")
        response = json.loads(answer)
        if response.get("id") != ident:
            raise EvalError(f"gsl eval answered request {response.get('id')!r}, not {ident}")
        if not response.get("ok"):
            raise EvalError(response.get("error", "gsl eval failed"))
        return response

    def _stderr(self) -> str:
        try:
            self._process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            pass
        self._said.seek(0)
        return self._said.read().decode("utf-8", "replace").strip()


def _region(region: Region) -> List[float]:
    box = [float(v) for v in region]
    if len(box) != 6:
        raise ValueError(f"a region is [x0, y0, z0, x1, y1, z1], not {region!r}")
    return box


def _points(points: Iterable[Sequence[float]]) -> List[List[float]]:
    out = [[float(c) for c in point] for point in points]
    if any(len(point) != 3 for point in out):
        raise ValueError("each point is (x, y, z)")
    return out


class Document:
    """A GSL document in a :class:`Session`. ``defs`` on any question are GSL statements in the document's scope;
    one naming a declared name replaces it for that question (``defs="skin = 1.6;"``). ``expr`` defaults to the union
    of the document's returns. Lengths are millimetres."""

    def __init__(self, session: Session, which: Mapping[str, str]):
        self._session = session
        self._which = dict(which)

    def _ask(self, op: str, **fields: Any) -> Dict[str, Any]:
        request = {"op": op, **self._which}
        request.update({k: v for k, v in fields.items() if v is not None and v != ""})
        return self._session.request(request)

    def describe(self) -> Description:
        r = self._ask("describe")
        return Description(_ranges(r["sliders"]) or {}, tuple(r["names"]), tuple(r["returns"]))

    def eval(self, exprs: Union[str, Sequence[str]], points: Iterable[Sequence[float]], *, defs: str = "") -> Fields:
        """Each expression's value and gradient at each point (``.dp``, ``.dtime`` and ``.d(var)`` are expressions
        too). Send many points at once: each question costs a round trip."""
        exprs = [exprs] if isinstance(exprs, str) else list(exprs)
        r = self._ask("eval", exprs=exprs, points=_points(points), defs=defs)
        return Fields([Field._of(group) for group in r["groups"]])

    def extremum(self, expr: Optional[str] = None, *, region: Region, mode: str = "min", tolerance: float = 0.0,
                 lipschitz: float = 0.0, defs: str = "") -> Extremum:
        """The global min and/or max (``mode``) over the region. ``lipschitz`` 0 assumes 1, right for a distance
        field and min/max of them, not for a sum of two, a TPMS lattice or a blend that overstates distance."""
        r = self._ask("extremum", expr=expr, region=_region(region), mode=mode, tolerance=tolerance or None,
                      lipschitz=lipschitz or None, defs=defs)
        return Extremum(r["expr"], r["mode"], r["assumes"]["lipschitz"], r["tolerance"], r["evals"],
                        r.get("min"), _vec(r.get("minAt")), r.get("minLower"), r.get("minConverged"),
                        r.get("max"), _vec(r.get("maxAt")), r.get("maxUpper"), r.get("maxConverged"))

    def certify(self, expr: Optional[str] = None, *, region: Region, threshold: float, tolerance: float = 0.0,
                params: Optional[Mapping[str, Sequence[float]]] = None, over_sliders: bool = False,
                defs: str = "") -> Certificate:
        """Prove ``expr >= threshold`` over the region, and over each parameter's range in ``params`` or every
        slider's declared range (``over_sliders``)."""
        r = self._ask("certify", expr=expr, region=_region(region), threshold=float(threshold),
                      tolerance=tolerance or None, defs=defs, overSliders=over_sliders or None,
                      params={k: [float(v[0]), float(v[1])] for k, v in params.items()} if params else None)
        return Certificate(r["expr"], r["verdict"], r["threshold"], r["tolerance"], r["evals"], r.get("regionLower"),
                           r.get("regionUpper"), _ranges(r.get("params")), r.get("reason"), _vec(r.get("counterBox")),
                           r.get("counterUpper"), _ranges(r.get("counterParams")), _vec(r.get("worstBox")),
                           r.get("worstLower"), _ranges(r.get("worstParams")))

    def bounds(self, expr: Optional[str] = None, *, region: Optional[Region] = None, resolution: int = 0,
               defs: str = "") -> Bounds:
        r = self._ask("bounds", expr=expr, region=_region(region) if region is not None else None,
                      resolution=resolution or None, defs=defs)
        return Bounds(r["expr"], r["found"], _vec(r["bboxMin"]), _vec(r["bboxMax"]), _vec(r["cellSize"]),
                      r["resolution"], r.get("minValue"), _vec(r["minAt"]))  # type: ignore[arg-type]

    def surface(self, points: Iterable[Sequence[float]], expr: Optional[str] = None, *, max_iters: int = 0,
                tolerance: float = 0.0, defs: str = "") -> Tuple[SurfaceSample, ...]:
        """Each point moved onto ``expr = 0`` along the gradient, with the surface normal there."""
        r = self._ask("surface", expr=expr, points=_points(points), maxIters=max_iters or None,
                      tolerance=tolerance or None, defs=defs)
        return tuple(SurfaceSample(_vec(s["point"]), _vec(s["normal"]), s.get("value"), s["converged"])  # type: ignore[arg-type]
                     for s in r["samples"])

    def render(self, out: Union[str, "os.PathLike[str]"], expr: Optional[str] = None, *,
               views: Union[str, Sequence[str]] = "iso", width: int = 640, height: int = 480, light: bool = False,
               extent: float = 0.0, time: float = 0.0, defs: str = "") -> Picture:
        """Draw ``expr`` -- or the document's returns, each in its own colour -- into the PNG ``out``, as ``gsl compile
        -t png`` draws it: from ``views`` (presets such as ``"iso"``, ``"top"``, or ``"x,y,z"`` directions; several
        tile two across), each ``width`` by ``height``, on the light ground if ``light``, over a scene of half-size
        ``extent`` mm (0 frames the geometry) at animation ``time``. ``defs`` set sliders for this picture alone, and
        ``expr`` can cut it open (``"max(body, p.y)"``)."""
        cameras = views if isinstance(views, str) else ",".join(views)
        r = self._ask("render", out=os.path.abspath(os.fspath(out)), expr=expr, views=cameras, width=int(width),
                      height=int(height), light=light or None, extent=extent or None, time=time or None, defs=defs)
        return Picture(r["out"], r["width"], r["height"])

    def volume(self, expr: Optional[str] = None, *, region: Region, relative_tolerance: float = 0.0) -> Volume:
        """The volume inside the region, bracketed (``relative_tolerance`` 0.05 is quick; 0.01 takes seconds)."""
        r = self._ask("volume", expr=expr, region=_region(region), relativeTolerance=relative_tolerance or None)
        return Volume(r["expr"], r["lowerMm3"], r["upperMm3"], r["evals"])
