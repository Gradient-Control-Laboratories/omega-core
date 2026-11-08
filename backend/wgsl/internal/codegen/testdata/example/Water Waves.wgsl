fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn gsl_remap(p: vec3<f32>, gsl_param_12_10: f32, gsl_param_15_10: f32, gsl_param_18_10: f32, gsl_param_24_13: f32, gsl_param_30_12: f32, gsl_param_33_14: f32, time: f32, noiseTex: texture_3d<f32>) -> f32 {
  let waveSize = gsl_param_30_12;
  let waveMorph = gsl_param_24_13;
  return noise_gradientLattice4(vec4<f32>(p.x / waveSize, p.y / waveSize, p.z / waveSize, time / waveMorph), noiseTex) * waveSize / 10.0;
}

fn noise_cellHash4(cell4: vec4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let cx = cell4.x;
  let cw = cell4.w;
  let cy = cell4.y;
  let cz = cell4.z;
  let hA = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cx + 17.0 * cw, cy + 29.0 * cw, cz + 41.0 * cw))) & vec3<i32>(63), 0).rgb);
  let hB = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cy + 53.0 * cw, cz + 61.0 * cw, cx + 73.0 * cw))) & vec3<i32>(63), 0).rgb);
  return vec4<f32>(hA.x, hA.y, hA.z, hB.x);
}

fn noise_cornerGradient4(cell4: vec4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return normalize(noise_cellHash4(cell4, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLattice4_gsl_anon(pi: vec4<f32>, pf: vec4<f32>, u: f32, v: f32, w: f32, t: f32, acc: f32, o: vec4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let wx = o.x * u + (1.0 - o.x) * (1.0 - u);
  let wy = o.y * v + (1.0 - o.y) * (1.0 - v);
  let wz = o.z * w + (1.0 - o.z) * (1.0 - w);
  let wt = o.w * t + (1.0 - o.w) * (1.0 - t);
  let g = noise_cornerGradient4(pi + o, noiseTex);
  let d = pf - o;
  return acc + wx * wy * wz * wt * dot(g, d);
}

fn noise_gradientLattice4(q4: vec4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let pi = floor(q4);
  let pf = q4 - pi;
  let u = ease_cubic(pf.x, noiseTex);
  let v = ease_cubic(pf.y, noiseTex);
  let w = ease_cubic(pf.z, noiseTex);
  let t = ease_cubic(pf.w, noiseTex);
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 16; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLattice4_gsl_anon(pi, pf, u, v, w, t, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_12_10: f32, gsl_param_15_10: f32, gsl_param_18_10: f32, gsl_param_24_13: f32, gsl_param_30_12: f32, gsl_param_33_14: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let waveHeight = gsl_param_33_14;
  let waveVx = gsl_param_12_10;
  let waveVy = gsl_param_15_10;
  let waveVz = gsl_param_18_10;
  let drift = p - vec3<f32>(waveVx, waveVy, waveVz) * time;
  let waves = waveHeight * gsl_remap(drift, gsl_param_12_10, gsl_param_15_10, gsl_param_18_10, gsl_param_24_13, gsl_param_30_12, gsl_param_33_14, time, noiseTex);
  let water = (length(p) - 12.0 - waves) / (1.0 + waveHeight);
  let gsl_s1 = gsl_shape(water, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_12_10: f32, gsl_param_15_10: f32, gsl_param_18_10: f32, gsl_param_24_13: f32, gsl_param_30_12: f32, gsl_param_33_14: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.15, 0.45, 0.75), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_12_10: f32, gsl_param_15_10: f32, gsl_param_18_10: f32, gsl_param_24_13: f32, gsl_param_30_12: f32, gsl_param_33_14: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.333, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.6, 24.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
const gsl_foldK = array<vec4<f32>, 16>(vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 1.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 0.0, 1.0, 0.0), vec4<f32>(0.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(1.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 1.0, 0.0, 1.0), vec4<f32>(1.0, 1.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 1.0, 1.0), vec4<f32>(1.0, 0.0, 1.0, 1.0), vec4<f32>(0.0, 1.0, 1.0, 1.0), vec4<f32>(1.0, 1.0, 1.0, 1.0));
