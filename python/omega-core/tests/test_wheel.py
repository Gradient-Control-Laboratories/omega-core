"""The installed omega-core wheel: its programs are there, report the package's version, load only the system's
libraries, and mesh a model right. Standard library and pytest only, so it runs against a bare install."""

from __future__ import annotations

import shutil
import struct
import subprocess
import sys
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
