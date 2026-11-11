"""The installed omega-core wheel: its programs are there, report the package's version, load only the system's
libraries, mesh a model right and draw it. Standard library and pytest only, so it runs against a bare install."""

from __future__ import annotations

import shutil
import struct
import subprocess
import sys
import zlib
from collections import Counter
from pathlib import Path

import pytest

import omega_core

BOX = "@scene(40, 5)\nblk = box([0, 0, 0], [10, 20, 30]);\nreturn blk;\n"


def run(*args: str, cwd: Path | None = None) -> str:
    done = subprocess.run(list(args), capture_output=True, text=True, cwd=cwd, timeout=300)
    assert done.returncode == 0, done.stdout + done.stderr
    return done.stdout.strip()


@pytest.mark.parametrize("name", ["gsl", "compile"])
def test_each_program_is_there_and_reports_the_package_version(name: str) -> None:
    path = omega_core.executable(name)
    assert path.is_file()
    assert run(str(path), "--version") == f"{name} {omega_core.version()}"


def test_the_gsl_command_runs_the_bundled_program() -> None:
    assert run(sys.executable, "-m", "omega_core", "--version") == f"gsl {omega_core.version()}"


def test_the_licences_of_what_it_links_are_shipped() -> None:
    licences = Path(omega_core.__file__).parent / "licenses"
    third_party = (licences / "THIRD_PARTY_LICENSES.txt").read_text()
    for component in ("Go standard library", "Rust standard library", "fidget-core", "epaint", "openvdb", "tbb",
                      "imath", "blosc", "zlib"):
        assert component in third_party, component
    assert (licences / "LICENSE").is_file()


def test_another_name_is_refused() -> None:
    with pytest.raises(ValueError):
        omega_core.executable("serve")


def test_compile_loads_only_the_system_libraries() -> None:
    compile_ = str(omega_core.executable("compile"))
    if sys.platform == "darwin":
        libraries = run("otool", "-L", compile_).splitlines()[1:]
        foreign = [line.strip() for line in libraries if not line.strip().startswith(("/usr/lib/", "/System/"))]
    elif shutil.which("ldd"):
        libraries = run("ldd", compile_).splitlines()
        foreign = [line.strip() for line in libraries if "not found" in line or "render" in line]
    else:
        pytest.skip("no otool or ldd here")
    assert not foreign, foreign


def read_stl(path: Path) -> list[tuple[tuple[float, float, float], ...]]:
    data = path.read_bytes()
    (count,) = struct.unpack_from("<I", data, 80)
    assert len(data) == 84 + 50 * count, "a binary STL"
    return [tuple(struct.unpack_from("<3f", data, 84 + 50 * i + 12 + 12 * k) for k in range(3)) for i in range(count)]


def test_a_sharp_box_meshes_watertight_to_its_volume(tmp_path: Path) -> None:
    (tmp_path / "box.gsl").write_text(BOX)
    run(str(omega_core.executable()), "compile", "-t", "stl", "--algo", "mdc", "--subdivision", "8", "--bbox", "auto",
        "--no-meta", "-o", "box.stl", "box.gsl", cwd=tmp_path)
    triangles = read_stl(tmp_path / "box.stl")
    assert triangles
    volume = sum(a[0] * (b[1] * c[2] - b[2] * c[1]) - a[1] * (b[0] * c[2] - b[2] * c[0])
                 + a[2] * (b[0] * c[1] - b[1] * c[0]) for a, b, c in triangles) / 6
    assert volume == pytest.approx(6000, rel=1e-4)
    # Watertight: every edge is shared by exactly two triangles.
    key = lambda point: tuple(round(value, 4) for value in point)  # noqa: E731
    edges = Counter(tuple(sorted((key(p), key(q)))) for t in triangles for p, q in zip(t, t[1:] + t[:1]))
    assert set(edges.values()) == {2}


def test_marching_cubes_meshes_a_box_facing_outward(tmp_path: Path) -> None:
    # Every triangle faces out of the solid: the signed volume is the box's, positive, as dual contouring's is.
    (tmp_path / "box.gsl").write_text(BOX)
    run(str(omega_core.executable()), "compile", "-t", "stl", "--algo", "marching-cubes", "--subdivision", "7",
        "--bbox", "auto", "--no-meta", "-o", "box.stl", "box.gsl", cwd=tmp_path)
    triangles = read_stl(tmp_path / "box.stl")
    volume = sum(a[0] * (b[1] * c[2] - b[2] * c[1]) - a[1] * (b[0] * c[2] - b[2] * c[0])
                 + a[2] * (b[0] * c[1] - b[1] * c[0]) for a, b, c in triangles) / 6
    assert volume == pytest.approx(6000, rel=0.02)


def read_png(path: Path) -> tuple[int, int, list[bytes]]:
    """An 8-bit RGB PNG's width, height and rows of pixels, its filters undone."""
    data = path.read_bytes()
    assert data[:8] == b"\x89PNG\r\n\x1a\n", "a PNG"
    at, idat = 8, b""
    while at < len(data):
        (length,) = struct.unpack_from(">I", data, at)
        kind, body = data[at + 4:at + 8], data[at + 8:at + 8 + length]
        if kind == b"IHDR":
            width, height, depth, colour = struct.unpack_from(">IIBB", body)
            assert (depth, colour) == (8, 2), "8-bit RGB"
        elif kind == b"IDAT":
            idat += body
        at += 12 + length
    raw, stride, rows, previous = zlib.decompress(idat), 3 * width, [], bytes(3 * width)
    for y in range(height):
        kind, line = raw[y * (stride + 1)], bytearray(raw[y * (stride + 1) + 1:(y + 1) * (stride + 1)])
        for x in range(stride):
            left = line[x - 3] if x >= 3 else 0
            up, corner = previous[x], previous[x - 3] if x >= 3 else 0
            if kind == 1:
                line[x] = (line[x] + left) & 255
            elif kind == 2:
                line[x] = (line[x] + up) & 255
            elif kind == 3:
                line[x] = (line[x] + (left + up) // 2) & 255
            elif kind == 4:
                p = left + up - corner
                guess = min((abs(p - left), 0, left), (abs(p - up), 1, up), (abs(p - corner), 2, corner))[2]
                line[x] = (line[x] + guess) & 255
        rows.append(previous := bytes(line))
    return width, height, rows


def test_a_model_is_drawn_from_each_view_tiled_two_across(tmp_path: Path) -> None:
    (tmp_path / "box.gsl").write_text(BOX)
    run(str(omega_core.executable()), "compile", "-t", "png", "--views", "iso,top", "--size", "160x120", "--light",
        "--no-meta", "-o", "box.png", "box.gsl", cwd=tmp_path)
    width, height, rows = read_png(tmp_path / "box.png")
    assert (width, height) == (2 * 160 + 4, 120)  # two views and omega's 4-pixel gap
    # The box is drawn, its faces shaded: an empty scene's iso view holds the ground and the axes, 4 colours.
    colours = {rows[y][3 * x:3 * x + 3] for y in range(height) for x in range(160)}
    assert len(colours) >= 20, len(colours)
