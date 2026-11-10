# Third-party notices

The code and data authored for this repository are licensed under the MIT
License (see [LICENSE](LICENSE)). The components below are third-party, keep
their own licenses, and are not relicensed by the repository's MIT license.
Full license texts that are not already shipped beside the component are in the
appendices at the end of this file.

1. [Source included in this repository](#1-source-included-in-this-repository)
2. [Code and data adapted into first-party files](#2-code-and-data-adapted-into-first-party-files)
3. [Go modules](#3-go-modules)
4. [Rust crates](#4-rust-crates)
5. [Linked into built binaries, not included in this repository](#5-linked-into-built-binaries-not-included-in-this-repository)
6. [Appendices: license texts](#appendices)

## 1. Source included in this repository

### Fidget — MPL-2.0

`internal/fidget/vendor/` (11 crates), from
<https://github.com/mkeeter/fidget> by Matt Keeter, vendored at upstream commit
`e329a898` **with local modifications** (for example the `MeshOracle` op in
`fidget-core`, `fidget-jit` panicking on `MeshOracle`, and wasm `getrandom`
selection). Every file in these crates, including the modified ones, remains
under MPL-2.0. The license text is `internal/fidget/vendor/LICENSE`.
The MPL is file-level: it does not apply to this repository's own files.

### three.js r178 — MIT

`examples/viewer/vendor/three/`. Copyright 2010-2025 three.js authors. License
text: `examples/viewer/vendor/three/LICENSE`. The kept files are
byte-identical to the upstream r178 tag.

Third-party code embedded in `three.core.js` / `three.webgpu.js`:

- **earcut** v3.0.1, ISC, Copyright (c) 2024 Mapbox (Appendix B).
- **cyrb53** hash, Copyright (c) 2018 bryc, public domain.

Only the files the viewer loads are kept: `three.core.js`, `three.webgpu.js`,
`three.tsl.js`, and under `jsm/` OrbitControls, the `lines/` Line2 family,
`OBJLoader`, `STLLoader` and `libs/fflate.module.js`. fflate is MIT, Copyright
(c) 2023 Arjun Barrett.

### CodeMirror 6 and Lezer — MIT

`examples/viewer/vendor/codemirror/codemirror.js`, built from `package.json`
and `package-lock.json`. Copyright Marijn Haverbeke and others. The bundle
carries no notices of its own, so the full notices are in Appendix A.

### Manifold 3.4.x — Apache-2.0

`examples/viewer/vendor/manifold/` (`manifold.js`, `manifold.wasm`), from
<https://github.com/elalish/manifold>. Copyright The Manifold Authors. License
text: `examples/viewer/vendor/manifold/LICENSE`. Manifold has no NOTICE file.
The WebAssembly build also contains:

- **Clipper2**, Boost Software License 1.0, Copyright Angus Johnson 2010-2025
  (Appendix C).
- **quickhull**, derived from the public-domain work of Antti Kuukka.
- **linalg.h**, public domain, by Sterling Orsten.
- **libc++ / libc++abi** from the toolchain (Apache-2.0 with LLVM exception)
  and the Emscripten runtime glue (MIT).

### Fonts

| Font | Files | License | Notes |
|---|---|---|---|
| IBM Plex Mono | `examples/viewer/fonts/IBMPlexMono-*.ttf` | SIL OFL 1.1 | Copyright 2017 IBM Corp., Reserved Font Name "Plex". Unmodified v2.005 from <https://github.com/IBM/plex>. License: `examples/viewer/fonts/IBMPlexMono-OFL.txt`. |
| Public Sans | `examples/viewer/fonts/PublicSans-*.otf` | SIL OFL 1.1 | Copyright 2015 The Public Sans Project Authors (<https://github.com/uswds/public-sans>). Based on Libre Franklin, Copyright (c) 2015 Impallari Type. The U.S. Web Design System modifications are CC0. License: `examples/viewer/fonts/PublicSans-OFL.txt`. |

### Icons

- `examples/viewer/icons/code.svg` is the Heroicons `code-bracket` outline icon,
  MIT, Copyright (c) Tailwind Labs, Inc. (Appendix E).
- `examples/viewer/icons/eyedropper.svg` is Lucide's `pipette` icon, ISC,
  Copyright (c) Lucide Icons and Contributors; Lucide also carries a Feather
  (MIT) portion (Appendix E).

### QCAD "Unicode" CXF stroke font — public domain

`stdlib/text/_tools/Unicode.cxf` is byte-identical to
<https://github.com/qcad/qcad/blob/master/fonts/Unicode.cxf> (SHA-256
`07606da523689ddb7ea11a1a83b7d426aa905981bb00e45caaebb72504c7d26d`). Latin
glyphs are by Andrew Mustun, Cyrillic by Eugene Osintsev, and Japanese (Kana,
Kanji, full-width forms; U+3041 and above) by Yoshimune Kobayashi. The glyphs in
`stdlib/text/text.gsl` (U+0021 to U+3015) are generated from it.

QCAD's `LICENSE.txt` (<https://github.com/qcad/qcad/blob/master/LICENSE.txt>)
states: "Other CXF fonts in directory 'fonts' are released as public domain
(all copyright is waived)." Hershey, osifont and VL-Gothic are excepted, and
this file is not one of them.

Provenance of the Japanese glyphs: Kobayashi's 2002 notes say the Japanese
font was generated from the bitmap font of the Kochi (東風) font family. That
most likely means the Shinonome 16-dot font by Yasuyuki Furukawa, which is
public domain (the efont project). No separate license from Mustun, Osintsev or
Kobayashi has been located, so the public-domain status rests on QCAD's
statement above. Provided as is, without warranty.

### Materials database — CC0 1.0

`stdlib/material/bulk/physical/physical.gsl` is generated by `tool/physicallybased` from the
[physicallybased.info](https://physicallybased.info) materials database
(snapshot 2026-09-01), released under CC0 1.0.

## 2. Code and data adapted into first-party files

- **Inigo Quilez** — signed-distance primitives, Bezier distance, Voronoi-edge
  distance, ambient occlusion, capsule intersection and cosine-palette code in
  `stdlib/` (`builtin/2d.gsl`, `builtin/3d.gsl`,
  `builtin/booleans.gsl`, `builtin/sweep.gsl`, `builtin/quad_bezier_curves.gsl`,
  `noise/noise.gsl`, `material/bulk/wood/wood.gsl`), the viewer renderer and the
  Fidget mesher are adapted from his articles and Shadertoy shaders
  (<https://iquilezles.org/articles/distfunctions/>, `/distfunctions2d/`,
  `/voronoilines/`, `/intersectors/`). His site releases code that implements an
  algorithm or technique under the MIT License. Copyright (c) Inigo Quilez
  (Appendix D, MIT text).
- **Hash without Sine** (`hash22` / `hash23` by David Hoskins,
  <https://www.shadertoy.com/view/4djSRW>) in `examples/viewer/render/gsl-shading.js`
  and `internal/fidget/internal/mesh/litglass.go`. MIT License,
  Copyright (c) 2014 David Hoskins (Appendix I).
- **Marching-cubes tables** (`EDGE_TABLE`, `TRI_TABLE` in
  `internal/fidget/internal/rust/src/marching_cubes.rs`) are Cory
  Bloyd's public-domain marching cubes tables, via Paul Bourke, "Polygonising
  a scalar field" (1994).
- **Closest point on a triangle** in `mesh_oracle.rs` follows Christer
  Ericson, *Real-Time Collision Detection*, Morgan Kaufmann Publishers,
  (c) 2005 Elsevier Inc.
- **NumPy random** — `program/noise/spinodoid/ziggurat_tables.go`,
  `program/noise/spinodoid/numpy.go` and their Rust ports
  `internal/fidget/internal/tape-core/src/spinodoid/ziggurat_tables.rs`
  and `.../spinodoid/numpy.rs` reproduce NumPy v2.2.6's ziggurat
  constants and its SeedSequence, PCG64 and ziggurat-normal logic. NumPy's
  `random` code is dual-licensed NCSA / BSD-3-Clause, Copyright (c) 2019 Kevin
  Sheppard; its PCG64 is MIT, Copyright 2014 Melissa O'Neill (Appendix G).
- **Khronos PBR Neutral tone mapper** (`litrender.go`, `gsl-shading.js`)
  follows the reference implementation
  `PBR_Neutral/pbrNeutral.glsl` in <https://github.com/KhronosGroup/ToneMapping>,
  Copyright 2024 The Khronos Group, Inc., licensed under Apache-2.0 (text in
  Appendix F). The upstream repository ships no NOTICE file; its `.reuse/dep5`
  assigns the copyright and the Apache-2.0 license to that file.
- **EnvBRDFApprox** fit constants are from Brian Karis, "Physically Based
  Shading on Mobile" (Epic Games, 2014).
- **Spinodoid Gabor model** (`program/noise/spinodoid/`) implements the model
  of Kumar, Tan, Zheng and Kochmann, "Inverse-designed spinodoid metamaterials",
  npj Computational Materials 6:73 (2020).
- **Colour-map data** in `color.gsl` is sampled from the viridis, magma,
  inferno and plasma maps (CC0) and Google's Turbo map (Apache-2.0).

## 3. Go modules

| Module | Version | License |
|---|---|---|
| `github.com/alecthomas/participle/v2` | v2.1.4 | MIT |
| `github.com/google/go-cmp` | v0.7.0 | BSD-3-Clause |
| `google.golang.org/protobuf` | v1.36.11 | BSD-3-Clause |
| `gopkg.in/yaml.v3` | v3.0.1 | MIT and Apache-2.0 (see the module's `LICENSE` and `NOTICE`) |

## 4. Rust crates

The crates under `internal/fidget/internal/` depend on 303 crates.io
packages across their dependency graphs. Every one offers a permissive option
(most are `MIT OR Apache-2.0`) except the following.

- **MPL-2.0 only.** The Fidget crates (from crates.io at 0.5.1 in `render` and
  `tape-py`, vendored elsewhere) and `dynasm` / `dynasmrt` 5.1.0 (reached
  through `fidget-jit`). The MPL-2.0 text is `internal/fidget/vendor/LICENSE`.
  A binary that includes them must be accompanied by the source of those
  crates, which is available from crates.io and, for the vendored copies, this
  repository.
- **Fonts embedded by `epaint` 0.27 (egui `default_fonts`)** are compiled into
  the `render` library and the `omega_tape` Python extension:
  - Ubuntu-Light, Ubuntu Font Licence 1.0 (Appendix H).
  - NotoEmoji-Regular, SIL OFL 1.1.
  - Hack-Regular, MIT with the Bitstream Vera License and DejaVu public-domain
    additions.
  - emoji-icon-font, MIT, Copyright (c) John Slegers.
- **Apache-2.0 only:** `nalgebra`, `nalgebra-macros`, `simba`, `approx`, `prost`,
  `prost-derive`, `prost-build`, `prost-types`, `ab_glyph`, `ab_glyph_rasterizer`,
  `owned_ttf_parser`, `codespan-reporting`, `spirv`, `glutin_wgl_sys`,
  `gl_generator`, `khronos_api` and `target-lexicon` (Apache-2.0 with LLVM
  exception). The text is in Appendix F. None of these ships a NOTICE file.
- **Unicode data:** `unicode-ident` is `(MIT OR Apache-2.0) AND Unicode-3.0`;
  `regex-syntax` and `unicode-width` embed Unicode data under MIT OR
  Apache-2.0. The Unicode License v3 is <https://www.unicode.org/license.txt>.

Binary distributions of `render` and `omega_tape` must also include the MIT,
BSD and ISC copyright notices of the crates they link.

## 5. Linked into built binaries, not included in this repository

The `render` crate (`librender`, and the `omega_tape` Python extension that
embeds it) is built against the following, which `build.rs` expects as git
submodules under `internal/fidget/vendor/`:

| Component | License |
|---|---|
| [OpenVDB](https://github.com/AcademySoftwareFoundation/openvdb) | Apache-2.0 |
| [oneTBB](https://github.com/uxlfoundation/oneTBB) | Apache-2.0 |
| [Imath](https://github.com/AcademySoftwareFoundation/Imath) | BSD-3-Clause |
| [c-blosc](https://github.com/Blosc/c-blosc) | BSD-3-Clause |
| [zlib](https://zlib.net) | zlib |

The `backend/mesh/manifold` Go package links `manifold_glue`, which wraps
the [Manifold](https://github.com/elalish/manifold) C++ library (Apache-2.0).

Anyone distributing those binaries must make the corresponding MPL-2.0 source
(Fidget, dynasmrt) available and ship the notices above.

---

# Appendices

## Appendix A — CodeMirror 6 and Lezer (MIT)

```text
CodeMirror 6 bundle (examples/viewer/vendor/codemirror/codemirror.js) -- MIT

MIT License

Copyright (C) 2018-2021 by Marijn Haverbeke <marijn@haverbeke.berlin> and others

  Applies to: @codemirror/autocomplete 6.20.1; @codemirror/commands 6.10.3; @codemirror/language 6.12.3; @codemirror/lint 6.9.5; @codemirror/search 6.6.0; @codemirror/state 6.6.0; @codemirror/view 6.41.1

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


-----

MIT License

Copyright (C) 2018 by Marijn Haverbeke <marijn@haverbeke.berlin> and others

  Applies to: @lezer/common 1.5.2; @lezer/highlight 1.2.3; @lezer/lr 1.4.10 (in lock; not detected in bundle); style-mod 4.1.3

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


-----

MIT License

Copyright (C) 2016 by Marijn Haverbeke <marijn@haverbeke.berlin> and others

  Applies to: w3c-keyname 2.2.8

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


-----

MIT License

Copyright (C) 2020 by Marijn Haverbeke <marijn@haverbeke.berlin>

  Applies to: crelt 1.0.6

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


-----

MIT License

Copyright (C) 2024 by Marijn Haverbeke <marijn@haverbeke.berlin>

  Applies to: @marijn/find-cluster-break 1.0.2

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.


-----

```

## Appendix B — earcut (ISC)

```text
ISC License

Copyright (c) 2024, Mapbox

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.

```

## Appendix C — Clipper2 (Boost Software License 1.0)

```text
Boost Software License - Version 1.0 - August 17th, 2003

Permission is hereby granted, free of charge, to any person or organization
obtaining a copy of the software and accompanying documentation covered by
this license (the "Software") to use, reproduce, display, distribute,
execute, and transmit the Software, and to prepare derivative works of the
Software, and to permit third-parties to whom the Software is furnished to
do so, all subject to the following:

The copyright notices in the Software and this entire statement, including
the above license grant, this restriction and the following disclaimer,
must be included in all copies of the Software, in whole or in part, and
all derivative works of the Software, unless such copies or derivative
works are solely in the form of machine-executable object code generated by
a source language processor.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE, TITLE AND NON-INFRINGEMENT. IN NO EVENT
SHALL THE COPYRIGHT HOLDERS OR ANYONE DISTRIBUTING THE SOFTWARE BE LIABLE
FOR ANY DAMAGES OR OTHER LIABILITY, WHETHER IN CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
DEALINGS IN THE SOFTWARE.

```

## Appendix D — Inigo Quilez (MIT)

```text
MIT License

Copyright (c) Inigo Quilez

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Appendix E — Heroicons (MIT) and Lucide (ISC, with Feather MIT)

```text
MIT License

Copyright (c) Tailwind Labs, Inc.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

ISC License

Copyright (c) 2026 Lucide Icons and Contributors

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

---

The following Lucide icons are derived from the Feather project:

airplay, alert-circle, alert-octagon, alert-triangle, aperture, arrow-down-circle, arrow-down-left, arrow-down-right, arrow-down, arrow-left-circle, arrow-left, arrow-right-circle, arrow-right, arrow-up-circle, arrow-up-left, arrow-up-right, arrow-up, at-sign, calendar, cast, check, chevron-down, chevron-left, chevron-right, chevron-up, chevrons-down, chevrons-left, chevrons-right, chevrons-up, circle, clipboard, clock, code, columns, command, compass, corner-down-left, corner-down-right, corner-left-down, corner-left-up, corner-right-down, corner-right-up, corner-up-left, corner-up-right, crosshair, database, divide-circle, divide-square, dollar-sign, download, external-link, feather, frown, hash, headphones, help-circle, info, italic, key, layout, life-buoy, link-2, link, loader, lock, log-in, log-out, maximize, meh, minimize, minimize-2, minus-circle, minus-square, minus, monitor, moon, more-horizontal, more-vertical, move, music, navigation-2, navigation, octagon, pause-circle, percent, plus-circle, plus-square, plus, power, radio, rss, search, server, share, shopping-bag, sidebar, smartphone, smile, square, table-2, tablet, target, terminal, trash-2, trash, triangle, tv, type, upload, x-circle, x-octagon, x-square, x, zoom-in, zoom-out

The MIT License (MIT) (for the icons listed above)

Copyright (c) 2013-present Cole Bemis

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

```

## Appendix F — Apache License 2.0, and Basis Universal NOTICE

```text
Apache License
Version 2.0, January 2004
http://www.apache.org/licenses/

TERMS AND CONDITIONS FOR USE, REPRODUCTION, AND DISTRIBUTION

1. Definitions.

"License" shall mean the terms and conditions for use, reproduction, and distribution as defined by Sections 1 through 9 of this document.

"Licensor" shall mean the copyright owner or entity authorized by the copyright owner that is granting the License.

"Legal Entity" shall mean the union of the acting entity and all other entities that control, are controlled by, or are under common control with that entity. For the purposes of this definition, "control" means (i) the power, direct or indirect, to cause the direction or management of such entity, whether by contract or otherwise, or (ii) ownership of fifty percent (50%) or more of the outstanding shares, or (iii) beneficial ownership of such entity.

"You" (or "Your") shall mean an individual or Legal Entity exercising permissions granted by this License.

"Source" form shall mean the preferred form for making modifications, including but not limited to software source code, documentation source, and configuration files.

"Object" form shall mean any form resulting from mechanical transformation or translation of a Source form, including but not limited to compiled object code, generated documentation, and conversions to other media types.

"Work" shall mean the work of authorship, whether in Source or Object form, made available under the License, as indicated by a copyright notice that is included in or attached to the work (an example is provided in the Appendix below).

"Derivative Works" shall mean any work, whether in Source or Object form, that is based on (or derived from) the Work and for which the editorial revisions, annotations, elaborations, or other modifications represent, as a whole, an original work of authorship. For the purposes of this License, Derivative Works shall not include works that remain separable from, or merely link (or bind by name) to the interfaces of, the Work and Derivative Works thereof.

"Contribution" shall mean any work of authorship, including the original version of the Work and any modifications or additions to that Work or Derivative Works thereof, that is intentionally submitted to Licensor for inclusion in the Work by the copyright owner or by an individual or Legal Entity authorized to submit on behalf of the copyright owner. For the purposes of this definition, "submitted" means any form of electronic, verbal, or written communication sent to the Licensor or its representatives, including but not limited to communication on electronic mailing lists, source code control systems, and issue tracking systems that are managed by, or on behalf of, the Licensor for the purpose of discussing and improving the Work, but excluding communication that is conspicuously marked or otherwise designated in writing by the copyright owner as "Not a Contribution."

"Contributor" shall mean Licensor and any individual or Legal Entity on behalf of whom a Contribution has been received by Licensor and subsequently incorporated within the Work.

2. Grant of Copyright License. Subject to the terms and conditions of this License, each Contributor hereby grants to You a perpetual, worldwide, non-exclusive, no-charge, royalty-free, irrevocable copyright license to reproduce, prepare Derivative Works of, publicly display, publicly perform, sublicense, and distribute the Work and such Derivative Works in Source or Object form.

3. Grant of Patent License. Subject to the terms and conditions of this License, each Contributor hereby grants to You a perpetual, worldwide, non-exclusive, no-charge, royalty-free, irrevocable (except as stated in this section) patent license to make, have made, use, offer to sell, sell, import, and otherwise transfer the Work, where such license applies only to those patent claims licensable by such Contributor that are necessarily infringed by their Contribution(s) alone or by combination of their Contribution(s) with the Work to which such Contribution(s) was submitted. If You institute patent litigation against any entity (including a cross-claim or counterclaim in a lawsuit) alleging that the Work or a Contribution incorporated within the Work constitutes direct or contributory patent infringement, then any patent licenses granted to You under this License for that Work shall terminate as of the date such litigation is filed.

4. Redistribution. You may reproduce and distribute copies of the Work or Derivative Works thereof in any medium, with or without modifications, and in Source or Object form, provided that You meet the following conditions:

     (a) You must give any other recipients of the Work or Derivative Works a copy of this License; and

     (b) You must cause any modified files to carry prominent notices stating that You changed the files; and

     (c) You must retain, in the Source form of any Derivative Works that You distribute, all copyright, patent, trademark, and attribution notices from the Source form of the Work, excluding those notices that do not pertain to any part of the Derivative Works; and

     (d) If the Work includes a "NOTICE" text file as part of its distribution, then any Derivative Works that You distribute must include a readable copy of the attribution notices contained within such NOTICE file, excluding those notices that do not pertain to any part of the Derivative Works, in at least one of the following places: within a NOTICE text file distributed as part of the Derivative Works; within the Source form or documentation, if provided along with the Derivative Works; or, within a display generated by the Derivative Works, if and wherever such third-party notices normally appear. The contents of the NOTICE file are for informational purposes only and do not modify the License. You may add Your own attribution notices within Derivative Works that You distribute, alongside or as an addendum to the NOTICE text from the Work, provided that such additional attribution notices cannot be construed as modifying the License.

     You may add Your own copyright statement to Your modifications and may provide additional or different license terms and conditions for use, reproduction, or distribution of Your modifications, or for any such Derivative Works as a whole, provided Your use, reproduction, and distribution of the Work otherwise complies with the conditions stated in this License.

5. Submission of Contributions. Unless You explicitly state otherwise, any Contribution intentionally submitted for inclusion in the Work by You to the Licensor shall be under the terms and conditions of this License, without any additional terms or conditions. Notwithstanding the above, nothing herein shall supersede or modify the terms of any separate license agreement you may have executed with Licensor regarding such Contributions.

6. Trademarks. This License does not grant permission to use the trade names, trademarks, service marks, or product names of the Licensor, except as required for reasonable and customary use in describing the origin of the Work and reproducing the content of the NOTICE file.

7. Disclaimer of Warranty. Unless required by applicable law or agreed to in writing, Licensor provides the Work (and each Contributor provides its Contributions) on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied, including, without limitation, any warranties or conditions of TITLE, NON-INFRINGEMENT, MERCHANTABILITY, or FITNESS FOR A PARTICULAR PURPOSE. You are solely responsible for determining the appropriateness of using or redistributing the Work and assume any risks associated with Your exercise of permissions under this License.

8. Limitation of Liability. In no event and under no legal theory, whether in tort (including negligence), contract, or otherwise, unless required by applicable law (such as deliberate and grossly negligent acts) or agreed to in writing, shall any Contributor be liable to You for damages, including any direct, indirect, special, incidental, or consequential damages of any character arising as a result of this License or out of the use or inability to use the Work (including but not limited to damages for loss of goodwill, work stoppage, computer failure or malfunction, or any and all other commercial damages or losses), even if such Contributor has been advised of the possibility of such damages.

9. Accepting Warranty or Additional Liability. While redistributing the Work or Derivative Works thereof, You may choose to offer, and charge a fee for, acceptance of support, warranty, indemnity, or other liability obligations and/or rights consistent with this License. However, in accepting such obligations, You may act only on Your own behalf and on Your sole responsibility, not on behalf of any other Contributor, and only if You agree to indemnify, defend, and hold each Contributor harmless for any liability incurred by, or claims asserted against, such Contributor by reason of your accepting any such warranty or additional liability.

END OF TERMS AND CONDITIONS

APPENDIX: How to apply the Apache License to your work.

To apply the Apache License to your work, attach the following boilerplate notice, with the fields enclosed by brackets "[]" replaced with your own identifying information. (Don't include the brackets!)  The text should be enclosed in the appropriate comment syntax for the file format. We also recommend that a file or class name and description of purpose be included on the same "printed page" as the copyright notice for easier identification within third-party archives.

Copyright [yyyy] [name of copyright owner]

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

NOTICE

Basis Universal™ Supercompressed GPU Texture Compression Library

Copyright © 2016–2026 Binomial LLC. 
All rights reserved except as granted under the [Apache 2.0 license](https://github.com/BinomialLLC/basis_universal/blob/master/LICENSE).
"Basis Universal" is a trademark of Binomial LLC.

The documents in the Basis Universal wiki, and the Basis Universal library, example, and tool source code, fall under the Apache 2.0 license, unless otherwise explicitly indicated. 

Redistributions or derivative works must include a readable copy of the attribution notices from this NOTICE file (see Apache License 2.0 §4(d)).

If you modify the Basis Universal source code, specifications, or wiki documents and redistribute the files, you must cause any modified files to carry prominent notices stating that you changed the files (see Apache 2.0 §4(b)).

**This software, documentation and specifications are provided "as is", without warranty of any kind (see Apache 2.0 §§7–8).**

```

## Appendix G — NumPy random (NCSA / BSD-3-Clause) and PCG64 (MIT)

```text
**This software is dual-licensed under the The University of Illinois/NCSA
Open Source License (NCSA) and The 3-Clause BSD License**

# NCSA Open Source License
**Copyright (c) 2019 Kevin Sheppard. All rights reserved.**

Developed by: Kevin Sheppard (<kevin.sheppard@economics.ox.ac.uk>,
<kevin.k.sheppard@gmail.com>)
[http://www.kevinsheppard.com](http://www.kevinsheppard.com)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal with
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimers.

Redistributions in binary form must reproduce the above copyright notice, this
list of conditions and the following disclaimers in the documentation and/or
other materials provided with the distribution.

Neither the names of Kevin Sheppard, nor the names of any contributors may be
used to endorse or promote products derived from this Software without specific
prior written permission.

**THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
CONTRIBUTORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS WITH
THE SOFTWARE.**


# 3-Clause BSD License
**Copyright (c) 2019 Kevin Sheppard. All rights reserved.**

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
   may be used to endorse or promote products derived from this software
   without specific prior written permission.

**THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF
THE POSSIBILITY OF SUCH DAMAGE.**

# Components

Many parts of this module have been derived from original sources, 
often the algorithm's designer. Component licenses are located with 
the component code.

# PCG64

## The MIT License

PCG Random Number Generation for C.

Copyright 2014 Melissa O'Neill <oneill@pcg-random.org>

Permission is hereby granted, free of charge, to any person obtaining 
a copy of this software and associated documentation files (the "Software"), 
to deal in the Software without restriction, including without limitation 
the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in 
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR 
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS 
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR 
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER 
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN 
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

```

## Appendix H — Ubuntu Font Licence 1.0

```text
-------------------------------
UBUNTU FONT LICENCE Version 1.0
-------------------------------

PREAMBLE
This licence allows the licensed fonts to be used, studied, modified and
redistributed freely. The fonts, including any derivative works, can be
bundled, embedded, and redistributed provided the terms of this licence
are met. The fonts and derivatives, however, cannot be released under
any other licence. The requirement for fonts to remain under this
licence does not require any document created using the fonts or their
derivatives to be published under this licence, as long as the primary
purpose of the document is not to be a vehicle for the distribution of
the fonts.

DEFINITIONS
"Font Software" refers to the set of files released by the Copyright
Holder(s) under this licence and clearly marked as such. This may
include source files, build scripts and documentation.

"Original Version" refers to the collection of Font Software components
as received under this licence.

"Modified Version" refers to any derivative made by adding to, deleting,
or substituting -- in part or in whole -- any of the components of the
Original Version, by changing formats or by porting the Font Software to
a new environment.

"Copyright Holder(s)" refers to all individuals and companies who have a
copyright ownership of the Font Software.

"Substantially Changed" refers to Modified Versions which can be easily
identified as dissimilar to the Font Software by users of the Font
Software comparing the Original Version with the Modified Version.

To "Propagate" a work means to do anything with it that, without
permission, would make you directly or secondarily liable for
infringement under applicable copyright law, except executing it on a
computer or modifying a private copy. Propagation includes copying,
distribution (with or without modification and with or without charging
a redistribution fee), making available to the public, and in some
countries other activities as well.

PERMISSION & CONDITIONS
This licence does not grant any rights under trademark law and all such
rights are reserved.

Permission is hereby granted, free of charge, to any person obtaining a
copy of the Font Software, to propagate the Font Software, subject to
the below conditions:

1) Each copy of the Font Software must contain the above copyright
notice and this licence. These can be included either as stand-alone
text files, human-readable headers or in the appropriate machine-
readable metadata fields within text or binary files as long as those
fields can be easily viewed by the user.

2) The font name complies with the following:
(a) The Original Version must retain its name, unmodified.
(b) Modified Versions which are Substantially Changed must be renamed to
avoid use of the name of the Original Version or similar names entirely.
(c) Modified Versions which are not Substantially Changed must be
renamed to both (i) retain the name of the Original Version and (ii) add
additional naming elements to distinguish the Modified Version from the
Original Version. The name of such Modified Versions must be the name of
the Original Version, with "derivative X" where X represents the name of
the new work, appended to that name.

3) The name(s) of the Copyright Holder(s) and any contributor to the
Font Software shall not be used to promote, endorse or advertise any
Modified Version, except (i) as required by this licence, (ii) to
acknowledge the contribution(s) of the Copyright Holder(s) or (iii) with
their explicit written permission.

4) The Font Software, modified or unmodified, in part or in whole, must
be distributed entirely under this licence, and must not be distributed
under any other licence. The requirement for fonts to remain under this
licence does not affect any document created using the Font Software,
except any version of the Font Software extracted from a document
created using the Font Software may only be distributed under this
licence.

TERMINATION
This licence becomes null and void if any of the above conditions are
not met.

DISCLAIMER
THE FONT SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO ANY WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT OF
COPYRIGHT, PATENT, TRADEMARK, OR OTHER RIGHT. IN NO EVENT SHALL THE
COPYRIGHT HOLDER BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
INCLUDING ANY GENERAL, SPECIAL, INDIRECT, INCIDENTAL, OR CONSEQUENTIAL
DAMAGES, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF THE USE OR INABILITY TO USE THE FONT SOFTWARE OR FROM OTHER
DEALINGS IN THE FONT SOFTWARE.

```

## Appendix I — Hash without Sine (MIT)

```text
MIT License

Copyright (c) 2014 David Hoskins.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
