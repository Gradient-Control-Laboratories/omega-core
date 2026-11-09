package codegen

// The noise tile is a single global texture threaded into every shader function
// that (transitively) calls noiseHash3 — the wgslFn pipeline binds textures by
// parameter name, so there is no module-global binding to reach instead. Mirrors
// the bvhSdf/vdbSdf slot threading, but it's one lone texture (textureLoad needs
// no sampler) and gated on a single bool rather than per-asset slots.
const noiseTexName = "noiseTex"
const noiseTexParam = noiseTexName + ": texture_3d<f32>"
