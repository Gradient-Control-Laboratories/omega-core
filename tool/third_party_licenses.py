#!/usr/bin/env python3
"""Write the licence texts of everything linked into the omega-gsl programs (gsl and compile).

THIRD_PARTY_NOTICES.md says what a binary distribution of omega's render library must carry: the MPL-2.0 terms,
and the MIT, BSD and ISC notices of what it links. This collects them from the sources the programs were built
from, never from a list kept by hand, so a dependency added or bumped is in the next wheel's file:

- Go: the standard library (GOROOT's LICENSE) and every module ``go list -deps`` says the two programs link.
- Rust: the standard library, and every crate ``cargo metadata`` resolves as a normal (linked) dependency of
  ``render`` for this machine's target -- each crate's own LICENSE / COPYING / NOTICE files, a ``LICENSES/``
  folder's, and epaint's font licences (``fonts/*.txt``). Build and dev dependencies are not linked and are left out.
- C/C++: the libraries render.a bundles (internal/fidget/vendor: OpenVDB, oneTBB, Imath, c-blosc, zlib).

A vendored crate's text is its folder's parent's (internal/fidget/vendor/LICENSE, MPL-2.0). A crate published with
no licence file at all gets the standard text of a licence its SPDX expression offers -- Apache-2.0 when it does,
else MIT with the crate's authors as the copyright holders, else MPL-2.0 -- from the Rust toolchain's SPDX texts.
One that offers none of these stops the build: a wheel never ships a component whose terms it does not carry.

Identical texts are written once, each under the list of components it covers. Standard library only.

usage: tool/third_party_licenses.py OUT.txt
"""

from __future__ import annotations

import hashlib
import json
import re
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
PROGRAMS = ("./cmd/compile", "./cmd/gsl")
NATIVE = ("openvdb", "tbb", "imath", "blosc", "zlib")  # what render's build.rs bundles into render.a
LICENSE_FILE = re.compile(r"^(LICEN[SC]E|COPYING|NOTICE|UNLICENSE|COPYRIGHT|third-party-programs)", re.IGNORECASE)
CHOICES = ("Apache-2.0", "MIT", "MPL-2.0")  # in order of preference, for a crate that ships no licence file


def run(*args: str) -> str:
    return subprocess.run(args, cwd=ROOT, capture_output=True, text=True, check=True).stdout


def licence_files(folder: Path) -> list[Path]:
    found = sorted(p for p in folder.iterdir() if p.is_file() and LICENSE_FILE.match(p.name))
    if (folder / "LICENSES").is_dir():
        found += sorted(p for p in (folder / "LICENSES").iterdir() if p.is_file())
    return found


class Texts:
    """Licence texts, each once, with the components it covers."""

    def __init__(self) -> None:
        self.texts: dict[str, tuple[str, str, list[str]]] = {}  # digest -> (title, text, components)
        self.components: list[tuple[str, str, str, list[str]]] = []  # (section, component, licence, digests)

    def add(self, section: str, component: str, licence: str, texts: list[tuple[str, str]]) -> None:
        digests = []
        for title, text in texts:
            text = text.replace("\r\n", "\n").strip() + "\n"
            digest = hashlib.sha256(text.encode()).hexdigest()[:12]
            self.texts.setdefault(digest, (title, text, []))[2].append(component)
            digests.append(digest)
        self.components.append((section, component, licence, digests))

    def write(self, out: Path) -> None:
        number = {digest: i + 1 for i, digest in enumerate(self.texts)}
        lines = ["Third-party licences of the omega-gsl programs (gsl and compile)", "",
                 "Everything below is linked into the two programs this wheel ships. Each component names the",
                 "texts that are its terms; each text follows once, with the components it covers. Written by",
                 "omega-core's tool/third_party_licenses.py from the sources the programs were built from. The",
                 "MPL-2.0 components' source is omega-core's repository (internal/fidget/vendor) and crates.io.", ""]
        section = None
        for this, component, licence, digests in self.components:
            if this != section:
                lines += ["", f"== {this} ==", ""]
                section = this
            lines.append(f"{component}  ({licence})  texts {', '.join(f'#{number[d]}' for d in digests)}")
        for digest, (title, text, covers) in self.texts.items():
            lines += ["", "=" * 100, f"#{number[digest]}  {title}", f"for: {', '.join(covers)}", "=" * 100, "", text]
        out.parent.mkdir(parents=True, exist_ok=True)
        out.write_text("\n".join(lines))


def standard_texts() -> dict[str, str]:
    """The Rust toolchain's SPDX licence texts (share/doc/rust/licenses), by SPDX id."""
    folder = Path(run("rustc", "--print", "sysroot").strip()) / "share/doc/rust/licenses"
    return {p.stem: p.read_text(errors="replace") for p in folder.glob("*.txt")} if folder.is_dir() else {}


def chosen(expression: str | None, authors: list[str], name: str, standard: dict[str, str],
           mpl: str) -> tuple[str, str]:
    """The text of the licence a crate with no licence file is taken under."""
    offered = set(re.findall(r"[A-Za-z0-9.+-]+", expression or "")) - {"OR", "AND", "WITH"}
    for spdx in CHOICES:
        if spdx not in offered:
            continue
        if spdx == "MPL-2.0":
            return spdx, mpl
        if spdx not in standard:
            raise SystemExit(f"third_party_licenses: the Rust toolchain has no {spdx} text for {name}")
        text = standard[spdx]
        if spdx == "MIT":
            holders = ", ".join(re.sub(r"\s*<[^>]*>", "", a) for a in authors) or f"the {name} authors"
            text = text.replace("<year> <copyright holders>", holders)
        return spdx, text
    raise SystemExit(f"third_party_licenses: {name} ships no licence file and its licence ({expression}) offers none "
                     f"of {', '.join(CHOICES)}: add its text by hand")


def go(texts: Texts) -> None:
    goroot = Path(run("go", "env", "GOROOT").strip())
    version = run("go", "env", "GOVERSION").strip()
    texts.add("Go", f"Go standard library {version}", "BSD-3-Clause",
              [(f"Go {version}: {p.name}", p.read_text()) for p in licence_files(goroot) if p.name != "README.md"])
    modules = {}
    for line in run("go", "list", "-deps", "-f", "{{with .Module}}{{.Path}}\t{{.Version}}\t{{.Dir}}{{end}}",
                    *PROGRAMS).splitlines():
        path, version, folder = line.split("\t")
        if path != "gradientcontrol.com/omega/core":
            modules[path] = (version, Path(folder))
    for path, (version, folder) in sorted(modules.items()):
        files = licence_files(folder)
        if not files:
            raise SystemExit(f"third_party_licenses: Go module {path} {version} has no licence file in {folder}")
        texts.add("Go modules", f"{path} {version}", "see texts", [(f"{path}: {p.name}", p.read_text()) for p in files])


def rust(texts: Texts) -> None:
    host = re.search(r"^host: (\S+)", run("rustc", "-vV"), re.MULTILINE).group(1)
    meta = json.loads(run("cargo", "metadata", "--format-version", "1", "--locked", "--offline",
                          "--filter-platform", host))
    packages = {p["id"]: p for p in meta["packages"]}
    nodes = {n["id"]: n for n in meta["resolve"]["nodes"]}
    members = set(meta["workspace_members"])
    vendor = ROOT / "internal/fidget/vendor"
    mpl = (vendor / "LICENSE").read_text()
    standard = standard_texts()
    root = next(i for i, p in packages.items() if p["name"] == "render" and i in members)
    linked, todo = set(), [root]
    while todo:
        here = todo.pop()
        if here in linked:
            continue
        linked.add(here)
        todo += [d["pkg"] for d in nodes[here]["deps"] if any(k["kind"] is None for k in d["dep_kinds"])]
    sysroot_version = run("rustc", "--version").split()[1]
    if "Apache-2.0" not in standard:
        raise SystemExit("third_party_licenses: the Rust toolchain has no SPDX texts (share/doc/rust/licenses)")
    texts.add("Rust", f"Rust standard library {sysroot_version}", "MIT OR Apache-2.0; taken under Apache-2.0",
              [("Apache-2.0 (standard text)", standard["Apache-2.0"])])
    for ident in sorted(linked, key=lambda i: (packages[i]["name"], packages[i]["version"])):
        package = packages[ident]
        folder = Path(package["manifest_path"]).parent
        first_party = ident in members and vendor not in folder.parents
        if first_party:
            continue  # omega-core's own crates: its MIT licence, shipped as LICENSE
        name = f"{package['name']} {package['version']}"
        licence = package["license"] or "see texts"
        files = licence_files(folder)
        if files:
            found = [(f"{name}: {p.relative_to(folder)}", p.read_text(errors="replace")) for p in files]
        elif vendor in folder.parents:
            found = [("MPL-2.0 (standard text)", mpl)]
        else:
            spdx, text = chosen(package["license"], package["authors"], package["name"], standard, mpl)
            found = [(f"{spdx} (standard text)" if spdx != "MIT" else f"MIT ({name})", text)]
            licence += f"; ships no licence file: taken under {spdx}"
        if package["name"] == "epaint":  # egui's default fonts, compiled in
            fonts = sorted((folder / "fonts").glob("*.txt"))
            if not fonts:
                raise SystemExit("third_party_licenses: epaint's font licences (fonts/*.txt) are missing")
            found += [(f"{name}: {p.relative_to(folder)}", p.read_text(errors="replace")) for p in fonts]
        texts.add("Rust crates", name, licence, found)


def native(texts: Texts) -> None:
    for name in NATIVE:
        folder = ROOT / "internal/fidget/vendor" / name
        revision = run("git", "-C", str(folder), "describe", "--tags", "--always").strip()
        files = licence_files(folder)
        if not files:
            raise SystemExit(f"third_party_licenses: no licence file in {folder}: are the submodules checked out?")
        texts.add("C and C++ libraries in render", f"{name} {revision}", "see texts",
                  [(f"{name}: {p.relative_to(folder)}", p.read_text(errors="replace")) for p in files])


def main() -> None:
    if len(sys.argv) != 2:
        raise SystemExit(__doc__.strip().splitlines()[-1])
    texts = Texts()
    go(texts)
    rust(texts)
    native(texts)
    out = Path(sys.argv[1])
    texts.write(out)
    print(f"third_party_licenses: {len(texts.components)} components, {len(texts.texts)} distinct texts, "
          f"{out.stat().st_size // 1024} KB -> {out}")


if __name__ == "__main__":
    main()
