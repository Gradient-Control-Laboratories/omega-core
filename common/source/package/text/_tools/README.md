# text — tooling

CXF source font and the `cxf2gsl` converter that generates the `text` GSL
package (`common/source/package/text/text.gsl`) consumed by the `text()`
renderer.

## Source & license

`Unicode.cxf` is the complete "Unicode" stroke font from the
[QCAD](https://qcad.org) project (RibbonSoft GmbH) — 2745 glyphs, authored by
Andrew Mustun (Latin), Eugene Osintsev (Cyrillic), and Yoshimune Kobayashi
(Japanese, including ~1960 Kanji).

We currently **generate** up to U+3040 (Latin, Greek, Cyrillic, symbols, and
CJK punctuation — no Kana, no Kanji). Compile-time const-fold dispatch and
partial-const specialization let `text("Hi", ...)` collapse to just the
glyphs the call actually uses, so a per-character dispatch cascade no longer
costs anything at runtime — meaning the full range, Kana and Kanji included,
can be generated whenever needed (raise the converter's `-hi`).

Per QCAD's `LICENSE.txt`, the bundled CXF fonts — other than Hershey, osifont,
and VL-Gothic — are released as **public domain** (all copyright is waived):

> Other CXF fonts in directory 'fonts' are released as public domain (all
> copyright is waived).

This is also recorded in the app's third-party notices
(`common/web/about/third-party-notices.md`).

## Regenerating `text/text.gsl`

`cxf2gsl` converts a CXF font into a self-contained GSL package (no imports,
so it builds in every backend). Latin, Greek, Cyrillic, symbols, CJK
punctuation (520 glyphs).

Regenerate from `common/`:

    go run ./source/package/text/_tools/cxf2gsl -in source/package/text/_tools/Unicode.cxf -out source/package/text/text.gsl

`cxf2gsl` also supports `-ascii "string"` (terminal preview) and `-png <path>`
(glyph sprite sheet) for visual inspection, plus `-lo`/`-hi`/`-include` for
custom subsets.
