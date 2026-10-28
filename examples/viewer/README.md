# GSL WebGPU Standalone 3D Viewer

A zero-dependency client-side HTML5/CSS/JavaScript WebGPU application for visualizing procedural
Geometry Shading Language (GSL) models and hybrid CAD polygon meshes in real time.

## Quickstart

Double-click `index.html` or run:

```bash
open examples/viewer/index.html
```

(Requires any WebGPU-enabled browser: Chrome 113+, Edge 113+, Arc, or Safari 18+ on macOS Sequoia).

## Features

- **Interactive 3D Viewport**: Smooth orbit controls (left drag to rotate, middle click to pan,
  wheel to zoom) and standard view presets (Top, Front, Right, Isometric).
- **Continuous Animation**: Play/Pause toggle (spacebar), scrubbable timeline, and speed multipliers
  evaluating `u_time` at 60+ FPS.
- **Parametric Sliders**: Auto-loads companion `.json` metadata emitted by `bin/gslc`, dynamically
  rendering HTML range sliders that update WebGPU uniform buffers at 60 FPS without recompiling.
- **Hybrid Traditional Mesh + Procedural SDF**: Renders CAD polygon meshes (STL/OBJ) and procedural
  SDFs together with shared hardware depth testing (`depth01 = clipPos.z / clipPos.w`), wireframe
  overlay, and layer visibility toggles.
- **Cross-Section Slicing**: Interactive cutting planes along the Z axis to inspect internal
  lattice infills and wall clearances.
- **Lighting Presets**: Studio Neutral, Warm Sunset, Cool Industrial, Dramatic, and Metallic.
- **Hot-Reloading**: Drag-and-drop any `.wgsl` shader or `.stl`/`.obj` mesh directly onto the window.
- **High-Res Snapshot**: One-click PNG capture.

## Compiling Models for the Viewer

```bash
# Compile any GSL model to WGSL with companion metadata:
bin/gslc -t wgsl -o examples/viewer/shaders/my_model.wgsl my_model.gsl
```
