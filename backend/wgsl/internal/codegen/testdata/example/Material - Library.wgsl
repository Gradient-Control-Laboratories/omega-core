fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
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

fn noise_cellularF1F2PairLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct {
  var gsl_result : gsl_struct;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  gsl_result.v_3 = gsl_cellularScan(q, 1, noiseTex).nearCell;
  gsl_result.v_4 = gsl_cellularScan(q, 1, noiseTex).farCell;
  return gsl_result;
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn color_srgbToLinearChannel(s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x, noiseTex), color_srgbToLinearChannel(c.y, noiseTex), color_srgbToLinearChannel(c.z, noiseTex));
}

fn color_cbrt(x: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn color_rgb2oklab(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let lin = color_linearize(c, noiseTex);
  let l = dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), lin);
  let l_ = color_cbrt(l, noiseTex);
  let m = dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), lin);
  let m_ = color_cbrt(m, noiseTex);
  let s = dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), lin);
  let s_ = color_cbrt(s, noiseTex);
  let lms = vec3<f32>(l_, m_, s_);
  return vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), lms), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), lms), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), lms));
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

fn builtin_sphere(p: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - radius;
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
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

fn gsl_remap(p: vec3<f32>, color_white: vec3<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 60.0, noiseTex) + 0.5 * noise_gradientLattice(p / 30.0, noiseTex);
}

fn gsl_remap_2(p: vec3<f32>, color_white: vec3<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(p / 25.0, noiseTex);
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
}

fn noise_cellularF1F2Lattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = gsl_cellularScan(q, 1, noiseTex).f1;
  gsl_result.v_2 = gsl_cellularScan(q, 1, noiseTex).f2;
  return gsl_result;
}

fn noise_cellularEdges(p: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_result_3 = noise_cellularF1F2Lattice(p / size, noiseTex);
  let f2 = gsl_result_3.v_2;
  let f1 = gsl_result_3.v_1;
  return (f2 - f1) * size;
}

fn texture_paintMaterial(field: f32, edge: f32, width: f32, matA: mat4x4<f32>, matB: mat4x4<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_blend(matA, matB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn noise_cellularHash(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
}

fn noise_voronoiPoint(c: vec3<f32>, randomness: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return c + 0.5 + (noise_cellularHash(c, noiseTex) - 0.5) * randomness;
}

fn noise_voronoiMetric(v: vec3<f32>, metric: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(select(max(abs(v.x), max(abs(v.y), abs(v.z))), abs(v.x) + abs(v.y) + abs(v.z), metric < 1.5), length(v), metric < 0.5);
}

fn noise_voronoi_gsl_anon(base: vec3<f32>, randomness: f32, q: vec3<f32>, metric: f32, k: f32, a: vec2<f32>, o: vec3<f32>, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let d = noise_voronoiMetric(noise_voronoiPoint(base + o, randomness, noiseTex) - q, metric, noiseTex);
  let m = min(a.x, d);
  return vec2<f32>(m, a.y * exp((m - a.x) / k) + exp((m - d) / k));
}

fn noise_voronoi(p: vec3<f32>, size: f32, randomness: f32, metric: f32, smoothness: f32, noiseTex: texture_3d<f32>) -> f32 {
  let q = p / size;
  let base = floor(q);
  let k = max(smoothness, 0.0001);
  var gsl_fold = vec2<f32>(64.0, 0.0);
  for (var gsl_i = 0; gsl_i < 27; gsl_i = gsl_i + 1) {
    gsl_fold = noise_voronoi_gsl_anon(base, randomness, q, metric, k, gsl_fold, gsl_foldK_2[gsl_i], noiseTex);
  }
  let acc = gsl_fold;
  let smoothed = acc.x - k * log(acc.y);
  return max(select(acc.x, smoothed, smoothness > 0.0), 0.0) * size;
}

fn gsl_remap_3(p: vec3<f32>, color_white: vec3<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_383 = clamp((-noise_voronoi(p, 5.0, 1.0, 0.0, 0.0, noiseTex) / 5.0 - -0.38) / 0.08 * 0.5 + 0.5, 0.0, 1.0);
  return gsl_tmp_383 * gsl_tmp_383 * (3.0 - 2.0 * gsl_tmp_383);
}

fn texture_paint(field: f32, edge: f32, width: f32, colorA: vec3<f32>, colorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return color_mixOklab(colorA, colorB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn gsl_remap_4(p: vec3<f32>, color_white: vec3<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_623 = clamp((-noise_voronoi(p, 5.0, 1.0, 0.0, 0.0, noiseTex) / 5.0 - -0.38) / 0.08 * 0.5 + 0.5, 0.0, 1.0);
  return gsl_tmp_623 * gsl_tmp_623 * (3.0 - 2.0 * gsl_tmp_623);
}

fn texture_coverageDisc(gsl_var_distance: f32, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let drawn = max(radius, 0.5 * gsl_pw);
  let soft = max(0.5 * gsl_pw, 0.01 * radius);
  return (1.0 - texture_mask(gsl_var_distance, drawn, soft, noiseTex)) * radius / drawn * radius / drawn;
}

fn texture_squareWaveIntegral(x: f32, period: f32, noiseTex: texture_3d<f32>) -> f32 {
  let whole = floor(x / period);
  return whole * 0.5 * period + max(x - whole * period - 0.5 * period, 0.0);
}

fn texture_squareWave(x: f32, period: f32, noiseTex: texture_3d<f32>) -> f32 {
  let exact = select(1.0, 0.0, fract(x / period) < 0.5);
  let width = gsl_pw;
  let averaged = (texture_squareWaveIntegral(x + 0.5 * width, period, noiseTex) - texture_squareWaveIntegral(x - 0.5 * width, period, noiseTex)) / max(width, 0.000001);
  return select(exact, averaged, width > 0.0001 * period);
}

fn noise_driftingCellXY(base: vec3<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return base + vec3<f32>(gsl_floored(i, 3.0) - 1.0, floor(i / 3.0) - 1.0, 0.0);
}

fn noise_driftingHashXY(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x + 211.0, cell.y + 5.0, -97.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x + 211.0, cell.y + 5.0, -97.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_driftingLife(time: f32, h: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> f32 {
  return fract(time / life + fract(dot(h, vec3<f32>(7.1, 3.7, 5.3))));
}

fn noise_driftingStart(time: f32, h: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> f32 {
  let grown = sin(3.1415927 * noise_driftingLife(time, h, life, noiseTex));
  return 3.0 * (1.0 - grown) * (1.0 - grown);
}

fn noise_driftingWall(q: vec3<f32>, a: vec3<f32>, b: vec3<f32>, startA: f32, startB: f32, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let toA = q - a;
  let lengthA = max(length(toA), 0.000001);
  let toB = q - b;
  let lengthB = max(length(toB), 0.000001);
  let gap = dot(b - a, a + b - 2.0 * q) / (lengthA + lengthB) + startB - startA;
  let u = toA / lengthA;
  let v = toB / lengthB;
  let across = cross(u, v);
  let apart = select(cross(across, u) - u * dot(across, across) / max(1.0 + dot(u, v), 0.5), v - u, dot(u, v) < 0.0);
  return gap / max(length(apart / size), 0.000001);
}

fn texture_reliefRoughness(roughness: f32, slope: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sqrt(roughness * roughness + (1.0 - texture_fadeDetail(size, noiseTex) * gsl_dispDrawn) * slope * slope);
}

fn noise_cellularEdgesZBlendLattice_gsl_anon(pi: vec2<f32>, zBand: f32, zBlendT: f32, pf: vec2<f32>, a: vec2<f32>, o2: vec2<f32>, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cellLo = vec3<f32>(pi.x + o2.x, pi.y + o2.y, zBand);
  let hLo = noise_cellularHash(cellLo, noiseTex);
  let cellHi = vec3<f32>(pi.x + o2.x, pi.y + o2.y, zBand + 1.0);
  let hHi = noise_cellularHash(cellHi, noiseTex);
  let h = hLo + (hHi - hLo) * ease_cubic(zBlendT, noiseTex);
  let v = pf - (o2 + vec2<f32>(h.x, h.y));
  let d = dot(v, v);
  return vec2<f32>(min(a.x, d), min(max(a.x, d), a.y));
}

fn noise_cellularEdgesZBlendLattice(qxy: vec2<f32>, zBand: f32, zBlendT: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let pi = floor(qxy);
  let pf = qxy - pi;
  var gsl_fold = vec2<f32>(64.0, 64.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = noise_cellularEdgesZBlendLattice_gsl_anon(pi, zBand, zBlendT, pf, gsl_fold, gsl_foldK_3[gsl_i], noiseTex);
  }
  let acc = gsl_fold;
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = sqrt(acc.x);
  gsl_result.v_2 = sqrt(acc.y);
  return gsl_result;
}

fn noise_cellularEdgesZBlend(p: vec3<f32>, size: f32, zBandSize: f32, noiseTex: texture_3d<f32>) -> f32 {
  let qxy = vec2<f32>(p.x / size, p.y / size);
  let zBand = floor(p.z / zBandSize);
  let zBlendT = fract(p.z / zBandSize);
  let gsl_result_5 = noise_cellularEdgesZBlendLattice(qxy, zBand, zBlendT, noiseTex);
  let f2 = gsl_result_5.v_2;
  let f1 = gsl_result_5.v_1;
  return (f2 - f1) * size;
}

fn texture_coverageLine(gsl_var_distance: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let half = max(0.5 * width, 0.5 * gsl_pw);
  return (1.0 - texture_mask(gsl_var_distance, half, 0.25 * half, noiseTex)) * 0.5 * width / half;
}

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let la = color_rgb2oklab(a, noiseTex);
  let lb = color_rgb2oklab(b, noiseTex);
  return color_oklab2rgb(la + (lb - la) * t, noiseTex);
}

fn material_hiddenShare__spec1(fade: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - fade * gsl_dispDrawn;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn material_withEmissive(m: mat4x4<f32>, rgb: vec3<f32>, strength: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], vec4<f32>(rgb.x, rgb.y, rgb.z, strength), m[2], m[3]);
}

fn material_translucent(m: mat4x4<f32>, transmission: f32, absorptionDistance: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(transmission, absorptionDistance, volume.z, volume.w), m[3]);
}

fn material_withTransmission(m: mat4x4<f32>, transmission: f32, absorptionDistance: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_translucent(m, transmission, absorptionDistance, noiseTex);
}

fn noise_gradientLatticeVector_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, e: vec3<f32>, acc: vec3<f32>, o: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let weight = (o.x * e.x + (1.0 - o.x) * (1.0 - e.x)) * (o.y * e.y + (1.0 - o.y) * (1.0 - e.y)) * (o.z * e.z + (1.0 - o.z) * (1.0 - e.z));
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + weight * vec3<f32>(dot(g, d), dot(vec3<f32>(g.y, g.z, g.x), d), dot(vec3<f32>(g.z, g.x, g.y), d));
}

fn noise_gradientLatticeVector(q: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let pi = floor(q);
  let pf = q - pi;
  let e = vec3<f32>(ease_cubic(pf.x, noiseTex), ease_cubic(pf.y, noiseTex), ease_cubic(pf.z, noiseTex));
  var gsl_fold = vec3<f32>(0.0, 0.0, 0.0);
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLatticeVector_gsl_anon(pi, pf, e, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn stone_gradientAt(q: vec3<f32>, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return noise_gradientLattice(q / size, noiseTex);
}

fn stone_warpAt(q: vec3<f32>, size: f32, strength: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return q + vec3<f32>(stone_gradientAt(q, size, noiseTex), stone_gradientAt(q + vec3<f32>(37.0, 0.0, 0.0), size, noiseTex), stone_gradientAt(q + vec3<f32>(0.0, 59.0, 0.0), size, noiseTex)) * strength;
}

fn stone_veinedMarble__spec1(p: vec3<f32>, q: vec3<f32>, foldSize: f32, foldDepth: f32, veinScale: f32, veinWidth: f32, sharpness: f32, crackShare: f32, facing: f32, clusterSize: f32, clusterLevel: f32, haloStrength: f32, fineStrength: f32, groundColor: vec3<f32>, cloudColor: vec3<f32>, veinColor: vec3<f32>, accentColor: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let groundLab = color_rgb2oklab(groundColor, noiseTex);
  let folded = stone_warpAt(q, foldSize, foldDepth, noiseTex);
  let cloud = texture_mask(stone_gradientAt(folded + vec3<f32>(19.0, 3.0, 47.0), veinScale * 1.5, noiseTex) + 0.5 * stone_gradientAt(folded + vec3<f32>(19.0, 3.0, 47.0), veinScale * 0.75, noiseTex), 0.15, 0.35, noiseTex);
  let zone = texture_mask(stone_gradientAt(q + vec3<f32>(5.0, 101.0, 37.0), clusterSize, noiseTex), clusterLevel, 0.06, noiseTex);
  let jagged = stone_warpAt(folded, veinScale * 0.06, veinScale * 0.02, noiseTex);
  let gsl_result = noise_cellularF1F2PairLattice(jagged / veinScale, noiseTex);
  let farCell = gsl_result.v_4;
  let farFeature = farCell + fract(textureLoad(noiseTex, vec3<i32>(floor(farCell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((farCell) / 64.0)) & vec3<i32>(63), 0).rgb);
  let nearCell = gsl_result.v_3;
  let nearFeature = nearCell + fract(textureLoad(noiseTex, vec3<i32>(floor(nearCell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((nearCell) / 64.0)) & vec3<i32>(63), 0).rgb);
  let wallNormal = normalize(farFeature - nearFeature);
  let sheetLike = texture_mask(abs(wallNormal.x), facing, 0.1, noiseTex);
  let crackNoise = stone_gradientAt((nearFeature + farFeature) * 0.5 * veinScale + vec3<f32>(3.0, 11.0, 7.0), veinScale * 1.5, noiseTex);
  let crackGate = sheetLike * texture_mask(crackNoise, (0.5 - crackShare) * 0.6, 0.04, noiseTex);
  let cracked = crackGate;
  let f2 = gsl_result.v_2;
  let f1 = gsl_result.v_1;
  let crackWall = (f2 - f1) * veinScale * 0.5;
  let wall = crackWall;
  let along = stone_gradientAt(q + vec3<f32>(71.0, 13.0, 29.0), veinScale * 0.7, noiseTex);
  let halfWidth = 0.5 * veinWidth * (0.2 + 1.3 * texture_mask(along, -0.1, 0.5, noiseTex));
  let halo = haloStrength * zone * cracked * (1.0 - texture_mask(wall, 2.5 * halfWidth, 1.5 * halfWidth, noiseTex));
  let grain = (fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.35) + vec3<f32>(9.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.35) + vec3<f32>(9.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.025 * (1.0 - texture_mask(gsl_pw, 0.13125, 0.04375, noiseTex));
  let ground = groundLab + (color_rgb2oklab(cloudColor, noiseTex) - groundLab) * clamp(0.5 * cloud + halo, 0.0, 1.0) + vec3<f32>(grain, 0.0, 0.0);
  let veinLab = color_rgb2oklab(veinColor, noiseTex);
  let tinted = veinLab + (color_rgb2oklab(accentColor, noiseTex) - veinLab) * texture_mask(along, 0.3, 0.25, noiseTex);
  let vein = zone * cracked * (1.0 - texture_mask(wall, halfWidth, max(halfWidth / sharpness, gsl_pw), noiseTex));
  let gsl_result_2 = noise_cellularF1F2PairLattice((jagged + vec3<f32>(31.0, 17.0, 5.0)) / (veinScale / 3.0), noiseTex);
  let farFine = gsl_result_2.v_4;
  let nearFine = gsl_result_2.v_3;
  let fineNormal = normalize(farFine + fract(textureLoad(noiseTex, vec3<i32>(floor(farFine)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((farFine) / 64.0)) & vec3<i32>(63), 0).rgb) - nearFine - fract(textureLoad(noiseTex, vec3<i32>(floor(nearFine)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((nearFine) / 64.0)) & vec3<i32>(63), 0).rgb));
  let fineCracked = (0.3 + 0.7 * zone) * texture_mask(abs(fineNormal.x), facing * 0.7, 0.1, noiseTex) * select(0.0, 1.0, fract(textureLoad(noiseTex, vec3<i32>(floor(nearFine + farFine + vec3<f32>(5.0, 2.0, 13.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((nearFine + farFine + vec3<f32>(5.0, 2.0, 13.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x < 0.5);
  let g2 = gsl_result_2.v_2;
  let g1 = gsl_result_2.v_1;
  let fineWall = (g2 - g1) * veinScale / 6.0;
  let fineHalf = 0.3 * halfWidth;
  let fine = fineStrength * fineCracked * (1.0 - texture_mask(fineWall, fineHalf, max(fineHalf / sharpness, gsl_pw), noiseTex));
  return color_oklab2rgb(ground + (tinted - ground) * max(vein, 0.7 * fine), noiseTex);
}

fn stone_grainOf(c: vec3<f32>, size: f32, s: f32, zone: f32, zoneWeight: f32, clump: f32, seed: vec3<f32>, scatter: f32, level: f32, noiseTex: texture_3d<f32>) -> f32 {
  let at = (c + fract(textureLoad(noiseTex, vec3<i32>(floor(c)) & vec3<i32>(63), 0).rgb)) * size;
  let cluster = zoneWeight * noise_gradientLattice(at / (zone * s) + seed, noiseTex) + 0.6 * noise_gradientLattice(at / (clump * s) + seed * 1.7, noiseTex);
  return select(0.0, 1.0, cluster - scatter * fract(textureLoad(noiseTex, vec3<i32>(floor(c + seed)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((c + seed) / 64.0)) & vec3<i32>(63), 0).rgb).x > level);
}

fn stone_quartzOf(c: vec3<f32>, g: f32, s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_grainOf(c, g, s, 4.0, 0.75, 1.2, vec3<f32>(7.3, 3.1, 11.7), 0.8, -0.31, noiseTex);
}

fn stone_maficOf(c: vec3<f32>, t: f32, s: f32, nearQuartz: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_grainOf(c, t, s, 4.0, 0.25, 0.6, vec3<f32>(13.1, 7.7, 2.3), 0.25, 0.17 - 0.15 * nearQuartz, noiseTex);
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
  let gsl_tmp_251 = poreBase + vec2<f32>(0.0, 0.0);
  let gsl_tmp_252 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_251.x, gsl_tmp_251.y, gsl_tmp_251.x - gsl_tmp_251.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_251.x, gsl_tmp_251.y, gsl_tmp_251.x - gsl_tmp_251.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_253 = floor(gsl_tmp_252);
  let gsl_tmp_254 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_251.x, gsl_tmp_251.y, gsl_tmp_253))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_251.x, gsl_tmp_251.y, gsl_tmp_253)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let sizeShare = wood_poreSizeShare(progress, ringPorosity, poreZone, noiseTex);
  let size = sizeShare.x;
  let poreShare = sizeShare.y;
  let gsl_tmp_255 = select(10.0, min(10.0, length(vec2<f32>(length(poreQ - (gsl_tmp_251 + gsl_tmp_254.xy)), max(abs(gsl_tmp_252 - gsl_tmp_253 - 0.5) - 0.2 - 0.25 * gsl_tmp_254.x, 0.0) * 8.0)) / size), gsl_tmp_254.z < poreShare);
  let gsl_tmp_256 = poreBase + vec2<f32>(1.0, 0.0);
  let gsl_tmp_257 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_256.x, gsl_tmp_256.y, gsl_tmp_256.x - gsl_tmp_256.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_256.x, gsl_tmp_256.y, gsl_tmp_256.x - gsl_tmp_256.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_258 = floor(gsl_tmp_257);
  let gsl_tmp_259 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_256.x, gsl_tmp_256.y, gsl_tmp_258))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_256.x, gsl_tmp_256.y, gsl_tmp_258)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_260 = select(gsl_tmp_255, min(gsl_tmp_255, length(vec2<f32>(length(poreQ - (gsl_tmp_256 + gsl_tmp_259.xy)), max(abs(gsl_tmp_257 - gsl_tmp_258 - 0.5) - 0.2 - 0.25 * gsl_tmp_259.x, 0.0) * 8.0)) / size), gsl_tmp_259.z < poreShare);
  let gsl_tmp_261 = poreBase + vec2<f32>(0.0, 1.0);
  let gsl_tmp_262 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_261.x, gsl_tmp_261.y, gsl_tmp_261.x - gsl_tmp_261.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_261.x, gsl_tmp_261.y, gsl_tmp_261.x - gsl_tmp_261.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_263 = floor(gsl_tmp_262);
  let gsl_tmp_264 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_261.x, gsl_tmp_261.y, gsl_tmp_263))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_261.x, gsl_tmp_261.y, gsl_tmp_263)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_265 = select(gsl_tmp_260, min(gsl_tmp_260, length(vec2<f32>(length(poreQ - (gsl_tmp_261 + gsl_tmp_264.xy)), max(abs(gsl_tmp_262 - gsl_tmp_263 - 0.5) - 0.2 - 0.25 * gsl_tmp_264.x, 0.0) * 8.0)) / size), gsl_tmp_264.z < poreShare);
  let gsl_tmp_266 = poreBase + vec2<f32>(1.0, 1.0);
  let gsl_tmp_267 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_266.x, gsl_tmp_266.y, gsl_tmp_266.x - gsl_tmp_266.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_266.x, gsl_tmp_266.y, gsl_tmp_266.x - gsl_tmp_266.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_268 = floor(gsl_tmp_267);
  let gsl_tmp_269 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_266.x, gsl_tmp_266.y, gsl_tmp_268))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_266.x, gsl_tmp_266.y, gsl_tmp_268)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let nearestPore = select(gsl_tmp_265, min(gsl_tmp_265, length(vec2<f32>(length(poreQ - (gsl_tmp_266 + gsl_tmp_269.xy)), max(abs(gsl_tmp_267 - gsl_tmp_268 - 0.5) - 0.2 - 0.25 * gsl_tmp_269.x, 0.0) * 8.0)) / size), gsl_tmp_269.z < poreShare);
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

fn animated_incandescentStrength(celsius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return exp(16.483099 - 23870.0 / (max(celsius, 0.0) + 273.15));
}

fn animated_incandescentColor(celsius: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(1.0, max(0.43 * (1.0 - exp(-(celsius - 560.0) / 280.0)), 0.0), 0.0);
}

fn playful_glitterShare(density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - exp(-4.18879 * density * density * density);
}

fn playful_glitterFlakes(f1: f32, flakeCellSize: f32, density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return texture_fadeTo(texture_coverageDisc(f1, density * flakeCellSize, noiseTex), playful_glitterShare(density, noiseTex), flakeCellSize, noiseTex);
}

fn gcl_closure_gsl_anon(x: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(0.2, 0.0, 0.0);
}

fn gcl_closure_driftingSeedXY_2(time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = gcl_closure_gsl_anon(born * size, noiseTex) / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingCellsLatticeXY_gsl_anon(base: vec3<f32>, flat: vec3<f32>, size: vec3<f32>, life: f32, time: f32, best: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = noise_driftingCellXY(base, i, noiseTex);
  let h = noise_driftingHashXY(cell, noiseTex);
  let arrival = length(flat - gcl_closure_driftingSeedXY_2(time, cell, h, size, life, noiseTex)) + noise_driftingStart(time, h, life, noiseTex);
  return select(best, vec2<f32>(arrival, i), arrival < best.x);
}

fn gcl_closure_driftingSeedXY(time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = gcl_closure_gsl_anon(born * size, noiseTex) / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingSeedXY_3(time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = gcl_closure_gsl_anon(born * size, noiseTex) / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingCellsLatticeXY_gsl_anon_2(first: vec3<f32>, flat: vec3<f32>, firstSeed: vec3<f32>, size: vec3<f32>, life: f32, time: f32, firstStart: f32, firstAge: f32, nearest: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = noise_driftingCellXY(first, i, noiseTex);
  let h = noise_driftingHashXY(cell, noiseTex);
  let toWall = noise_driftingWall(flat, firstSeed, gcl_closure_driftingSeedXY_3(time, cell, h, size, life, noiseTex), firstStart, noise_driftingStart(time, h, life, noiseTex), size, noiseTex);
  return select(nearest, vec2<f32>(toWall, min(firstAge, noise_driftingLife(time, h, life, noiseTex) * life)), i != 4.0 && toWall < nearest.x);
}

fn gcl_closure_driftingCellsLatticeXY(time: f32, q: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> gsl_struct_3 {
  let flat = vec3<f32>(q.x, q.y, 0.0);
  let base = floor(flat);
  var gsl_fold = vec2<f32>(1000.0, 0.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = gcl_closure_driftingCellsLatticeXY_gsl_anon(base, flat, size, life, time, gsl_fold, f32(gsl_i), noiseTex);
  }
  let owner = gsl_fold;
  let first = noise_driftingCellXY(base, owner.y, noiseTex);
  let firstHash = noise_driftingHashXY(first, noiseTex);
  let firstSeed = gcl_closure_driftingSeedXY(time, first, firstHash, size, life, noiseTex);
  let firstStart = noise_driftingStart(time, firstHash, life, noiseTex);
  let firstAge = noise_driftingLife(time, firstHash, life, noiseTex) * life;
  var gsl_fold_2 = vec2<f32>(1000.0, 0.0);
  for (var gsl_i_2 = 0; gsl_i_2 < 9; gsl_i_2 = gsl_i_2 + 1) {
    gsl_fold_2 = gcl_closure_driftingCellsLatticeXY_gsl_anon_2(first, flat, firstSeed, size, life, time, firstStart, firstAge, gsl_fold_2, f32(gsl_i_2), noiseTex);
  }
  let wall = gsl_fold_2;
  let drift = (firstSeed - vec3<f32>(first.x + 0.25 + 0.5 * firstHash.x, first.y + 0.25 + 0.5 * firstHash.y, 0.0)) * size;
  var gsl_result : gsl_struct_3;
  gsl_result.v_1 = vec4<f32>(wall.x, wall.y, firstAge, fract(dot(firstHash, vec3<f32>(3.1, 7.7, 5.9))));
  gsl_result.v_2 = drift;
  return gsl_result;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 10> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-362.5, -25.0, -25.0);
  let gsl_partHi = vec3<f32>(-312.5, 25.0, 25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_sphere(p - vec3<f32>(-337.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let marbleBall = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-287.5, -25.0, -25.0);
  let gsl_partHi_2 = vec3<f32>(-237.5, 25.0, 25.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(-262.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let graniteBall = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-212.5, -25.0, -25.0);
  let gsl_partHi_3 = vec3<f32>(-162.5, 25.0, 25.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_sphere(p - vec3<f32>(-187.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let oakBall = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-137.5, -25.0, -25.0);
  let gsl_partHi_4 = vec3<f32>(-87.5, 25.0, 25.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_sphere(p - vec3<f32>(-112.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let rustedBall = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(-62.5, -25.0, -25.0);
  let gsl_partHi_5 = vec3<f32>(-12.5, 25.0, 25.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = builtin_sphere(p - vec3<f32>(-37.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let patinaBall = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(12.5, -25.0, -25.0);
  let gsl_partHi_6 = vec3<f32>(62.5, 25.0, 25.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = builtin_sphere(p - vec3<f32>(37.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let glitterBall = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(87.5, -25.0, -25.0);
  let gsl_partHi_7 = vec3<f32>(137.5, 25.0, 25.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = builtin_sphere(p - vec3<f32>(112.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let checkerBall = gsl_part_7;
  let gsl_tmp_512 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010944857 * sin(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.01742606 * sin(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020284768 * sin(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.022745797 * sin(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023844866 * sin(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022270106 * sin(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.017706865 * sin(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.011648065 * sin(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
  let gsl_tmp_513 = 0.0 + 0.018241428 * cos(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_512) - 0.64525276 * time + 0.0) + 0.029043432 * cos(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_512) - 0.73128647 * time + 2.1) + 0.03380795 * cos(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_512) - 0.81732017 * time + 4.4) + 0.037909664 * cos(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_512) - 0.90335387 * time + 1.3) + 0.039741445 * cos(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_512) - 1.0324044 * time + 5.2) + 0.037116844 * cos(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_512) - 1.2044718 * time + 3.3) + 0.029511444 * cos(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_512) - 1.4625729 * time + 0.7) + 0.019413443 * cos(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_512) - 1.8067077 * time + 2.9);
  let waterHeight = gsl_tmp_513;
  let waterBall = (builtin_sphere(p - vec3<f32>(187.5, 0.0, 0.0), 25.0, noiseTex) - waterHeight) / 1.2928541;
  let gsl_partLo_8 = vec3<f32>(237.5, -25.0, -25.0);
  let gsl_partHi_8 = vec3<f32>(287.5, 25.0, 25.0);
  let gsl_partBox_8 = gsl_boxDistance(p, gsl_partLo_8, gsl_partHi_8);
  var gsl_part_8 = gsl_partBox_8;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_8 <= gsl_vanishView.w) {
    let gsl_partVanish_8 = gsl_vanish(gsl_partLo_8, gsl_partHi_8);
    gsl_part_8 = 1e10;
    if (gsl_partVanish_8.x > 0.0) { gsl_part_8 = builtin_sphere(p - vec3<f32>(262.5, 0.0, 0.0), 25.0, noiseTex) - 0.0 + gsl_partVanish_8.y; }
    if (abs(gsl_part_8) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_8), gsl_partVanish_8.zw); }
  }
  let lavaBall = gsl_part_8;
  let gsl_partLo_9 = vec3<f32>(312.5, -25.0, -25.0);
  let gsl_partHi_9 = vec3<f32>(362.5, 25.0, 25.0);
  let gsl_partBox_9 = gsl_boxDistance(p, gsl_partLo_9, gsl_partHi_9);
  var gsl_part_9 = gsl_partBox_9;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_9 <= gsl_vanishView.w) {
    let gsl_partVanish_9 = gsl_vanish(gsl_partLo_9, gsl_partHi_9);
    gsl_part_9 = 1e10;
    if (gsl_partVanish_9.x > 0.0) { gsl_part_9 = builtin_sphere(p - vec3<f32>(337.5, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_9.y; }
    if (abs(gsl_part_9) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_9), gsl_partVanish_9.zw); }
  }
  let skinBall = gsl_part_9;
  return array<f32, 10>(marbleBall, graniteBall, oakBall, rustedBall, patinaBall, glitterBall, checkerBall, waterBall, lavaBall, skinBall);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2], 3u);
  let gsl_s4 = gsl_shape(gsl_d[3], 4u);
  let gsl_s5 = gsl_shape(gsl_d[4], 5u);
  let gsl_s6 = gsl_shape(gsl_d[5], 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7], 8u);
  let gsl_s9 = gsl_shape(gsl_d[8], 9u);
  let gsl_s10 = gsl_shape(gsl_d[9], 10u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    var gsl_simpleValue = vec3<f32>(0.94624364, 0.9462436, 0.94624346);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      gsl_simpleValue = mix(gsl_simpleValue, stone_veinedMarble__spec1(p, p, 25.0, 10.0, 10.0, 0.4, 0.8, 0.45, -1.0, 30.0, -0.2, 0.25, 0.5, color_white, vec3<f32>(0.9, 0.9, 0.9), color_gray, color_gray, noiseTex), gsl_simpleDetail);
    }
    let marbleColor = gsl_simpleValue;
    return vec4<f32>(marbleColor, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    var gsl_detailValue = vec3<f32>(0.55051714, 0.5512089, 0.5521506);
    if (max(gsl_pw, gsl_detailPw) < 1.2) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_1 = p + (noise_gradientLatticeVector(p / 5.2, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 5.2 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.975) / 0.325 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.056;
      let gsl_tmp_2 = gsl_tmp_1 / 4.0;
      let gsl_tmp_5 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).nearCell;
      let gsl_tmp_25 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_5 + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_5 + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_29 = noise_gradientLattice(p / 1.2 + vec3<f32>(1.7, 9.1, 4.4), noiseTex) + 0.15 * (1.0 - texture_mask(gsl_pw, 0.18, 0.06, noiseTex)) * noise_gradientLattice(p / 0.48 + vec3<f32>(8.2, 2.6, 7.3), noiseTex) + 0.4 * noise_gradientLattice(p / 3.6 + vec3<f32>(4.2, 6.6, 1.3), noiseTex);
      let gsl_tmp_27 = normalize(gsl_tmp_25 - 0.5 + vec3<f32>(0.0001, 0.0, 0.0));
      let gsl_tmp_28 = dot(p, gsl_tmp_27);
      let gsl_tmp_4 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f2;
      let gsl_tmp_3 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f1;
      let gsl_tmp_6 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).farCell;
      let gsl_tmp_31 = 1.0 - texture_mask((gsl_tmp_4 * gsl_tmp_4 - gsl_tmp_3 * gsl_tmp_3) / (2.0 * max(length(gsl_tmp_6 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_6)) & vec3<i32>(63), 0).rgb) - gsl_tmp_5 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_5)) & vec3<i32>(63), 0).rgb)), 0.001)) * 4.0, 0.24, 0.2, noiseTex);
      let gsl_tmp_30 = texture_fadeTo(abs(noise_gradientLattice(p / 0.6 + vec3<f32>(6.1, 4.9, 2.2), noiseTex)), 0.157, 0.6, noiseTex) + 0.5 * texture_fadeTo(abs(noise_gradientLattice(p / 0.24 + vec3<f32>(1.1, 3.5, 8.8), noiseTex)), 0.157, 0.24, noiseTex);
      let gsl_tmp_32 = vec3<f32>(0.5981807, 0.000000000048424487, 0.000000022296533) + vec3<f32>((gsl_tmp_25.y - 0.5) * 0.1 + 0.1 * gsl_tmp_29 + 0.015 * noise_gradientLattice((p - gsl_tmp_27 * gsl_tmp_28) / 0.32 + gsl_tmp_27 * gsl_tmp_28 / 3.2, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.12, 0.04, noiseTex)) + 0.04 * gsl_tmp_31 + 0.12 * (gsl_tmp_30 - 0.235) * (0.5 + gsl_tmp_31), 0.0, 0.0) - vec3<f32>(0.0, 0.000000000048424487, 0.000000022296533) * 0.3 * gsl_tmp_31;
      let gsl_tmp_7 = (gsl_tmp_1 + p + (noise_gradientLatticeVector(p / 2.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 2.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 0.384 - p + vec3<f32>(148.4, 45.2, 23.6)) / 1.8;
      let gsl_tmp_10 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).nearCell;
      let gsl_tmp_13 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_10 + vec3<f32>(7.0, 29.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_10 + vec3<f32>(7.0, 29.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_15 = stone_quartzOf(gsl_tmp_10, 1.8, 4.0, noiseTex);
      let gsl_tmp_24 = select(0.0, 1.0 - gsl_tmp_15, gsl_tmp_13.x < 0.15);
      let gsl_tmp_34 = gsl_tmp_32 + (vec3<f32>(0.5981807, 0.000000000048424487, 0.000000022296533) + vec3<f32>(0.06 + (gsl_tmp_13.y - 0.5) * 0.06 + 0.04 * gsl_tmp_29 + 0.12 * (gsl_tmp_30 - 0.235), -0.000000000024212243, -0.000000011148266) - gsl_tmp_32) * gsl_tmp_24;
      let gsl_tmp_8 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).f1;
      let gsl_tmp_33 = 1.0 - texture_mask(gsl_tmp_8, 0.3, 0.25, noiseTex);
      let gsl_tmp_9 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).f2;
      let gsl_tmp_11 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).farCell;
      let gsl_tmp_12 = (gsl_tmp_9 * gsl_tmp_9 - gsl_tmp_8 * gsl_tmp_8) / (2.0 * max(length(gsl_tmp_11 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_11)) & vec3<i32>(63), 0).rgb) - gsl_tmp_10 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_10)) & vec3<i32>(63), 0).rgb)), 0.001));
      let gsl_tmp_14 = 2.0 * texture_mask(gsl_tmp_12 * 1.8, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_22 = gsl_tmp_15 * (0.5 + 0.5 * gsl_tmp_14) + stone_quartzOf(gsl_tmp_11, 1.8, 4.0, noiseTex) * (0.5 - 0.5 * gsl_tmp_14);
      let gsl_tmp_35 = gsl_tmp_34 + (vec3<f32>(0.775, 0.00000000006476233, 0.000000029819127) + vec3<f32>((gsl_tmp_13.z - 0.5) * 0.12 - 0.1 * gsl_tmp_33 + 0.05 * dot(gsl_tmp_7 - gsl_tmp_10 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_10)) & vec3<i32>(63), 0).rgb), normalize(gsl_tmp_13 - 0.5 + vec3<f32>(0.0, 0.0001, 0.0))) + 0.01 * (1.0 - texture_mask(gsl_tmp_12 * 1.8, 0.08, 0.06, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex)) + 0.02 * gsl_tmp_29 + 0.06 * (gsl_tmp_30 - 0.235), -0.003 * gsl_tmp_33, -0.01 * gsl_tmp_33 + (gsl_tmp_13.y - 0.5) * 0.02) - gsl_tmp_34) * gsl_tmp_22 + vec3<f32>(0.15 * noise_gradientLattice(p / 12.0 + vec3<f32>(2.9, 8.3, 6.1), noiseTex) + 0.08 * noise_gradientLattice(p / 5.6 + vec3<f32>(9.4, 1.2, 3.8), noiseTex), 0.0, 0.0);
      let gsl_tmp_16 = (gsl_tmp_1 + vec3<f32>(21.2, 246.8, 68.4)) / 0.6;
      let gsl_tmp_19 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).nearCell;
      let gsl_tmp_26 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_19 + vec3<f32>(3.0, 0.0, 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_19 + vec3<f32>(3.0, 0.0, 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_18 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).f2;
      let gsl_tmp_17 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).f1;
      let gsl_tmp_20 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).farCell;
      let gsl_tmp_21 = 2.0 * texture_mask((gsl_tmp_18 * gsl_tmp_18 - gsl_tmp_17 * gsl_tmp_17) / (2.0 * max(length(gsl_tmp_20 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_20)) & vec3<i32>(63), 0).rgb) - gsl_tmp_19 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_19)) & vec3<i32>(63), 0).rgb)), 0.001)) * 0.6, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_23 = stone_maficOf(gsl_tmp_19, 0.6, 4.0, gsl_tmp_15, noiseTex) * (0.5 + 0.5 * gsl_tmp_21) + stone_maficOf(gsl_tmp_20, 0.6, 4.0, gsl_tmp_15, noiseTex) * (0.5 - 0.5 * gsl_tmp_21);
      gsl_detailValue = color_oklab2rgb(vec3<f32>(0.6418677, -0.000099999954, -0.00039997892) + (gsl_tmp_35 + (vec3<f32>(0.30524173, 0.0, 0.0) + vec3<f32>((gsl_tmp_26.y - 0.5) * 0.06 + select(0.0, 0.12, gsl_tmp_26.z > 0.96), 0.0, 0.0) + select(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.03, 0.01, 0.02), gsl_tmp_26.z < 0.08) - gsl_tmp_35) * gsl_tmp_23 - vec3<f32>(0.6418677, -0.000099999954, -0.00039997892)) * (1.0 - texture_mask(gsl_pw, 0.9, 0.3, noiseTex)), noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue_2 = vec3<f32>(0.55051714, 0.5512089, 0.5521506);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, gsl_detailValue, gsl_simpleDetail);
    }
    let graniteColor = gsl_simpleValue_2;
    return vec4<f32>(graniteColor, 1.0);
  }
  if (gsl_colorIdx == 3u) {
    let gsl_tmp_664 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
    let gsl_tmp_665 = length(gsl_tmp_664.xy);
    let gsl_tmp_666 = gsl_tmp_664.xy / max(gsl_tmp_665, 0.001) * 40.0;
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let gsl_tmp_637 = vec2<f32>(gsl_tmp_665 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_666.x, gsl_tmp_666.y, gsl_tmp_664.z * 0.25), color_white, color_gray, noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_664.x, gsl_tmp_664.y, gsl_tmp_664.z * 0.2), color_white, color_gray, noiseTex), 0.0);
    let gsl_tmp_638 = gsl_tmp_637.x;
    let gsl_tmp_639 = floor(gsl_tmp_638);
    let gsl_tmp_640 = wood_ringBoundary(gsl_tmp_639, 3.0, noiseTex);
    let gsl_tmp_641 = wood_ringBoundary(gsl_tmp_639 + 1.0, 3.0, noiseTex);
    let gsl_tmp_643 = gsl_tmp_638 >= gsl_tmp_641;
    let gsl_tmp_642 = gsl_tmp_638 < gsl_tmp_640;
    let gsl_tmp_644 = select(select(gsl_tmp_640, gsl_tmp_641, gsl_tmp_643), wood_ringBoundary(gsl_tmp_639 - 1.0, 3.0, noiseTex), gsl_tmp_642);
    let gsl_tmp_646 = (gsl_tmp_638 - gsl_tmp_644) / (select(select(gsl_tmp_641, wood_ringBoundary(gsl_tmp_639 + 2.0, 3.0, noiseTex), gsl_tmp_643), gsl_tmp_640, gsl_tmp_642) - gsl_tmp_644);
    let gsl_tmp_645 = select(select(gsl_tmp_639, gsl_tmp_639 + 1.0, gsl_tmp_643), gsl_tmp_639 - 1.0, gsl_tmp_642);
    let gsl_tmp_647 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_645, gsl_tmp_645 + 11.0, 3.0 * gsl_tmp_645 + 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_645, gsl_tmp_645 + 11.0, 3.0 * gsl_tmp_645 + 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let gsl_tmp_648 = texture_mask(gsl_tmp_646, 0.4 + (gsl_tmp_647.x - 0.5) * 0.027, 0.3, noiseTex) * (1.0 + (gsl_tmp_647.y - 0.5) * 0.18);
    let gsl_tmp_635 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
    let gsl_tmp_636 = length(gsl_tmp_635.xy);
    var gsl_detailValue_2 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.075) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_649 = (atan2(gsl_tmp_635.y, gsl_tmp_635.x) + 3.1415927) / 6.2831855 * 3351.0 - 0.5;
      let gsl_tmp_650 = floor(gsl_tmp_649);
      let gsl_tmp_652 = wood_fibreFile(gsl_floored(gsl_tmp_650, 3351.0), gsl_tmp_636, gsl_tmp_635.z, noiseTex);
      let gsl_tmp_651 = gsl_tmp_649 - gsl_tmp_650;
      gsl_detailValue_2 = (gsl_tmp_652 + (wood_fibreFile(gsl_floored((gsl_tmp_650 + 1.0), 3351.0), gsl_tmp_636, gsl_tmp_635.z, noiseTex) - gsl_tmp_652) * gsl_tmp_651 * gsl_tmp_651 * (3.0 - 2.0 * gsl_tmp_651) - 0.5) * 0.08 * (1.0 - 0.6 * gsl_tmp_648) * texture_fadeDetail(min(0.15, 6.2831855 * gsl_tmp_636 / 3351.0), noiseTex);
      gsl_pw = gsl_detailOuterPw_2;
    }
    let gsl_tmp_655 = wood_poreSizeShare__spec1(gsl_tmp_646, 0.3, noiseTex).x;
    var gsl_detailValue_3 = 0.05 * gsl_tmp_655 * gsl_tmp_655;
    if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_3 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_653 = wood_poreDistanceAt(gsl_tmp_635, gsl_tmp_646, 0.5, 1.0, 0.3, noiseTex);
      let gsl_tmp_654 = texture_fadeDetail(0.28 * gsl_tmp_653.y, noiseTex);
      gsl_detailValue_3 = (1.0 - texture_mask(gsl_tmp_653.x, 0.28, 0.06, noiseTex)) * gsl_tmp_654 + 0.05 * gsl_tmp_653.y * gsl_tmp_653.y * (1.0 - gsl_tmp_654);
      gsl_pw = gsl_detailOuterPw_3;
    }
    var gsl_simpleValue_3 = vec3<f32>(0.5854705, 0.48488966, 0.378912);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_656 = (atan2(gsl_tmp_635.y, gsl_tmp_635.x) + 3.1415927) / 6.2831855 * 250.0;
      let gsl_tmp_657 = gsl_floored(floor(gsl_tmp_656), 250.0);
      let gsl_tmp_658 = gsl_tmp_635.z / 12.0 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_657, gsl_tmp_657 + 17.0, 3.0 * gsl_tmp_657 + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_657, gsl_tmp_657 + 17.0, 3.0 * gsl_tmp_657 + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
      let gsl_tmp_659 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_657, floor(gsl_tmp_658), gsl_tmp_657 + 401.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_657, floor(gsl_tmp_658), gsl_tmp_657 + 401.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_660 = vec3<f32>(0.62887883, 0.019482663, 0.048678815) + vec3<f32>(-0.044444397, -0.0012675361, -0.0034771787) * gsl_tmp_648 + vec3<f32>((fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_645, gsl_tmp_645 + 5.0, 2.0 * gsl_tmp_645 + 29.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_645, gsl_tmp_645 + 5.0, 2.0 * gsl_tmp_645 + 29.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.08 + gsl_detailValue_2 + 0.1 * 1.0 * select(0.0, 1.0, gsl_tmp_659.z < 0.5) * (1.0 - texture_mask(abs(gsl_tmp_656 - floor(gsl_tmp_656) - 0.2 - 0.6 * gsl_tmp_659.x) * 6.2831855 * gsl_tmp_636 / 250.0, 0.15, 0.05, noiseTex)) * (1.0 - texture_mask(abs(gsl_tmp_658 - floor(gsl_tmp_658) - 0.5) * 2.0 / (0.35 + 0.65 * gsl_tmp_659.y), 1.0, 0.1, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex)) + 0.0, 0.0, 0.0) + vec3<f32>(0.042521644, 0.0007117121, -0.0038856908) * texture_mask(gsl_tmp_637.x, 999979.3, 0.3, noiseTex);
      let gsl_tmp_661 = gsl_tmp_660 + (vec3<f32>(0.0, 0.0, 0.0) - gsl_tmp_660) * 0.0 * (1.0 - texture_mask(abs(gsl_tmp_637.x - 999979.3 + 0.3), 0.3, 0.1, noiseTex));
      let gsl_tmp_662 = gsl_tmp_661 + (vec3<f32>(0.0, 0.0, 0.0) - gsl_tmp_661) * 0.0;
      let gsl_tmp_663 = gsl_tmp_662 + (vec3<f32>(0.4547599, 0.03102755, 0.056627084) - gsl_tmp_662) * 0.85 * gsl_detailValue_3 - vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_376 = color_oklab2rgb(gsl_tmp_663 + (vec3<f32>(0.41371727, 0.012464928, 0.032169826) - gsl_tmp_663) * 0.8 * gsl_tmp_637.y, noiseTex);
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, gsl_tmp_376, gsl_simpleDetail);
    }
    let oakColor = gsl_simpleValue_3;
    return vec4<f32>(oakColor, 1.0);
  }
  if (gsl_colorIdx == 4u) {
    var gsl_simpleValue_4 = vec3<f32>(0.6824685, 0.58169425, 0.44390574);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_377 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_378 = vec2<f32>(gsl_tmp_377.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_377.y + 0.5);
      let gsl_tmp_379 = vec2<f32>(gsl_tmp_378.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_378.y + 0.25);
      let gsl_tmp_380 = vec2<f32>(gsl_tmp_379.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_379.y + 0.0);
      let gsl_tmp_381 = vec2<f32>(gsl_tmp_380.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_380.y + 0.0);
      let gsl_tmp_382 = vec2<f32>(gsl_tmp_381.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_381.y + 0.0);
      let gsl_tmp_37 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_382.x / max(gsl_tmp_382.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_38 = 0.35 * texture_mask(gsl_tmp_37, -0.3, 0.3, noiseTex);
      let gsl_tmp_398 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_399 = vec2<f32>(gsl_tmp_398.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_398.y + 0.6);
      let gsl_tmp_400 = vec2<f32>(gsl_tmp_399.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_399.y + 0.36);
      let gsl_tmp_401 = vec2<f32>(gsl_tmp_400.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_400.y + 0.216);
      let gsl_tmp_402 = vec2<f32>(gsl_tmp_401.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_401.y + 0.1296);
      let gsl_tmp_403 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_404 = vec2<f32>(gsl_tmp_403.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_403.y + 0.6);
      let gsl_tmp_405 = vec2<f32>(gsl_tmp_404.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_404.y + 0.36);
      let gsl_tmp_406 = vec2<f32>(gsl_tmp_405.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_405.y + 0.216);
      let gsl_tmp_407 = vec2<f32>(gsl_tmp_406.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_406.y + 0.1296);
      let gsl_tmp_408 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_409 = vec2<f32>(gsl_tmp_408.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_408.y + 0.5);
      let gsl_tmp_410 = vec2<f32>(gsl_tmp_409.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_409.y + 0.25);
      let gsl_tmp_411 = vec2<f32>(gsl_tmp_410.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_410.y + 0.0);
      let gsl_tmp_412 = vec2<f32>(gsl_tmp_411.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_411.y + 0.0);
      let gsl_tmp_413 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_414 = vec2<f32>(gsl_tmp_413.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_413.y + 0.5);
      let gsl_tmp_415 = vec2<f32>(gsl_tmp_414.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_414.y + 0.25);
      let gsl_tmp_416 = vec2<f32>(gsl_tmp_415.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_415.y + 0.0);
      let gsl_tmp_417 = vec2<f32>(gsl_tmp_416.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_416.y + 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      let gsl_tmp_384 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), 1.0);
      let gsl_tmp_385 = vec2<f32>(gsl_tmp_384.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_384.y + 0.5);
      let gsl_tmp_386 = vec2<f32>(gsl_tmp_385.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_385.y + 0.25);
      let gsl_tmp_387 = vec2<f32>(gsl_tmp_386.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_386.y + 0.0);
      let gsl_tmp_388 = vec2<f32>(gsl_tmp_387.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_387.y + 0.0);
      let gsl_tmp_389 = vec2<f32>(gsl_tmp_388.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_388.y + 0.0);
      let gsl_tmp_39 = texture_fadeTo(gsl_remap_3(p + (noise_gradientLatticeVector(p / 4.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 4.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.75) / 0.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.44, color_white, color_gray, noiseTex) * texture_mask(2.0 * gsl_tmp_389.x / max(gsl_tmp_389.y, 0.000001), -0.15, 0.2, noiseTex), 0.159, 3.0, noiseTex) - 0.159;
      let gsl_tmp_418 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_419 = vec2<f32>(gsl_tmp_418.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_418.y + 0.6);
      let gsl_tmp_420 = vec2<f32>(gsl_tmp_419.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_419.y + 0.36);
      let gsl_tmp_421 = vec2<f32>(gsl_tmp_420.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_420.y + 0.216);
      let gsl_tmp_422 = vec2<f32>(gsl_tmp_421.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_421.y + 0.1296);
      let gsl_tmp_423 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_424 = vec2<f32>(gsl_tmp_423.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_423.y + 0.6);
      let gsl_tmp_425 = vec2<f32>(gsl_tmp_424.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_424.y + 0.36);
      let gsl_tmp_426 = vec2<f32>(gsl_tmp_425.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_425.y + 0.216);
      let gsl_tmp_427 = vec2<f32>(gsl_tmp_426.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_426.y + 0.1296);
      let gsl_tmp_428 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_429 = vec2<f32>(gsl_tmp_428.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_428.y + 0.5);
      let gsl_tmp_430 = vec2<f32>(gsl_tmp_429.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_429.y + 0.25);
      let gsl_tmp_431 = vec2<f32>(gsl_tmp_430.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_430.y + 0.0);
      let gsl_tmp_432 = vec2<f32>(gsl_tmp_431.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_431.y + 0.0);
      let gsl_tmp_433 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_434 = vec2<f32>(gsl_tmp_433.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_433.y + 0.5);
      let gsl_tmp_435 = vec2<f32>(gsl_tmp_434.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_434.y + 0.25);
      let gsl_tmp_436 = vec2<f32>(gsl_tmp_435.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_435.y + 0.0);
      let gsl_tmp_437 = vec2<f32>(gsl_tmp_436.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_436.y + 0.0);
      let gsl_tmp_438 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_439 = vec2<f32>(gsl_tmp_438.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_438.y + 0.6);
      let gsl_tmp_440 = vec2<f32>(gsl_tmp_439.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_439.y + 0.36);
      let gsl_tmp_441 = vec2<f32>(gsl_tmp_440.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_440.y + 0.216);
      let gsl_tmp_442 = vec2<f32>(gsl_tmp_441.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_441.y + 0.1296);
      let gsl_tmp_443 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_444 = vec2<f32>(gsl_tmp_443.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_443.y + 0.6);
      let gsl_tmp_445 = vec2<f32>(gsl_tmp_444.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_444.y + 0.36);
      let gsl_tmp_446 = vec2<f32>(gsl_tmp_445.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_445.y + 0.216);
      let gsl_tmp_447 = vec2<f32>(gsl_tmp_446.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_446.y + 0.1296);
      let gsl_tmp_448 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_449 = vec2<f32>(gsl_tmp_448.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_448.y + 0.5);
      let gsl_tmp_450 = vec2<f32>(gsl_tmp_449.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_449.y + 0.25);
      let gsl_tmp_451 = vec2<f32>(gsl_tmp_450.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_450.y + 0.0);
      let gsl_tmp_452 = vec2<f32>(gsl_tmp_451.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_451.y + 0.0);
      let gsl_tmp_453 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_454 = vec2<f32>(gsl_tmp_453.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_453.y + 0.5);
      let gsl_tmp_455 = vec2<f32>(gsl_tmp_454.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_454.y + 0.25);
      let gsl_tmp_456 = vec2<f32>(gsl_tmp_455.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_455.y + 0.0);
      let gsl_tmp_457 = vec2<f32>(gsl_tmp_456.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_456.y + 0.0);
      gsl_simpleValue_4 = mix(gsl_simpleValue_4, texture_paint(gsl_tmp_37, 0.0, 0.08, color_mixOklab(vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.6585775, 0.47784516, 0.14561833), gsl_tmp_38, noiseTex), color_mixOklab(vec3<f32>(0.6585775, 0.47784516, 0.14561833), color_oklab2rgb(vec3<f32>(0.4104389, 0.04798136, 0.060527567) + vec3<f32>(0.09 * (2.0 * (gsl_tmp_402.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_407.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_412.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_417.y + 0.0, 0.000001)) - 0.11 * gsl_tmp_39, 0.006 * (2.0 * (gsl_tmp_422.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_427.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_432.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_437.y + 0.0, 0.000001)) - 0.01 * gsl_tmp_39, 0.025 * (2.0 * (gsl_tmp_442.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_447.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_452.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_457.y + 0.0, 0.000001)) - 0.02 * gsl_tmp_39), noiseTex), texture_mask(gsl_tmp_37, 0.25, 0.25, noiseTex), noiseTex), noiseTex), gsl_simpleDetail);
    }
    let rustedColor = gsl_simpleValue_4;
    return vec4<f32>(rustedColor, 1.0);
  }
  if (gsl_colorIdx == 5u) {
    var gsl_simpleValue_5 = vec3<f32>(0.7927733, 0.85689414, 0.7547779);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_474 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_475 = vec2<f32>(gsl_tmp_474.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_474.y + 0.5);
      let gsl_tmp_476 = vec2<f32>(gsl_tmp_475.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_475.y + 0.25);
      let gsl_tmp_477 = vec2<f32>(gsl_tmp_476.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_476.y + 0.0);
      let gsl_tmp_478 = vec2<f32>(gsl_tmp_477.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_477.y + 0.0);
      let gsl_tmp_479 = vec2<f32>(gsl_tmp_478.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_478.y + 0.0);
      let gsl_tmp_47 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_479.x / max(gsl_tmp_479.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_48 = 0.35 * texture_mask(gsl_tmp_47, -0.3, 0.3, noiseTex);
      let gsl_tmp_602 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_603 = vec2<f32>(gsl_tmp_602.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_602.y + 0.6);
      let gsl_tmp_604 = vec2<f32>(gsl_tmp_603.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_603.y + 0.36);
      let gsl_tmp_605 = vec2<f32>(gsl_tmp_604.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_604.y + 0.216);
      let gsl_tmp_606 = vec2<f32>(gsl_tmp_605.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_605.y + 0.1296);
      let gsl_tmp_607 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_608 = vec2<f32>(gsl_tmp_607.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_607.y + 0.6);
      let gsl_tmp_609 = vec2<f32>(gsl_tmp_608.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_608.y + 0.36);
      let gsl_tmp_610 = vec2<f32>(gsl_tmp_609.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_609.y + 0.216);
      let gsl_tmp_611 = vec2<f32>(gsl_tmp_610.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_610.y + 0.1296);
      let gsl_tmp_612 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_613 = vec2<f32>(gsl_tmp_612.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_612.y + 0.5);
      let gsl_tmp_614 = vec2<f32>(gsl_tmp_613.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_613.y + 0.25);
      let gsl_tmp_615 = vec2<f32>(gsl_tmp_614.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_614.y + 0.0);
      let gsl_tmp_616 = vec2<f32>(gsl_tmp_615.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_615.y + 0.0);
      let gsl_tmp_617 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_618 = vec2<f32>(gsl_tmp_617.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_617.y + 0.5);
      let gsl_tmp_619 = vec2<f32>(gsl_tmp_618.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_618.y + 0.25);
      let gsl_tmp_620 = vec2<f32>(gsl_tmp_619.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_619.y + 0.0);
      let gsl_tmp_621 = vec2<f32>(gsl_tmp_620.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_620.y + 0.0);
      let gsl_tmp_622 = 2.0 * (gsl_tmp_606.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_611.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_616.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_621.y + 0.0, 0.000001);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      let gsl_tmp_624 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), 1.0);
      let gsl_tmp_625 = vec2<f32>(gsl_tmp_624.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_624.y + 0.5);
      let gsl_tmp_626 = vec2<f32>(gsl_tmp_625.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_625.y + 0.25);
      let gsl_tmp_627 = vec2<f32>(gsl_tmp_626.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_626.y + 0.0);
      let gsl_tmp_628 = vec2<f32>(gsl_tmp_627.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_627.y + 0.0);
      let gsl_tmp_629 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), 1.0);
      let gsl_tmp_630 = vec2<f32>(gsl_tmp_629.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_629.y + 0.5);
      let gsl_tmp_631 = vec2<f32>(gsl_tmp_630.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_630.y + 0.25);
      let gsl_tmp_632 = vec2<f32>(gsl_tmp_631.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_631.y + 0.0);
      let gsl_tmp_633 = vec2<f32>(gsl_tmp_632.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_632.y + 0.0);
      let gsl_tmp_634 = texture_fadeTo(gsl_remap_4(p + (noise_gradientLatticeVector(p / 4.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 4.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.75) / 0.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.44, color_white, color_gray, noiseTex) * texture_mask(2.0 * (gsl_tmp_628.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex)) / max(gsl_tmp_633.y + 0.0, 0.000001), -0.15, 0.2, noiseTex), 0.159, 3.0, noiseTex) - 0.159;
      gsl_simpleValue_5 = mix(gsl_simpleValue_5, texture_paint(gsl_tmp_47, 0.0, 0.08, color_mixOklab(vec3<f32>(0.9695, 0.8112, 0.7497), vec3<f32>(0.67572874, 0.9464596, 0.7818527), gsl_tmp_48, noiseTex), color_mixOklab(vec3<f32>(0.67572874, 0.9464596, 0.7818527), color_oklab2rgb(vec3<f32>(0.702035, -0.058106415, -0.015318653) + vec3<f32>(0.09 * gsl_tmp_622 - 0.11 * gsl_tmp_634, 0.006 * gsl_tmp_622 - 0.01 * gsl_tmp_634, 0.025 * gsl_tmp_622 - 0.02 * gsl_tmp_634), noiseTex), texture_mask(gsl_tmp_47, 0.25, 0.25, noiseTex), noiseTex), noiseTex), gsl_simpleDetail);
    }
    let patinaColor = gsl_simpleValue_5;
    return vec4<f32>(patinaColor, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    var gsl_simpleValue_6 = vec3<f32>(0.0032853691, 0.003285369, 0.003285368);
    if (gsl_simpleDetail > 0.0) {
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let gsl_tmp_49 = p / 1.0;
      let gsl_tmp_50 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_49, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      gsl_simpleValue_6 = mix(gsl_simpleValue_6, color_mixOklab(color_black, color_white, gsl_tmp_50, noiseTex), gsl_simpleDetail);
    }
    let glitterColor = gsl_simpleValue_6;
    return vec4<f32>(glitterColor, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    var gsl_simpleValue_7 = vec3<f32>(0.38857287, 0.38857284, 0.3885728);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_82 = texture_squareWave(p.x, 20.0, noiseTex);
      let gsl_tmp_83 = texture_squareWave(p.y, 20.0, noiseTex);
      let gsl_tmp_85 = gsl_tmp_82 + gsl_tmp_83 - 2.0 * gsl_tmp_82 * gsl_tmp_83;
      let gsl_tmp_84 = texture_squareWave(p.z, 20.0, noiseTex);
      gsl_simpleValue_7 = mix(gsl_simpleValue_7, color_mixOklab(color_white, color_black, gsl_tmp_85 + gsl_tmp_84 - 2.0 * gsl_tmp_85 * gsl_tmp_84, noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_7, 1.0);
  }
  if (gsl_colorIdx == 8u) {
    var gsl_simpleValue_8 = vec3<f32>(0.048666414, 0.1840356, 0.30572936);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_512 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010944857 * sin(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.01742606 * sin(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020284768 * sin(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.022745797 * sin(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023844866 * sin(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022270106 * sin(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.017706865 * sin(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.011648065 * sin(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
      let gsl_tmp_513 = 0.0 + 0.018241428 * cos(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_512) - 0.64525276 * time + 0.0) + 0.029043432 * cos(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_512) - 0.73128647 * time + 2.1) + 0.03380795 * cos(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_512) - 0.81732017 * time + 4.4) + 0.037909664 * cos(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_512) - 0.90335387 * time + 1.3) + 0.039741445 * cos(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_512) - 1.0324044 * time + 5.2) + 0.037116844 * cos(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_512) - 1.2044718 * time + 3.3) + 0.029511444 * cos(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_512) - 1.4625729 * time + 0.7) + 0.019413443 * cos(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_512) - 1.8067077 * time + 2.9);
      let gsl_tmp_514 = texture_mask(gsl_tmp_513 + 0.037855294 * texture_fadeTo(noise_gradientLattice((p - vec3<f32>(0.95782626, 0.28734788, 0.0) * 1.3692689 * time) / 0.5 + vec3<f32>(13.0, 71.0, 5.0), noiseTex), 0.0, 0.5, noiseTex), 0.14711061, 0.018927647, noiseTex);
      gsl_simpleValue_8 = mix(gsl_simpleValue_8, color_mixOklab(color_mixOklab(vec3<f32>(0.04, 0.16, 0.29), vec3<f32>(0.1, 0.42, 0.44), 0.5 * texture_mask(gsl_tmp_513, 0.06309216, 0.06309216, noiseTex), noiseTex), vec3<f32>(0.92, 0.94, 0.94), gsl_tmp_514, noiseTex), gsl_simpleDetail);
    }
    let waterColor = gsl_simpleValue_8;
    return vec4<f32>(waterColor, 1.0);
  }
  if (gsl_colorIdx == 9u) {
    var gsl_simpleValue_9 = vec3<f32>(0.079512164, 0.072891675, 0.069725566);
    if (gsl_simpleDetail > 0.0) {
      let gsl_result_4 = gcl_closure_driftingCellsLatticeXY(time, p / 15.0, vec3<f32>(15.0, 15.0, 15.0), 20.0, noiseTex);
      let cells = gsl_result_4.v_1;
      let drift = gsl_result_4.v_2;
      gsl_simpleValue_9 = mix(gsl_simpleValue_9, color_mixOklab(vec3<f32>(0.05, 0.045, 0.045), vec3<f32>(0.14, 0.13, 0.12), texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let lavaColor = gsl_simpleValue_9;
    return vec4<f32>(lavaColor, 1.0);
  }
  if (gsl_colorIdx == 10u) {
    var gsl_simpleValue_10 = vec3<f32>(0.8056361, 0.8831174, 0.8809763);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_teal = vec3<f32>(0.0, 0.5, 0.5);
      gsl_simpleValue_10 = mix(gsl_simpleValue_10, color_mixOklab(color_white, color_teal, texture_fadeTo(texture_coverageLine(noise_cellularEdgesZBlend(p, 5.0, 2.5, noiseTex), 1.0, noiseTex), 0.228, 5.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let skinColor = gsl_simpleValue_10;
    return vec4<f32>(skinColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.15, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.6, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) {
    var gsl_detailValue_4 = vec2<f32>(0.0, 0.0);
    if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_4 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_599 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
      let gsl_tmp_600 = length(gsl_tmp_599.xy);
      let gsl_tmp_601 = gsl_tmp_599.xy / max(gsl_tmp_600, 0.001) * 40.0;
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      let gsl_tmp_583 = gsl_tmp_600 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_601.x, gsl_tmp_601.y, gsl_tmp_599.z * 0.25), color_white, color_gray, noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_599.x, gsl_tmp_599.y, gsl_tmp_599.z * 0.2), color_white, color_gray, noiseTex);
      let gsl_tmp_584 = floor(gsl_tmp_583);
      let gsl_tmp_585 = wood_ringBoundary(gsl_tmp_584, 3.0, noiseTex);
      let gsl_tmp_586 = wood_ringBoundary(gsl_tmp_584 + 1.0, 3.0, noiseTex);
      let gsl_tmp_588 = gsl_tmp_583 >= gsl_tmp_586;
      let gsl_tmp_587 = gsl_tmp_583 < gsl_tmp_585;
      let gsl_tmp_589 = select(select(gsl_tmp_585, gsl_tmp_586, gsl_tmp_588), wood_ringBoundary(gsl_tmp_584 - 1.0, 3.0, noiseTex), gsl_tmp_587);
      gsl_detailValue_4 = wood_poreSurfaceAt(map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex), (gsl_tmp_583 - gsl_tmp_589) / (select(select(gsl_tmp_586, wood_ringBoundary(gsl_tmp_584 + 2.0, 3.0, noiseTex), gsl_tmp_588), gsl_tmp_585, gsl_tmp_587) - gsl_tmp_589), 1.5, 0.5, 1.0, 0.0, noiseTex) * 1.0;
      gsl_pw = gsl_detailOuterPw_4;
    }
    var gsl_simpleValue_11 = mat4x4<f32>(vec4<f32>(0.6685386, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_368 = gsl_detailValue_4;
      let gsl_tmp_370 = 0.5851758 + 0.3148242 * gsl_tmp_368.y;
      let gsl_tmp_371 = vec2<f32>(gsl_tmp_370, gsl_tmp_368.x);
      let gsl_simpleFull = material_withDisplacement(wood_woodFinish(sqrt(sqrt(gsl_tmp_371.x * gsl_tmp_371.x * gsl_tmp_371.x * gsl_tmp_371.x + 0.0825 * material_hiddenShare__spec1(1.0 - texture_mask(gsl_pw, 0.105, 0.035, noiseTex), noiseTex))), noiseTex), gsl_tmp_371.y, noiseTex);
      gsl_simpleValue_11 = mat4x4<f32>(mix(gsl_simpleValue_11[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_11[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_11[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_11[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let oakMat = gsl_simpleValue_11;
    let gsl_m = oakMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    var gsl_simpleValue_12 = mat4x4<f32>(vec4<f32>(0.71344036, 0.43948862, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_377 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_378 = vec2<f32>(gsl_tmp_377.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_377.y + 0.5);
      let gsl_tmp_379 = vec2<f32>(gsl_tmp_378.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_378.y + 0.25);
      let gsl_tmp_380 = vec2<f32>(gsl_tmp_379.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_379.y + 0.0);
      let gsl_tmp_381 = vec2<f32>(gsl_tmp_380.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_380.y + 0.0);
      let gsl_tmp_382 = vec2<f32>(gsl_tmp_381.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_381.y + 0.0);
      let gsl_tmp_37 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_382.x / max(gsl_tmp_382.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_38 = 0.35 * texture_mask(gsl_tmp_37, -0.3, 0.3, noiseTex);
      let gsl_simpleFull_2 = texture_paintMaterial(gsl_tmp_37, 0.0, 0.08, material_blend(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_38, noiseTex), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
      gsl_simpleValue_12 = mat4x4<f32>(mix(gsl_simpleValue_12[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_12[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_12[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_12[3], gsl_simpleFull_2[3], gsl_simpleDetail));
    }
    let rustedMat = gsl_simpleValue_12;
    let gsl_m = rustedMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    var gsl_simpleValue_13 = mat4x4<f32>(vec4<f32>(0.66949147, 0.43948862, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_474 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_475 = vec2<f32>(gsl_tmp_474.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_474.y + 0.5);
      let gsl_tmp_476 = vec2<f32>(gsl_tmp_475.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_475.y + 0.25);
      let gsl_tmp_477 = vec2<f32>(gsl_tmp_476.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_476.y + 0.0);
      let gsl_tmp_478 = vec2<f32>(gsl_tmp_477.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_477.y + 0.0);
      let gsl_tmp_479 = vec2<f32>(gsl_tmp_478.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_478.y + 0.0);
      let gsl_tmp_47 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_479.x / max(gsl_tmp_479.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_48 = 0.35 * texture_mask(gsl_tmp_47, -0.3, 0.3, noiseTex);
      let gsl_simpleFull_3 = texture_paintMaterial(gsl_tmp_47, 0.0, 0.08, material_blend(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_48, noiseTex), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
      gsl_simpleValue_13 = mat4x4<f32>(mix(gsl_simpleValue_13[0], gsl_simpleFull_3[0], gsl_simpleDetail), mix(gsl_simpleValue_13[1], gsl_simpleFull_3[1], gsl_simpleDetail), mix(gsl_simpleValue_13[2], gsl_simpleFull_3[2], gsl_simpleDetail), mix(gsl_simpleValue_13[3], gsl_simpleFull_3[3], gsl_simpleDetail));
    }
    let patinaMat = gsl_simpleValue_13;
    let gsl_m = patinaMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    var gsl_simpleValue_14 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, material_polishedMetal[1] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[1]) * 0.06335398, material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_49 = p / 1.0;
      let gsl_tmp_50 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_49, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      let gsl_simpleFull_4 = material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_50, noiseTex);
      gsl_simpleValue_14 = mat4x4<f32>(mix(gsl_simpleValue_14[0], gsl_simpleFull_4[0], gsl_simpleDetail), mix(gsl_simpleValue_14[1], gsl_simpleFull_4[1], gsl_simpleDetail), mix(gsl_simpleValue_14[2], gsl_simpleFull_4[2], gsl_simpleDetail), mix(gsl_simpleValue_14[3], gsl_simpleFull_4[3], gsl_simpleDetail));
    }
    let glitterMat = gsl_simpleValue_14;
    let gsl_m = glitterMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 8u) {
    var gsl_simpleValue_15 = mat4x4<f32>(vec4<f32>(0.2354835, 0.0, 1.3399427, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_512 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010944857 * sin(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.01742606 * sin(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020284768 * sin(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.022745797 * sin(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023844866 * sin(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022270106 * sin(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.017706865 * sin(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.011648065 * sin(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
      let gsl_tmp_513 = 0.0 + 0.018241428 * cos(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_512) - 0.64525276 * time + 0.0) + 0.029043432 * cos(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_512) - 0.73128647 * time + 2.1) + 0.03380795 * cos(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_512) - 0.81732017 * time + 4.4) + 0.037909664 * cos(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_512) - 0.90335387 * time + 1.3) + 0.039741445 * cos(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_512) - 1.0324044 * time + 5.2) + 0.037116844 * cos(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_512) - 1.2044718 * time + 3.3) + 0.029511444 * cos(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_512) - 1.4625729 * time + 0.7) + 0.019413443 * cos(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_512) - 1.8067077 * time + 2.9);
      let gsl_tmp_514 = texture_mask(gsl_tmp_513 + 0.037855294 * texture_fadeTo(noise_gradientLattice((p - vec3<f32>(0.95782626, 0.28734788, 0.0) * 1.3692689 * time) / 0.5 + vec3<f32>(13.0, 71.0, 5.0), noiseTex), 0.0, 0.5, noiseTex), 0.14711061, 0.018927647, noiseTex);
      let gsl_simpleFull_5 = material_blend(mat4x4<f32>(vec4<f32>(0.23280893, 0.0, 1.34, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.7, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_514, noiseTex);
      gsl_simpleValue_15 = mat4x4<f32>(mix(gsl_simpleValue_15[0], gsl_simpleFull_5[0], gsl_simpleDetail), mix(gsl_simpleValue_15[1], gsl_simpleFull_5[1], gsl_simpleDetail), mix(gsl_simpleValue_15[2], gsl_simpleFull_5[2], gsl_simpleDetail), mix(gsl_simpleValue_15[3], gsl_simpleFull_5[3], gsl_simpleDetail));
    }
    let waterMat = gsl_simpleValue_15;
    let gsl_m = material_withTransmission(waterMat, 0.6, 50.0, noiseTex); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 9u) {
    var gsl_simpleValue_16 = mat4x4<f32>(vec4<f32>(0.494, 0.0, 1.58, 0.0), vec4<f32>(1.0, 0.026, 0.0, 0.146), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_result_4 = gcl_closure_driftingCellsLatticeXY(time, p / 15.0, vec3<f32>(15.0, 15.0, 15.0), 20.0, noiseTex);
      let cells = gsl_result_4.v_1;
      let drift = gsl_result_4.v_2;
      let gsl_simpleFull_6 = material_withDisplacement(material_withEmissive(material_make(sqrt(texture_reliefRoughness((0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)) * (0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)), 0.2, 2.25, noiseTex)), 0.0, 1.58, noiseTex), animated_incandescentColor(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex), 20.0 * (animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex) + (animated_incandescentStrength(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex) - animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex)) * 15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), noiseTex), (-0.225 * exp(-(cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * 15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw) + 0.15 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.84375, 0.28125, noiseTex)), noiseTex);
      gsl_simpleValue_16 = mat4x4<f32>(mix(gsl_simpleValue_16[0], gsl_simpleFull_6[0], gsl_simpleDetail), mix(gsl_simpleValue_16[1], gsl_simpleFull_6[1], gsl_simpleDetail), mix(gsl_simpleValue_16[2], gsl_simpleFull_6[2], gsl_simpleDetail), mix(gsl_simpleValue_16[3], gsl_simpleFull_6[3], gsl_simpleDetail));
    }
    let lavaMat = gsl_simpleValue_16;
    let gsl_m = lavaMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 10u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = 64.000000;
  var second = 64.000000;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
    let h = gsl_cellularHash(pi + o, oct, noiseTex);
    let v = pf - (o + h);
    let d2 = dot(v, v);
    if (d2 < best) {
      second = best;
      secondCell = bestCell;
      secondH = bestH;
      best = d2;
      bestCell = pi + o;
      bestH = h;
    } else if (d2 < second) {
      second = d2;
      secondCell = pi + o;
      secondH = h;
    }
  }
  var out: gsl_CellularScan;
  out.f1 = sqrt(best);
  out.f2 = sqrt(second);
  // Both closest points, so both gradients are available without a second hash.
  out.feature = bestCell + bestH;
  out.farFeature = secondCell + secondH;
  out.nearCell = bestCell;
  out.farCell = secondCell;
  return out;
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

const gsl_foldK_2 = array<vec3<f32>, 27>(vec3<f32>(-1.0, -1.0, -1.0), vec3<f32>(0.0, -1.0, -1.0), vec3<f32>(1.0, -1.0, -1.0), vec3<f32>(-1.0, 0.0, -1.0), vec3<f32>(0.0, 0.0, -1.0), vec3<f32>(1.0, 0.0, -1.0), vec3<f32>(-1.0, 1.0, -1.0), vec3<f32>(0.0, 1.0, -1.0), vec3<f32>(1.0, 1.0, -1.0), vec3<f32>(-1.0, -1.0, 0.0), vec3<f32>(0.0, -1.0, 0.0), vec3<f32>(1.0, -1.0, 0.0), vec3<f32>(-1.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(-1.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(-1.0, -1.0, 1.0), vec3<f32>(0.0, -1.0, 1.0), vec3<f32>(1.0, -1.0, 1.0), vec3<f32>(-1.0, 0.0, 1.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(-1.0, 1.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

const gsl_foldK_3 = array<vec2<f32>, 9>(vec2<f32>(-1.0, -1.0), vec2<f32>(0.0, -1.0), vec2<f32>(1.0, -1.0), vec2<f32>(-1.0, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(1.0, 0.0), vec2<f32>(-1.0, 1.0), vec2<f32>(0.0, 1.0), vec2<f32>(1.0, 1.0));

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 10> {
  var gsl_detailValue_5 = vec2<f32>(0.0, 0.0);
  let gsl_tmp_599 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
  let gsl_tmp_600 = length(gsl_tmp_599.xy);
  let gsl_tmp_601 = gsl_tmp_599.xy / max(gsl_tmp_600, 0.001) * 40.0;
  let color_white = vec3<f32>(1.0, 1.0, 1.0);
  let color_gray = vec3<f32>(0.5, 0.5, 0.5);
  let gsl_tmp_583 = gsl_tmp_600 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_601.x, gsl_tmp_601.y, gsl_tmp_599.z * 0.25), color_white, color_gray, noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_599.x, gsl_tmp_599.y, gsl_tmp_599.z * 0.2), color_white, color_gray, noiseTex);
  let gsl_tmp_584 = floor(gsl_tmp_583);
  let gsl_tmp_585 = wood_ringBoundary(gsl_tmp_584, 3.0, noiseTex);
  let gsl_tmp_586 = wood_ringBoundary(gsl_tmp_584 + 1.0, 3.0, noiseTex);
  let gsl_tmp_588 = gsl_tmp_583 >= gsl_tmp_586;
  let gsl_tmp_587 = gsl_tmp_583 < gsl_tmp_585;
  let gsl_tmp_589 = select(select(gsl_tmp_585, gsl_tmp_586, gsl_tmp_588), wood_ringBoundary(gsl_tmp_584 - 1.0, 3.0, noiseTex), gsl_tmp_587);
  if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_5 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_5 = wood_poreSurfaceAt(map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex), (gsl_tmp_583 - gsl_tmp_589) / (select(select(gsl_tmp_586, wood_ringBoundary(gsl_tmp_584 + 2.0, 3.0, noiseTex), gsl_tmp_588), gsl_tmp_585, gsl_tmp_587) - gsl_tmp_589), 1.5, 0.5, 1.0, 0.0, noiseTex) * 1.0;
    gsl_pw = gsl_detailOuterPw_5;
  }
  let gsl_tmp_368 = gsl_detailValue_5;
  let gsl_tmp_370 = 0.5851758 + 0.3148242 * gsl_tmp_368.y;
  let gsl_tmp_371 = vec2<f32>(gsl_tmp_370, gsl_tmp_368.x);
  var gsl_simpleValue_17 = mat4x4<f32>(vec4<f32>(0.6685386, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_7 = material_withDisplacement(wood_woodFinish(sqrt(sqrt(gsl_tmp_371.x * gsl_tmp_371.x * gsl_tmp_371.x * gsl_tmp_371.x + 0.0825 * material_hiddenShare__spec1(1.0 - texture_mask(gsl_pw, 0.105, 0.035, noiseTex), noiseTex))), noiseTex), gsl_tmp_371.y, noiseTex);
    gsl_simpleValue_17 = mat4x4<f32>(mix(gsl_simpleValue_17[0], gsl_simpleFull_7[0], gsl_simpleDetail), mix(gsl_simpleValue_17[1], gsl_simpleFull_7[1], gsl_simpleDetail), mix(gsl_simpleValue_17[2], gsl_simpleFull_7[2], gsl_simpleDetail), mix(gsl_simpleValue_17[3], gsl_simpleFull_7[3], gsl_simpleDetail));
  }
  let oakMat = gsl_simpleValue_17;
  let gsl_tmp_377 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
  let gsl_tmp_378 = vec2<f32>(gsl_tmp_377.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_377.y + 0.5);
  let gsl_tmp_379 = vec2<f32>(gsl_tmp_378.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_378.y + 0.25);
  let gsl_tmp_380 = vec2<f32>(gsl_tmp_379.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_379.y + 0.0);
  let gsl_tmp_381 = vec2<f32>(gsl_tmp_380.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_380.y + 0.0);
  let gsl_tmp_382 = vec2<f32>(gsl_tmp_381.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_381.y + 0.0);
  let gsl_tmp_37 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_382.x / max(gsl_tmp_382.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
  let gsl_tmp_38 = 0.35 * texture_mask(gsl_tmp_37, -0.3, 0.3, noiseTex);
  var gsl_simpleValue_18 = mat4x4<f32>(vec4<f32>(0.71344036, 0.43948862, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_8 = texture_paintMaterial(gsl_tmp_37, 0.0, 0.08, material_blend(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_38, noiseTex), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
    gsl_simpleValue_18 = mat4x4<f32>(mix(gsl_simpleValue_18[0], gsl_simpleFull_8[0], gsl_simpleDetail), mix(gsl_simpleValue_18[1], gsl_simpleFull_8[1], gsl_simpleDetail), mix(gsl_simpleValue_18[2], gsl_simpleFull_8[2], gsl_simpleDetail), mix(gsl_simpleValue_18[3], gsl_simpleFull_8[3], gsl_simpleDetail));
  }
  let rustedMat = gsl_simpleValue_18;
  let gsl_tmp_474 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
  let gsl_tmp_475 = vec2<f32>(gsl_tmp_474.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_474.y + 0.5);
  let gsl_tmp_476 = vec2<f32>(gsl_tmp_475.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_475.y + 0.25);
  let gsl_tmp_477 = vec2<f32>(gsl_tmp_476.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_476.y + 0.0);
  let gsl_tmp_478 = vec2<f32>(gsl_tmp_477.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_477.y + 0.0);
  let gsl_tmp_479 = vec2<f32>(gsl_tmp_478.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_478.y + 0.0);
  let gsl_tmp_47 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_479.x / max(gsl_tmp_479.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
  let gsl_tmp_48 = 0.35 * texture_mask(gsl_tmp_47, -0.3, 0.3, noiseTex);
  var gsl_simpleValue_19 = mat4x4<f32>(vec4<f32>(0.66949147, 0.43948862, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_9 = texture_paintMaterial(gsl_tmp_47, 0.0, 0.08, material_blend(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_48, noiseTex), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
    gsl_simpleValue_19 = mat4x4<f32>(mix(gsl_simpleValue_19[0], gsl_simpleFull_9[0], gsl_simpleDetail), mix(gsl_simpleValue_19[1], gsl_simpleFull_9[1], gsl_simpleDetail), mix(gsl_simpleValue_19[2], gsl_simpleFull_9[2], gsl_simpleDetail), mix(gsl_simpleValue_19[3], gsl_simpleFull_9[3], gsl_simpleDetail));
  }
  let patinaMat = gsl_simpleValue_19;
  let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let gsl_tmp_49 = p / 1.0;
  let gsl_tmp_50 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_49, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
  var gsl_simpleValue_20 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, material_polishedMetal[1] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[1]) * 0.06335398, material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_10 = material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_50, noiseTex);
    gsl_simpleValue_20 = mat4x4<f32>(mix(gsl_simpleValue_20[0], gsl_simpleFull_10[0], gsl_simpleDetail), mix(gsl_simpleValue_20[1], gsl_simpleFull_10[1], gsl_simpleDetail), mix(gsl_simpleValue_20[2], gsl_simpleFull_10[2], gsl_simpleDetail), mix(gsl_simpleValue_20[3], gsl_simpleFull_10[3], gsl_simpleDetail));
  }
  let glitterMat = gsl_simpleValue_20;
  let gsl_tmp_512 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010944857 * sin(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.01742606 * sin(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020284768 * sin(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.022745797 * sin(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023844866 * sin(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022270106 * sin(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.017706865 * sin(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.011648065 * sin(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
  let gsl_tmp_513 = 0.0 + 0.018241428 * cos(0.35342917 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_512) - 0.64525276 * time + 0.0) + 0.029043432 * cos(0.45396015 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_512) - 0.73128647 * time + 2.1) + 0.03380795 * cos(0.5670575 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_512) - 0.81732017 * time + 4.4) + 0.037909664 * cos(0.6927212 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_512) - 0.90335387 * time + 1.3) + 0.039741445 * cos(0.90477866 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_512) - 1.0324044 * time + 5.2) + 0.037116844 * cos(1.2315043 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_512) - 1.2044718 * time + 3.3) + 0.029511444 * cos(1.8158406 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_512) - 1.4625729 * time + 0.7) + 0.019413443 * cos(2.7708848 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_512) - 1.8067077 * time + 2.9);
  let gsl_tmp_514 = texture_mask(gsl_tmp_513 + 0.037855294 * texture_fadeTo(noise_gradientLattice((p - vec3<f32>(0.95782626, 0.28734788, 0.0) * 1.3692689 * time) / 0.5 + vec3<f32>(13.0, 71.0, 5.0), noiseTex), 0.0, 0.5, noiseTex), 0.14711061, 0.018927647, noiseTex);
  var gsl_simpleValue_21 = mat4x4<f32>(vec4<f32>(0.2354835, 0.0, 1.3399427, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_11 = material_blend(mat4x4<f32>(vec4<f32>(0.23280893, 0.0, 1.34, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.7, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_514, noiseTex);
    gsl_simpleValue_21 = mat4x4<f32>(mix(gsl_simpleValue_21[0], gsl_simpleFull_11[0], gsl_simpleDetail), mix(gsl_simpleValue_21[1], gsl_simpleFull_11[1], gsl_simpleDetail), mix(gsl_simpleValue_21[2], gsl_simpleFull_11[2], gsl_simpleDetail), mix(gsl_simpleValue_21[3], gsl_simpleFull_11[3], gsl_simpleDetail));
  }
  let waterMat = gsl_simpleValue_21;
  let gsl_result_4 = gcl_closure_driftingCellsLatticeXY(time, p / 15.0, vec3<f32>(15.0, 15.0, 15.0), 20.0, noiseTex);
  let cells = gsl_result_4.v_1;
  let drift = gsl_result_4.v_2;
  var gsl_simpleValue_22 = mat4x4<f32>(vec4<f32>(0.494, 0.0, 1.58, 0.0), vec4<f32>(1.0, 0.026, 0.0, 0.146), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_12 = material_withDisplacement(material_withEmissive(material_make(sqrt(texture_reliefRoughness((0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)) * (0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)), 0.2, 2.25, noiseTex)), 0.0, 1.58, noiseTex), animated_incandescentColor(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex), 20.0 * (animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex) + (animated_incandescentStrength(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex) - animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex)) * 15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), noiseTex), (-0.225 * exp(-(cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * 15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(15.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw) + 0.15 * noise_gradientLattice((p - drift) / 2.25 + vec3<f32>(7.0, 31.0, 13.0), noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.84375, 0.28125, noiseTex)), noiseTex);
    gsl_simpleValue_22 = mat4x4<f32>(mix(gsl_simpleValue_22[0], gsl_simpleFull_12[0], gsl_simpleDetail), mix(gsl_simpleValue_22[1], gsl_simpleFull_12[1], gsl_simpleDetail), mix(gsl_simpleValue_22[2], gsl_simpleFull_12[2], gsl_simpleDetail), mix(gsl_simpleValue_22[3], gsl_simpleFull_12[3], gsl_simpleDetail));
  }
  let lavaMat = gsl_simpleValue_22;
  return array<f32, 10>(0.0, 0.0, oakMat[2].z, rustedMat[2].z, patinaMat[2].z, glitterMat[2].z, 0.0, material_withTransmission(waterMat, 0.6, 50.0, noiseTex)[2].z, lavaMat[2].z, 0.0);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, noiseTex);
  let gsl_h = gsl_heights(p, time, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  let gsl_s3 = gsl_shape(gsl_d[2] - gsl_h[2] * gsl_dispK, 3u);
  let gsl_s4 = gsl_shape(gsl_d[3] - gsl_h[3] * gsl_dispK, 4u);
  let gsl_s5 = gsl_shape(gsl_d[4] - gsl_h[4] * gsl_dispK, 5u);
  let gsl_s6 = gsl_shape(gsl_d[5] - gsl_h[5] * gsl_dispK, 6u);
  let gsl_s7 = gsl_shape(gsl_d[6], 7u);
  let gsl_s8 = gsl_shape(gsl_d[7] - gsl_h[7] * gsl_dispK, 8u);
  let gsl_s9 = gsl_shape(gsl_d[8] - gsl_h[8] * gsl_dispK, 9u);
  let gsl_s10 = gsl_shape(gsl_d[9], 10u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7), gsl_s8), gsl_s9), gsl_s10);
}

struct gsl_struct {
  v_1 : f32,
  v_2 : f32,
  v_3 : vec3<f32>,
  v_4 : vec3<f32>,
}
struct gsl_struct_2 {
  v_1 : f32,
  v_2 : f32,
}
struct gsl_struct_3 {
  v_1 : vec4<f32>,
  v_2 : vec3<f32>,
}
