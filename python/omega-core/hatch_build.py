"""Put one platform's ``gsl`` and ``compile`` into the omega-core wheel, and tag the wheel for that platform.

The programs are built first, by ``tool/build-wheel.sh`` (``compile`` with the ``omega_static`` tag, so it loads only
the system's libraries); this hook only packages them. It reads them from ``OMEGA_CORE_BIN`` (default: the build
script's ``target/static/bin``), refuses to build a wheel when either is missing or reports another version than the
package's, and copies them into ``omega_core/_bin`` with a ``manifest.json`` naming the version, the platform tag and
the omega-core commit they were built from. The licence texts go to ``omega_core/licenses``: omega-core's, and
``THIRD_PARTY_LICENSES.txt``, which the build script writes beside the programs (``tool/third_party_licenses.py``)
and without which no wheel is built.

The tag is ``OMEGA_CORE_TAG`` when set (CI sets ``manylinux_2_28_<arch>`` for programs built in that image), else this
machine's: ``macosx_11_0_<arch>`` on macOS (the build script sets that deployment target), ``linux_<arch>`` on Linux,
which PyPI refuses, so a Linux wheel for PyPI is always one CI built and named.
"""

from __future__ import annotations

import json
import os
import platform as _platform
import shutil
import stat
import subprocess
import sys
from pathlib import Path

from hatchling.builders.hooks.plugin.interface import BuildHookInterface

HERE = Path(__file__).resolve().parent
REPO = HERE.parent.parent
PROGRAMS = ("gsl", "compile")
LICENSES = {
    "LICENSE": REPO / "LICENSE",
    "THIRD_PARTY_NOTICES.md": REPO / "THIRD_PARTY_NOTICES.md",
    "MPL-2.0.txt": REPO / "internal/fidget/vendor/LICENSE",
}


def host_tag() -> str:
    machine = _platform.machine().lower()
    arch = {"arm64": "arm64", "aarch64": "aarch64", "x86_64": "x86_64", "amd64": "x86_64"}.get(machine, machine)
    if sys.platform == "darwin":
        return f"macosx_11_0_{'arm64' if arch in ('arm64', 'aarch64') else 'x86_64'}"
    if sys.platform.startswith("linux"):
        return f"linux_{'aarch64' if arch in ('arm64', 'aarch64') else arch}"
    raise RuntimeError(f"omega-core has no build for {sys.platform} yet")


def commit() -> str:
    try:
        return subprocess.run(["git", "rev-parse", "HEAD"], cwd=REPO, capture_output=True, text=True,
                              check=True).stdout.strip()
    except (OSError, subprocess.CalledProcessError):
        return "unknown"


def stage(source: Path, into: Path, version: str, tag: str) -> None:
    """The programs from ``source`` and a manifest, laid out under ``into`` as the wheel ships them."""
    shutil.rmtree(into, ignore_errors=True)
    into.mkdir(parents=True)
    for name in PROGRAMS:
        built = source / name
        if not built.is_file():
            raise RuntimeError(f"no {name} in {source}: run tool/build-wheel.sh first (or set OMEGA_CORE_BIN)")
        said = subprocess.run([str(built), "--version"], capture_output=True, text=True).stdout.strip()
        if said != f"{name} {version}":
            raise RuntimeError(f"{built} --version says {said!r}, not '{name} {version}': rebuild it from this checkout")
        target = into / name
        shutil.copyfile(built, target)
        target.chmod(target.stat().st_mode | stat.S_IRWXU | stat.S_IRGRP | stat.S_IXGRP | stat.S_IROTH | stat.S_IXOTH)
    (into / "manifest.json").write_text(json.dumps(
        {"version": version, "tag": tag, "commit": commit(), "programs": list(PROGRAMS)}, indent=2) + "\n")


class CustomBuildHook(BuildHookInterface):
    def initialize(self, version: str, build_data: dict) -> None:
        if self.target_name != "wheel":
            return
        source = Path(os.environ.get("OMEGA_CORE_BIN") or REPO / "target/static/bin").resolve()
        tag = os.environ.get("OMEGA_CORE_TAG") or host_tag()
        staged = Path(self.directory) / ".omega-core-bin"
        stage(source, staged, self.metadata.version, tag)
        build_data["force_include"][str(staged)] = "omega_core/_bin"
        third_party = source / "THIRD_PARTY_LICENSES.txt"
        if not third_party.is_file():
            raise RuntimeError(f"no {third_party}: run tool/build-wheel.sh, which writes it")
        build_data["force_include"][str(third_party)] = "omega_core/licenses/THIRD_PARTY_LICENSES.txt"
        for name, path in LICENSES.items():
            build_data["force_include"][str(path)] = f"omega_core/licenses/{name}"
        build_data["tag"] = f"py3-none-{tag}"
        build_data["pure_python"] = False
        self.app.display_info(f"omega-core {self.metadata.version}: gsl and compile from {source}, tagged {tag}")

    def finalize(self, version: str, build_data: dict, artifact_path: str) -> None:
        if self.target_name == "wheel":
            shutil.rmtree(Path(self.directory) / ".omega-core-bin", ignore_errors=True)
