# GSL Interactive WebGPU Example Viewer

An interactive WebGPU 3D viewport and live GSL code editor for the `omega-core` SDK, served by the
local `gsl serve` HTTP server (`cmd/serve`).

## Quickstart

From the `omega-core` repository root, build and launch the viewer server:

```bash
./tool/build-cmds.sh
./bin/gsl serve --open
```

Or run directly with `go run`:

```bash
go run ./cmd/serve --open
```

The `--open` (`-o`) flag tells `serve` to automatically open `http://127.0.0.1:8080`
in your default WebGPU-capable browser (Chrome 113+, Edge 113+, or Safari 18+ on macOS Sequoia).

## Features

- **Unified Top Bar**:
  - **Code Toggle**: Slide out the 50% width CodeMirror 6 GSL/WGSL editor (`GSL` and `WGSL` tabs,
    line-wrap toggle, and built-in GSL language/library reference).
  - **Example Selector**: Switch between any `.gsl` file in `examples/` and immediately compile and
    render it.
  - **Save As**: When you modify an example (via the code editor, parameter sliders, 3D gizmos,
    mesh imports, or surface point picking), click **Save As** to save a new `.gsl` file into
    `examples/`.
- **WebGPU Raymarching Viewport (`render/`)**:
  - Multi-tier rendering (**Simple**, **Full** PBR with environment lighting, refractions, and
    surface displacement, and progressive **Trace** path tracing).
  - Interactive 3D gizmos (`@point`, `@direction`, `@transform`, `@spline`) and orthographic /
    perspective orientation gizmo.
  - Surface picking tools on the left canvas toolbar to insert `[x, y, z]` points or
    `[x, y, z], [nx, ny, nz]` point + normal pairs directly at the editor cursor.
- **Inspector Panels (`panels/`)**:
  - **Parameters**: Live sliders for `.min`/`.max` variables and color chips for `@color` controls
    that update uniforms at 60 FPS and rewrite GSL source spans, with the animation transport at
    the bottom: a `time` range, rewind / play-pause / fast-forward, the end time, and the
    synthesized Web Audio volume for programs with `audio(...)` outputs.
  - **Imports**: Drag-and-drop `.stl` or `.obj` files to render hybrid B-rep meshes alongside
    implicit SDFs.
  - **Scene**: The bounding box size and grid spacing in mm; a document's `@scene` directive sets
    them, and the Section ranges follow the size.
  - **Render**: Select viewer resolution, rest rendering tier, environment preset, detail threshold,
    and high-sample PNG render jobs.
  - **Section**: Slice the SDF field along an arbitrary plane with isolevel contours and
    adjustable model/section opacity.
  - **Bodies**: Every body the document returns with its visibility, color and surface; picking a
    color or a swatch from the material catalog writes the `.color` / `.material` lines into the
    GSL source, and the palette strip sets the viewer's default colors.
  - **Export**: Server-side Dual Contouring / Marching Cubes export (`POST /api/mesh`) with live
    stage progress, as 3MF (each body an object in its Bodies-panel color) or STL (a zip of one STL
    per body when several are shown). Only the bodies the Bodies panel shows are exported.

## Directory Layout

- `index.html`, `main.js`, `style.css`: Top-level application shell, HTTP `/api/*` client, and panel
  wiring.
- `render/`: Three.js/WebGPU raymarching viewport (`viewer.js`), TSL/WGSL render graph, shading,
  glass, displacement, path tracing, 3D gizmos, and audio worklet processor.
- `panels/`: Inspector panel controllers and styles (`slider-panel`, `mesh-import`, `export-panel`,
  `detail-slider`, `body-panel` with `body-palette`, `material-grid`, `material-assign` and
  `finish-tint`).
- `materials/`: The swatch catalog the Bodies panel reads (`catalog.json`, the swatch sheet, and the
  large previews), generated from the `@material` / `@finish` annotations in `stdlib/`.
- `ui/`: Shared UI widgets (`event-manager`, `notification`, `error`, `progress-bar`, `panel-resize`,
  `tooltip`).
- `editor/`: CodeMirror 6 GSL/WGSL editor, syntax diagnostics, search panel, and buffer tabs.
- `docs/`: Built-in GSL language guide and standard library reference viewer.
- `vendor/`: Bundled browser dependencies (Three.js WebGPU/TSL, CodeMirror 6, `three-mesh-bvh`,
  Manifold).
