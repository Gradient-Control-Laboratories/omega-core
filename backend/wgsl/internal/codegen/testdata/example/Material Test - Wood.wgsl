fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn map_axisAt(p: vec3<f32>, point: vec3<f32>, axis: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let d = p - point;
  let n = normalize(axis);
  let sz = sign(n.z);
  let s = sz + 1.0 - abs(sz);
  let a = -1.0 / (s + n.z);
  let b = n.x * n.y * a;
  let t1 = vec3<f32>(1.0 + s * n.x * n.x * a, s * b, -(s * n.x));
  let t2 = vec3<f32>(b, s + n.y * n.y * a, -n.y);
  return vec3<f32>(dot(d, t1), dot(d, t2), dot(d, n));
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn gsl_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 60.0, noiseTex) + 0.5 * noise_gradientLattice(p / 30.0, noiseTex);
}

fn gsl_remap_2(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 25.0, noiseTex);
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn material_hiddenShare__spec1(fade: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - fade * gsl_dispDrawn;
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn color_linearToSrgbChannel(l: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x, noiseTex), color_linearToSrgbChannel(c.y, noiseTex), color_linearToSrgbChannel(c.z, noiseTex));
}

fn color_oklab2rgb(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let l_ = c.x + 0.39633778 * c.y + 0.21580376 * c.z;
  let l = l_ * l_ * l_;
  let m_ = c.x - 0.105561346 * c.y - 0.06385417 * c.z;
  let m = m_ * m_ * m_;
  let s_ = c.x - 0.08948418 * c.y - 1.2914855 * c.z;
  let s = s_ * s_ * s_;
  let lms = vec3<f32>(l, m, s);
  let lin = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), lms), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), lms), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), lms));
  return color_gammaEncode(lin, noiseTex);
}

fn noise_cellHash3(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_cornerGradient(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return normalize(noise_cellHash3(cell, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLattice_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, u: f32, v: f32, w: f32, acc: f32, o: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let wx = o.x * u + (1.0 - o.x) * (1.0 - u);
  let wy = o.y * v + (1.0 - o.y) * (1.0 - v);
  let wz = o.z * w + (1.0 - o.z) * (1.0 - w);
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + wx * wy * wz * dot(g, d);
}

fn noise_gradientLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let pi = floor(q);
  let pf = q - pi;
  let u = ease_cubic(pf.x, noiseTex);
  let v = ease_cubic(pf.y, noiseTex);
  let w = ease_cubic(pf.z, noiseTex);
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLattice_gsl_anon(pi, pf, u, v, w, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn wood_ringBoundary(k: f32, seed: f32, noiseTex: texture_3d<f32>) -> f32 {
  return k + (fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(k, k + seed, 2.0 * k + 17.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(k, k + seed, 2.0 * k + 17.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.5;
}

fn wood_poreZoneShare(poreZone: f32, ringWidth: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(0.3, min(poreZone / ringWidth, 0.3), poreZone > 0.0);
}

fn wood_poreSizeShare(progress: f32, ringPorosity: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let semiSize = 1.0 - 0.5 * progress;
  let earlywoodRamp = 1.0 - texture_mask(progress, poreZone, poreZone, noiseTex);
  let ringSize = 0.35 + 0.65 * earlywoodRamp;
  let diffuseSize = 1.0 - 0.15 * progress;
  let size = select(semiSize + (ringSize - semiSize) * (2.0 * ringPorosity - 1.0), diffuseSize + (semiSize - diffuseSize) * 2.0 * ringPorosity, ringPorosity < 0.5);
  return vec2<f32>(size, 0.35 + 0.45 * max(2.0 * ringPorosity - 1.0, 0.0));
}

fn wood_poreDistanceAt(local: vec3<f32>, progress: f32, poreSize: f32, ringPorosity: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let poreCellSize = max(poreSize, 0.01);
  let poreQ = local.xy / poreCellSize;
  let poreBase = floor(poreQ - 0.5);
  let gsl_tmp_87 = poreBase + vec2<f32>(0.0, 0.0);
  let gsl_tmp_88 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_87.x, gsl_tmp_87.y, gsl_tmp_87.x - gsl_tmp_87.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_87.x, gsl_tmp_87.y, gsl_tmp_87.x - gsl_tmp_87.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_89 = floor(gsl_tmp_88);
  let gsl_tmp_90 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_87.x, gsl_tmp_87.y, gsl_tmp_89))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_87.x, gsl_tmp_87.y, gsl_tmp_89)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let sizeShare = wood_poreSizeShare(progress, ringPorosity, poreZone, noiseTex);
  let size = sizeShare.x;
  let poreShare = sizeShare.y;
  let gsl_tmp_91 = select(10.0, min(10.0, length(vec2<f32>(length(poreQ - (gsl_tmp_87 + gsl_tmp_90.xy)), max(abs(gsl_tmp_88 - gsl_tmp_89 - 0.5) - 0.2 - 0.25 * gsl_tmp_90.x, 0.0) * 8.0)) / size), gsl_tmp_90.z < poreShare);
  let gsl_tmp_92 = poreBase + vec2<f32>(1.0, 0.0);
  let gsl_tmp_93 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_92.x, gsl_tmp_92.y, gsl_tmp_92.x - gsl_tmp_92.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_92.x, gsl_tmp_92.y, gsl_tmp_92.x - gsl_tmp_92.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_94 = floor(gsl_tmp_93);
  let gsl_tmp_95 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_92.x, gsl_tmp_92.y, gsl_tmp_94))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_92.x, gsl_tmp_92.y, gsl_tmp_94)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_96 = select(gsl_tmp_91, min(gsl_tmp_91, length(vec2<f32>(length(poreQ - (gsl_tmp_92 + gsl_tmp_95.xy)), max(abs(gsl_tmp_93 - gsl_tmp_94 - 0.5) - 0.2 - 0.25 * gsl_tmp_95.x, 0.0) * 8.0)) / size), gsl_tmp_95.z < poreShare);
  let gsl_tmp_97 = poreBase + vec2<f32>(0.0, 1.0);
  let gsl_tmp_98 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_97.x, gsl_tmp_97.y, gsl_tmp_97.x - gsl_tmp_97.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_97.x, gsl_tmp_97.y, gsl_tmp_97.x - gsl_tmp_97.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_99 = floor(gsl_tmp_98);
  let gsl_tmp_100 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_97.x, gsl_tmp_97.y, gsl_tmp_99))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_97.x, gsl_tmp_97.y, gsl_tmp_99)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_101 = select(gsl_tmp_96, min(gsl_tmp_96, length(vec2<f32>(length(poreQ - (gsl_tmp_97 + gsl_tmp_100.xy)), max(abs(gsl_tmp_98 - gsl_tmp_99 - 0.5) - 0.2 - 0.25 * gsl_tmp_100.x, 0.0) * 8.0)) / size), gsl_tmp_100.z < poreShare);
  let gsl_tmp_102 = poreBase + vec2<f32>(1.0, 1.0);
  let gsl_tmp_103 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_102.x, gsl_tmp_102.y, gsl_tmp_102.x - gsl_tmp_102.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_102.x, gsl_tmp_102.y, gsl_tmp_102.x - gsl_tmp_102.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_104 = floor(gsl_tmp_103);
  let gsl_tmp_105 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_102.x, gsl_tmp_102.y, gsl_tmp_104))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_102.x, gsl_tmp_102.y, gsl_tmp_104)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let nearestPore = select(gsl_tmp_101, min(gsl_tmp_101, length(vec2<f32>(length(poreQ - (gsl_tmp_102 + gsl_tmp_105.xy)), max(abs(gsl_tmp_103 - gsl_tmp_104 - 0.5) - 0.2 - 0.25 * gsl_tmp_105.x, 0.0) * 8.0)) / size), gsl_tmp_105.z < poreShare);
  return vec2<f32>(nearestPore, size);
}

fn wood_poreSurfaceAt(local: vec3<f32>, progress: f32, ringWidth: f32, poreSize: f32, ringPorosity: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let poreHere = wood_poreDistanceAt(local, progress, poreSize, ringPorosity, wood_poreZoneShare(poreZone, ringWidth, noiseTex), noiseTex);
  let u = min(poreHere.x / 0.28, 1.0);
  let fade = texture_fadeDetail(poreSize * 0.56 * poreHere.y, noiseTex);
  return vec2<f32>(-0.28 * poreSize * poreHere.y * (1.0 - u * u) * fade, (1.0 - texture_mask(poreHere.x, 0.28, 0.06, noiseTex)) * fade);
}

fn wood_woodFinish(rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_make(rough, 0.0, 1.5, noiseTex);
}

fn wood_fibreShade(column: vec2<f32>, z: f32, noiseTex: texture_3d<f32>) -> f32 {
  let run = z / 8.0 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(column.x, column.y, column.x - column.y + 7919.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(column.x, column.y, column.x - column.y + 7919.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let index = floor(run);
  let here = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(column.x, column.y, index))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(column.x, column.y, index)) / 64.0)) & vec3<i32>(63), 0).rgb).x;
  let next = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(column.x, column.y, index + 1.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(column.x, column.y, index + 1.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x;
  return here + (next - here) * texture_mask(run - index, 0.925, 0.075, noiseTex);
}

fn wood_fibreFile(file: f32, radius: f32, z: f32, noiseTex: texture_3d<f32>) -> f32 {
  let q = radius / 0.45 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(file, file + 3.0, 2.0 * file + 7919.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(file, file + 3.0, 2.0 * file + 7919.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5;
  let cell = floor(q);
  let here = wood_fibreShade(vec2<f32>(file, cell), z, noiseTex);
  let t = q - cell;
  return here + (wood_fibreShade(vec2<f32>(file, cell + 1.0), z, noiseTex) - here) * t * t * (3.0 - 2.0 * t);
}

fn wood_poreSizeShare__spec1(progress: f32, poreZone: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let semiSize = 1.0 - 0.5 * progress;
  let earlywoodRamp = 1.0 - texture_mask(progress, poreZone, poreZone, noiseTex);
  let ringSize = 0.35 + 0.65 * earlywoodRamp;
  let size = semiSize + (ringSize - semiSize) * 1.0;
  return vec2<f32>(size, 0.8);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-25.0, -25.0, -25.0);
  let gsl_partHi = vec3<f32>(25.0, 25.0, 25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = length(p) - 25.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let ball = gsl_part;
  let gsl_s1 = gsl_shape(ball, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let gsl_tmp_165 = map_axisAt(p, vec3<f32>(0.0, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
    let gsl_tmp_166 = length(gsl_tmp_165.xy);
    let gsl_tmp_167 = gsl_tmp_165.xy / max(gsl_tmp_166, 0.001) * 40.0;
    let gsl_tmp_138 = vec2<f32>(gsl_tmp_166 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_167.x, gsl_tmp_167.y, gsl_tmp_165.z * 0.25), noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_165.x, gsl_tmp_165.y, gsl_tmp_165.z * 0.2), noiseTex), 0.0);
    let gsl_tmp_139 = gsl_tmp_138.x;
    let gsl_tmp_140 = floor(gsl_tmp_139);
    let gsl_tmp_141 = wood_ringBoundary(gsl_tmp_140, 3.0, noiseTex);
    let gsl_tmp_142 = wood_ringBoundary(gsl_tmp_140 + 1.0, 3.0, noiseTex);
    let gsl_tmp_144 = gsl_tmp_139 >= gsl_tmp_142;
    let gsl_tmp_143 = gsl_tmp_139 < gsl_tmp_141;
    let gsl_tmp_145 = select(select(gsl_tmp_141, gsl_tmp_142, gsl_tmp_144), wood_ringBoundary(gsl_tmp_140 - 1.0, 3.0, noiseTex), gsl_tmp_143);
    let gsl_tmp_147 = (gsl_tmp_139 - gsl_tmp_145) / (select(select(gsl_tmp_142, wood_ringBoundary(gsl_tmp_140 + 2.0, 3.0, noiseTex), gsl_tmp_144), gsl_tmp_141, gsl_tmp_143) - gsl_tmp_145);
    let gsl_tmp_146 = select(select(gsl_tmp_140, gsl_tmp_140 + 1.0, gsl_tmp_144), gsl_tmp_140 - 1.0, gsl_tmp_143);
    let gsl_tmp_148 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_146, gsl_tmp_146 + 11.0, 3.0 * gsl_tmp_146 + 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_146, gsl_tmp_146 + 11.0, 3.0 * gsl_tmp_146 + 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let gsl_tmp_149 = texture_mask(gsl_tmp_147, 0.4 + (gsl_tmp_148.x - 0.5) * 0.027, 0.3, noiseTex) * (1.0 + (gsl_tmp_148.y - 0.5) * 0.18);
    let gsl_tmp_136 = map_axisAt(p, vec3<f32>(0.0, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
    let gsl_tmp_137 = length(gsl_tmp_136.xy);
    var gsl_detailValue = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.075) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_150 = (atan2(gsl_tmp_136.y, gsl_tmp_136.x) + 3.1415927) / 6.2831855 * 3351.0 - 0.5;
      let gsl_tmp_151 = floor(gsl_tmp_150);
      let gsl_tmp_153 = wood_fibreFile(gsl_floored(gsl_tmp_151, 3351.0), gsl_tmp_137, gsl_tmp_136.z, noiseTex);
      let gsl_tmp_152 = gsl_tmp_150 - gsl_tmp_151;
      gsl_detailValue = (gsl_tmp_153 + (wood_fibreFile(gsl_floored((gsl_tmp_151 + 1.0), 3351.0), gsl_tmp_137, gsl_tmp_136.z, noiseTex) - gsl_tmp_153) * gsl_tmp_152 * gsl_tmp_152 * (3.0 - 2.0 * gsl_tmp_152) - 0.5) * 0.08 * (1.0 - 0.6 * gsl_tmp_149) * texture_fadeDetail(min(0.15, 6.2831855 * gsl_tmp_137 / 3351.0), noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    let gsl_tmp_156 = wood_poreSizeShare__spec1(gsl_tmp_147, 0.3, noiseTex).x;
    var gsl_detailValue_2 = 0.05 * gsl_tmp_156 * gsl_tmp_156;
    if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_154 = wood_poreDistanceAt(gsl_tmp_136, gsl_tmp_147, 0.5, 1.0, 0.3, noiseTex);
      let gsl_tmp_155 = texture_fadeDetail(0.28 * gsl_tmp_154.y, noiseTex);
      gsl_detailValue_2 = (1.0 - texture_mask(gsl_tmp_154.x, 0.28, 0.06, noiseTex)) * gsl_tmp_155 + 0.05 * gsl_tmp_154.y * gsl_tmp_154.y * (1.0 - gsl_tmp_155);
      gsl_pw = gsl_detailOuterPw_2;
    }
    var gsl_simpleValue = vec3<f32>(0.59386784, 0.47988868, 0.3859136);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_157 = (atan2(gsl_tmp_136.y, gsl_tmp_136.x) + 3.1415927) / 6.2831855 * 250.0;
      let gsl_tmp_158 = gsl_floored(floor(gsl_tmp_157), 250.0);
      let gsl_tmp_159 = gsl_tmp_136.z / 12.0 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_158, gsl_tmp_158 + 17.0, 3.0 * gsl_tmp_158 + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_158, gsl_tmp_158 + 17.0, 3.0 * gsl_tmp_158 + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
      let gsl_tmp_160 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_158, floor(gsl_tmp_159), gsl_tmp_158 + 401.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_158, floor(gsl_tmp_159), gsl_tmp_158 + 401.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_161 = vec3<f32>(0.6289975, 0.025104912, 0.04613428) + vec3<f32>(-0.044705804, -0.0014734669, -0.0036211603) * gsl_tmp_149 + vec3<f32>((fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_146, gsl_tmp_146 + 5.0, 2.0 * gsl_tmp_146 + 29.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_146, gsl_tmp_146 + 5.0, 2.0 * gsl_tmp_146 + 29.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.08 + gsl_detailValue + 0.1 * 1.0 * select(0.0, 1.0, gsl_tmp_160.z < 0.5) * (1.0 - texture_mask(abs(gsl_tmp_157 - floor(gsl_tmp_157) - 0.2 - 0.6 * gsl_tmp_160.x) * 6.2831855 * gsl_tmp_137 / 250.0, 0.15, 0.05, noiseTex)) * (1.0 - texture_mask(abs(gsl_tmp_159 - floor(gsl_tmp_159) - 0.5) * 2.0 / (0.35 + 0.65 * gsl_tmp_160.y), 1.0, 0.1, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex)) + 0.0, 0.0, 0.0) + vec3<f32>(0.042500425, 0.00060789636, -0.003918465) * texture_mask(gsl_tmp_138.x, 999979.3, 0.3, noiseTex);
      let gsl_tmp_162 = gsl_tmp_161 + (vec3<f32>(0.0, 0.0, 0.0) - gsl_tmp_161) * 0.0 * (1.0 - texture_mask(abs(gsl_tmp_138.x - 999979.3 + 0.3), 0.3, 0.1, noiseTex));
      let gsl_tmp_163 = gsl_tmp_162 + (vec3<f32>(0.0, 0.0, 0.0) - gsl_tmp_162) * 0.0;
      let gsl_tmp_164 = gsl_tmp_163 + (vec3<f32>(0.45476237, 0.037886485, 0.05228309) - gsl_tmp_163) * 0.85 * gsl_detailValue_2 - vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_116 = color_oklab2rgb(gsl_tmp_164 + (vec3<f32>(0.41366172, 0.0160504, 0.030504888) - gsl_tmp_164) * 0.8 * gsl_tmp_138.y, noiseTex);
      gsl_simpleValue = mix(gsl_simpleValue, gsl_tmp_116, gsl_simpleDetail);
    }
    let oakColor = gsl_simpleValue;
    return vec4<f32>(oakColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    var gsl_detailValue_3 = vec2<f32>(0.0, 0.0);
    if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_3 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_133 = map_axisAt(p, vec3<f32>(0.0, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
      let gsl_tmp_134 = length(gsl_tmp_133.xy);
      let gsl_tmp_135 = gsl_tmp_133.xy / max(gsl_tmp_134, 0.001) * 40.0;
      let gsl_tmp_117 = gsl_tmp_134 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_135.x, gsl_tmp_135.y, gsl_tmp_133.z * 0.25), noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_133.x, gsl_tmp_133.y, gsl_tmp_133.z * 0.2), noiseTex);
      let gsl_tmp_118 = floor(gsl_tmp_117);
      let gsl_tmp_119 = wood_ringBoundary(gsl_tmp_118, 3.0, noiseTex);
      let gsl_tmp_120 = wood_ringBoundary(gsl_tmp_118 + 1.0, 3.0, noiseTex);
      let gsl_tmp_122 = gsl_tmp_117 >= gsl_tmp_120;
      let gsl_tmp_121 = gsl_tmp_117 < gsl_tmp_119;
      let gsl_tmp_123 = select(select(gsl_tmp_119, gsl_tmp_120, gsl_tmp_122), wood_ringBoundary(gsl_tmp_118 - 1.0, 3.0, noiseTex), gsl_tmp_121);
      gsl_detailValue_3 = wood_poreSurfaceAt(map_axisAt(p, vec3<f32>(0.0, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex), (gsl_tmp_117 - gsl_tmp_123) / (select(select(gsl_tmp_120, wood_ringBoundary(gsl_tmp_118 + 2.0, 3.0, noiseTex), gsl_tmp_122), gsl_tmp_119, gsl_tmp_121) - gsl_tmp_123), 1.5, 0.5, 1.0, 0.0, noiseTex) * 1.0;
      gsl_pw = gsl_detailOuterPw_3;
    }
    var gsl_simpleValue_2 = mat4x4<f32>(vec4<f32>(0.6685386, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_108 = gsl_detailValue_3;
      let gsl_tmp_110 = 0.5851758 + 0.3148242 * gsl_tmp_108.y;
      let gsl_tmp_111 = vec2<f32>(gsl_tmp_110, gsl_tmp_108.x);
      let gsl_simpleFull = material_withDisplacement(wood_woodFinish(sqrt(sqrt(gsl_tmp_111.x * gsl_tmp_111.x * gsl_tmp_111.x * gsl_tmp_111.x + 0.0825 * material_hiddenShare__spec1(1.0 - texture_mask(gsl_pw, 0.105, 0.035, noiseTex), noiseTex))), noiseTex), gsl_tmp_111.y, noiseTex);
      gsl_simpleValue_2 = mat4x4<f32>(mix(gsl_simpleValue_2[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_2[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_2[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_2[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let oakMat = gsl_simpleValue_2;
    let gsl_m = oakMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));
