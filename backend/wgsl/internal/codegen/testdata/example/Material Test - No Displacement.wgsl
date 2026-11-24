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

fn gsl_remap(p: vec3<f32>, color_white: vec3<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_168 = clamp((-noise_voronoi(p, 5.0, 1.0, 0.0, 0.0, noiseTex) / 5.0 - -0.38) / 0.08 * 0.5 + 0.5, 0.0, 1.0);
  return gsl_tmp_168 * gsl_tmp_168 * (3.0 - 2.0 * gsl_tmp_168);
}

fn texture_paint(field: f32, edge: f32, width: f32, colorA: vec3<f32>, colorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return color_mixOklab(colorA, colorB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn gsl_remap_2(p: vec3<f32>, color_white: vec3<f32>, color_gray: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let gsl_tmp_358 = clamp((-noise_voronoi(p, 5.0, 1.0, 0.0, 0.0, noiseTex) / 5.0 - -0.38) / 0.08 * 0.5 + 0.5, 0.0, 1.0);
  return gsl_tmp_358 * gsl_tmp_358 * (3.0 - 2.0 * gsl_tmp_358);
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
  let gsl_result_4 = noise_cellularEdgesZBlendLattice(qxy, zBand, zBlendT, noiseTex);
  let f2 = gsl_result_4.v_2;
  let f1 = gsl_result_4.v_1;
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

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
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

fn playful_glitterShare(density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - exp(-4.18879 * density * density * density);
}

fn playful_glitterFlakes(f1: f32, flakeCellSize: f32, density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return texture_fadeTo(texture_coverageDisc(f1, density * flakeCellSize, noiseTex), playful_glitterShare(density, noiseTex), flakeCellSize, noiseTex);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-250.0, -25.0, -25.0);
  let gsl_partHi = vec3<f32>(-200.0, 25.0, 25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_sphere(p - vec3<f32>(-225.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let marbleBall = gsl_part;
  let gsl_partLo_2 = vec3<f32>(-175.0, -25.0, -25.0);
  let gsl_partHi_2 = vec3<f32>(-125.0, 25.0, 25.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_sphere(p - vec3<f32>(-150.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let graniteBall = gsl_part_2;
  let gsl_partLo_3 = vec3<f32>(-100.0, -25.0, -25.0);
  let gsl_partHi_3 = vec3<f32>(-50.0, 25.0, 25.0);
  let gsl_partBox_3 = gsl_boxDistance(p, gsl_partLo_3, gsl_partHi_3);
  var gsl_part_3 = gsl_partBox_3;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_3 <= gsl_vanishView.w) {
    let gsl_partVanish_3 = gsl_vanish(gsl_partLo_3, gsl_partHi_3);
    gsl_part_3 = 1e10;
    if (gsl_partVanish_3.x > 0.0) { gsl_part_3 = builtin_sphere(p - vec3<f32>(-75.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_3.y; }
    if (abs(gsl_part_3) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_3), gsl_partVanish_3.zw); }
  }
  let rustedBall = gsl_part_3;
  let gsl_partLo_4 = vec3<f32>(-25.0, -25.0, -25.0);
  let gsl_partHi_4 = vec3<f32>(25.0, 25.0, 25.0);
  let gsl_partBox_4 = gsl_boxDistance(p, gsl_partLo_4, gsl_partHi_4);
  var gsl_part_4 = gsl_partBox_4;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_4 <= gsl_vanishView.w) {
    let gsl_partVanish_4 = gsl_vanish(gsl_partLo_4, gsl_partHi_4);
    gsl_part_4 = 1e10;
    if (gsl_partVanish_4.x > 0.0) { gsl_part_4 = builtin_sphere(p - vec3<f32>(0.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_4.y; }
    if (abs(gsl_part_4) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_4), gsl_partVanish_4.zw); }
  }
  let patinaBall = gsl_part_4;
  let gsl_partLo_5 = vec3<f32>(50.0, -25.0, -25.0);
  let gsl_partHi_5 = vec3<f32>(100.0, 25.0, 25.0);
  let gsl_partBox_5 = gsl_boxDistance(p, gsl_partLo_5, gsl_partHi_5);
  var gsl_part_5 = gsl_partBox_5;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_5 <= gsl_vanishView.w) {
    let gsl_partVanish_5 = gsl_vanish(gsl_partLo_5, gsl_partHi_5);
    gsl_part_5 = 1e10;
    if (gsl_partVanish_5.x > 0.0) { gsl_part_5 = builtin_sphere(p - vec3<f32>(75.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_5.y; }
    if (abs(gsl_part_5) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_5), gsl_partVanish_5.zw); }
  }
  let glitterBall = gsl_part_5;
  let gsl_partLo_6 = vec3<f32>(125.0, -25.0, -25.0);
  let gsl_partHi_6 = vec3<f32>(175.0, 25.0, 25.0);
  let gsl_partBox_6 = gsl_boxDistance(p, gsl_partLo_6, gsl_partHi_6);
  var gsl_part_6 = gsl_partBox_6;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_6 <= gsl_vanishView.w) {
    let gsl_partVanish_6 = gsl_vanish(gsl_partLo_6, gsl_partHi_6);
    gsl_part_6 = 1e10;
    if (gsl_partVanish_6.x > 0.0) { gsl_part_6 = builtin_sphere(p - vec3<f32>(150.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_6.y; }
    if (abs(gsl_part_6) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_6), gsl_partVanish_6.zw); }
  }
  let checkerBall = gsl_part_6;
  let gsl_partLo_7 = vec3<f32>(200.0, -25.0, -25.0);
  let gsl_partHi_7 = vec3<f32>(250.0, 25.0, 25.0);
  let gsl_partBox_7 = gsl_boxDistance(p, gsl_partLo_7, gsl_partHi_7);
  var gsl_part_7 = gsl_partBox_7;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_7 <= gsl_vanishView.w) {
    let gsl_partVanish_7 = gsl_vanish(gsl_partLo_7, gsl_partHi_7);
    gsl_part_7 = 1e10;
    if (gsl_partVanish_7.x > 0.0) { gsl_part_7 = builtin_sphere(p - vec3<f32>(225.0, 0.0, 0.0), 25.0, noiseTex) + gsl_partVanish_7.y; }
    if (abs(gsl_part_7) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_7), gsl_partVanish_7.zw); }
  }
  let skinBall = gsl_part_7;
  let gsl_s1 = gsl_shape(marbleBall, 1u);
  let gsl_s2 = gsl_shape(graniteBall, 2u);
  let gsl_s3 = gsl_shape(rustedBall, 3u);
  let gsl_s4 = gsl_shape(patinaBall, 4u);
  let gsl_s5 = gsl_shape(glitterBall, 5u);
  let gsl_s6 = gsl_shape(checkerBall, 6u);
  let gsl_s7 = gsl_shape(skinBall, 7u);
  return gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_min(gsl_s1, gsl_s2), gsl_s3), gsl_s4), gsl_s5), gsl_s6), gsl_s7);
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
    var gsl_simpleValue_3 = vec3<f32>(0.6824685, 0.58169425, 0.44390574);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_162 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_163 = vec2<f32>(gsl_tmp_162.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_162.y + 0.5);
      let gsl_tmp_164 = vec2<f32>(gsl_tmp_163.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_163.y + 0.25);
      let gsl_tmp_165 = vec2<f32>(gsl_tmp_164.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_164.y + 0.0);
      let gsl_tmp_166 = vec2<f32>(gsl_tmp_165.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_165.y + 0.0);
      let gsl_tmp_167 = vec2<f32>(gsl_tmp_166.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_166.y + 0.0);
      let gsl_tmp_37 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_167.x / max(gsl_tmp_167.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_38 = 0.35 * texture_mask(gsl_tmp_37, -0.3, 0.3, noiseTex);
      let gsl_tmp_183 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_184 = vec2<f32>(gsl_tmp_183.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_183.y + 0.6);
      let gsl_tmp_185 = vec2<f32>(gsl_tmp_184.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_184.y + 0.36);
      let gsl_tmp_186 = vec2<f32>(gsl_tmp_185.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_185.y + 0.216);
      let gsl_tmp_187 = vec2<f32>(gsl_tmp_186.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_186.y + 0.1296);
      let gsl_tmp_188 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_189 = vec2<f32>(gsl_tmp_188.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_188.y + 0.6);
      let gsl_tmp_190 = vec2<f32>(gsl_tmp_189.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_189.y + 0.36);
      let gsl_tmp_191 = vec2<f32>(gsl_tmp_190.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_190.y + 0.216);
      let gsl_tmp_192 = vec2<f32>(gsl_tmp_191.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_191.y + 0.1296);
      let gsl_tmp_193 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_194 = vec2<f32>(gsl_tmp_193.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_193.y + 0.5);
      let gsl_tmp_195 = vec2<f32>(gsl_tmp_194.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_194.y + 0.25);
      let gsl_tmp_196 = vec2<f32>(gsl_tmp_195.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_195.y + 0.0);
      let gsl_tmp_197 = vec2<f32>(gsl_tmp_196.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_196.y + 0.0);
      let gsl_tmp_198 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_199 = vec2<f32>(gsl_tmp_198.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_198.y + 0.5);
      let gsl_tmp_200 = vec2<f32>(gsl_tmp_199.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_199.y + 0.25);
      let gsl_tmp_201 = vec2<f32>(gsl_tmp_200.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_200.y + 0.0);
      let gsl_tmp_202 = vec2<f32>(gsl_tmp_201.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_201.y + 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      let gsl_tmp_169 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), 1.0);
      let gsl_tmp_170 = vec2<f32>(gsl_tmp_169.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_169.y + 0.5);
      let gsl_tmp_171 = vec2<f32>(gsl_tmp_170.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_170.y + 0.25);
      let gsl_tmp_172 = vec2<f32>(gsl_tmp_171.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_171.y + 0.0);
      let gsl_tmp_173 = vec2<f32>(gsl_tmp_172.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_172.y + 0.0);
      let gsl_tmp_174 = vec2<f32>(gsl_tmp_173.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_173.y + 0.0);
      let gsl_tmp_39 = texture_fadeTo(gsl_remap(p + (noise_gradientLatticeVector(p / 4.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 4.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.75) / 0.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.44, color_white, color_gray, noiseTex) * texture_mask(2.0 * gsl_tmp_174.x / max(gsl_tmp_174.y, 0.000001), -0.15, 0.2, noiseTex), 0.159, 3.0, noiseTex) - 0.159;
      let gsl_tmp_203 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_204 = vec2<f32>(gsl_tmp_203.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_203.y + 0.6);
      let gsl_tmp_205 = vec2<f32>(gsl_tmp_204.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_204.y + 0.36);
      let gsl_tmp_206 = vec2<f32>(gsl_tmp_205.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_205.y + 0.216);
      let gsl_tmp_207 = vec2<f32>(gsl_tmp_206.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_206.y + 0.1296);
      let gsl_tmp_208 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_209 = vec2<f32>(gsl_tmp_208.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_208.y + 0.6);
      let gsl_tmp_210 = vec2<f32>(gsl_tmp_209.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_209.y + 0.36);
      let gsl_tmp_211 = vec2<f32>(gsl_tmp_210.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_210.y + 0.216);
      let gsl_tmp_212 = vec2<f32>(gsl_tmp_211.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_211.y + 0.1296);
      let gsl_tmp_213 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_214 = vec2<f32>(gsl_tmp_213.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_213.y + 0.5);
      let gsl_tmp_215 = vec2<f32>(gsl_tmp_214.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_214.y + 0.25);
      let gsl_tmp_216 = vec2<f32>(gsl_tmp_215.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_215.y + 0.0);
      let gsl_tmp_217 = vec2<f32>(gsl_tmp_216.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_216.y + 0.0);
      let gsl_tmp_218 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_219 = vec2<f32>(gsl_tmp_218.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_218.y + 0.5);
      let gsl_tmp_220 = vec2<f32>(gsl_tmp_219.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_219.y + 0.25);
      let gsl_tmp_221 = vec2<f32>(gsl_tmp_220.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_220.y + 0.0);
      let gsl_tmp_222 = vec2<f32>(gsl_tmp_221.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_221.y + 0.0);
      let gsl_tmp_223 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_224 = vec2<f32>(gsl_tmp_223.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_223.y + 0.6);
      let gsl_tmp_225 = vec2<f32>(gsl_tmp_224.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_224.y + 0.36);
      let gsl_tmp_226 = vec2<f32>(gsl_tmp_225.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_225.y + 0.216);
      let gsl_tmp_227 = vec2<f32>(gsl_tmp_226.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_226.y + 0.1296);
      let gsl_tmp_228 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_229 = vec2<f32>(gsl_tmp_228.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_228.y + 0.6);
      let gsl_tmp_230 = vec2<f32>(gsl_tmp_229.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_229.y + 0.36);
      let gsl_tmp_231 = vec2<f32>(gsl_tmp_230.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_230.y + 0.216);
      let gsl_tmp_232 = vec2<f32>(gsl_tmp_231.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_231.y + 0.1296);
      let gsl_tmp_233 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_234 = vec2<f32>(gsl_tmp_233.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_233.y + 0.5);
      let gsl_tmp_235 = vec2<f32>(gsl_tmp_234.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_234.y + 0.25);
      let gsl_tmp_236 = vec2<f32>(gsl_tmp_235.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_235.y + 0.0);
      let gsl_tmp_237 = vec2<f32>(gsl_tmp_236.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_236.y + 0.0);
      let gsl_tmp_238 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_239 = vec2<f32>(gsl_tmp_238.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_238.y + 0.5);
      let gsl_tmp_240 = vec2<f32>(gsl_tmp_239.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_239.y + 0.25);
      let gsl_tmp_241 = vec2<f32>(gsl_tmp_240.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_240.y + 0.0);
      let gsl_tmp_242 = vec2<f32>(gsl_tmp_241.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_241.y + 0.0);
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, texture_paint(gsl_tmp_37, 0.0, 0.08, color_mixOklab(vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.6585775, 0.47784516, 0.14561833), gsl_tmp_38, noiseTex), color_mixOklab(vec3<f32>(0.6585775, 0.47784516, 0.14561833), color_oklab2rgb(vec3<f32>(0.4104389, 0.04798136, 0.060527567) + vec3<f32>(0.09 * (2.0 * (gsl_tmp_187.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_192.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_197.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_202.y + 0.0, 0.000001)) - 0.11 * gsl_tmp_39, 0.006 * (2.0 * (gsl_tmp_207.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_212.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_217.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_222.y + 0.0, 0.000001)) - 0.01 * gsl_tmp_39, 0.025 * (2.0 * (gsl_tmp_227.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_232.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_237.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_242.y + 0.0, 0.000001)) - 0.02 * gsl_tmp_39), noiseTex), texture_mask(gsl_tmp_37, 0.25, 0.25, noiseTex), noiseTex), noiseTex), gsl_simpleDetail);
    }
    let rustedColor = gsl_simpleValue_3;
    return vec4<f32>(rustedColor, 1.0);
  }
  if (gsl_colorIdx == 4u) {
    var gsl_simpleValue_4 = vec3<f32>(0.7927733, 0.85689414, 0.7547779);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_259 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_260 = vec2<f32>(gsl_tmp_259.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_259.y + 0.5);
      let gsl_tmp_261 = vec2<f32>(gsl_tmp_260.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_260.y + 0.25);
      let gsl_tmp_262 = vec2<f32>(gsl_tmp_261.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_261.y + 0.0);
      let gsl_tmp_263 = vec2<f32>(gsl_tmp_262.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_262.y + 0.0);
      let gsl_tmp_264 = vec2<f32>(gsl_tmp_263.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_263.y + 0.0);
      let gsl_tmp_47 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_264.x / max(gsl_tmp_264.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_48 = 0.35 * texture_mask(gsl_tmp_47, -0.3, 0.3, noiseTex);
      let gsl_tmp_337 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_338 = vec2<f32>(gsl_tmp_337.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_337.y + 0.6);
      let gsl_tmp_339 = vec2<f32>(gsl_tmp_338.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_338.y + 0.36);
      let gsl_tmp_340 = vec2<f32>(gsl_tmp_339.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_339.y + 0.216);
      let gsl_tmp_341 = vec2<f32>(gsl_tmp_340.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_340.y + 0.1296);
      let gsl_tmp_342 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), 1.0);
      let gsl_tmp_343 = vec2<f32>(gsl_tmp_342.x + 0.6 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_342.y + 0.6);
      let gsl_tmp_344 = vec2<f32>(gsl_tmp_343.x + 0.36 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_343.y + 0.36);
      let gsl_tmp_345 = vec2<f32>(gsl_tmp_344.x + 0.216 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex), gsl_tmp_344.y + 0.216);
      let gsl_tmp_346 = vec2<f32>(gsl_tmp_345.x + 0.1296 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.1171875) / 0.0390625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.3125, noiseTex), gsl_tmp_345.y + 0.1296);
      let gsl_tmp_347 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_348 = vec2<f32>(gsl_tmp_347.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_347.y + 0.5);
      let gsl_tmp_349 = vec2<f32>(gsl_tmp_348.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_348.y + 0.25);
      let gsl_tmp_350 = vec2<f32>(gsl_tmp_349.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_349.y + 0.0);
      let gsl_tmp_351 = vec2<f32>(gsl_tmp_350.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_350.y + 0.0);
      let gsl_tmp_352 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 15.0) / 5.0 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 40.0, noiseTex), 1.0);
      let gsl_tmp_353 = vec2<f32>(gsl_tmp_352.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), gsl_tmp_352.y + 0.5);
      let gsl_tmp_354 = vec2<f32>(gsl_tmp_353.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_353.y + 0.25);
      let gsl_tmp_355 = vec2<f32>(gsl_tmp_354.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_354.y + 0.0);
      let gsl_tmp_356 = vec2<f32>(gsl_tmp_355.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_355.y + 0.0);
      let gsl_tmp_357 = 2.0 * (gsl_tmp_341.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.05859375) / 0.01953125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.15625, noiseTex)) / max(gsl_tmp_346.y + 0.0, 0.000001) + 0.4 * 2.0 * (gsl_tmp_351.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex)) / max(gsl_tmp_356.y + 0.0, 0.000001);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      let gsl_tmp_359 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), 1.0);
      let gsl_tmp_360 = vec2<f32>(gsl_tmp_359.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_359.y + 0.5);
      let gsl_tmp_361 = vec2<f32>(gsl_tmp_360.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_360.y + 0.25);
      let gsl_tmp_362 = vec2<f32>(gsl_tmp_361.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_361.y + 0.0);
      let gsl_tmp_363 = vec2<f32>(gsl_tmp_362.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_362.y + 0.0);
      let gsl_tmp_364 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 7.5) / 2.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 20.0, noiseTex), 1.0);
      let gsl_tmp_365 = vec2<f32>(gsl_tmp_364.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 3.75) / 1.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 10.0, noiseTex), gsl_tmp_364.y + 0.5);
      let gsl_tmp_366 = vec2<f32>(gsl_tmp_365.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.875) / 0.625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 5.0, noiseTex), gsl_tmp_365.y + 0.25);
      let gsl_tmp_367 = vec2<f32>(gsl_tmp_366.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.9375) / 0.3125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 2.5, noiseTex), gsl_tmp_366.y + 0.0);
      let gsl_tmp_368 = vec2<f32>(gsl_tmp_367.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.46875) / 0.15625 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.25, noiseTex), gsl_tmp_367.y + 0.0);
      let gsl_tmp_369 = texture_fadeTo(gsl_remap_2(p + (noise_gradientLatticeVector(p / 4.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 4.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.75) / 0.25 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.44, color_white, color_gray, noiseTex) * texture_mask(2.0 * (gsl_tmp_363.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.234375) / 0.078125 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 0.625, noiseTex)) / max(gsl_tmp_368.y + 0.0, 0.000001), -0.15, 0.2, noiseTex), 0.159, 3.0, noiseTex) - 0.159;
      gsl_simpleValue_4 = mix(gsl_simpleValue_4, texture_paint(gsl_tmp_47, 0.0, 0.08, color_mixOklab(vec3<f32>(0.9695, 0.8112, 0.7497), vec3<f32>(0.67572874, 0.9464596, 0.7818527), gsl_tmp_48, noiseTex), color_mixOklab(vec3<f32>(0.67572874, 0.9464596, 0.7818527), color_oklab2rgb(vec3<f32>(0.702035, -0.058106415, -0.015318653) + vec3<f32>(0.09 * gsl_tmp_357 - 0.11 * gsl_tmp_369, 0.006 * gsl_tmp_357 - 0.01 * gsl_tmp_369, 0.025 * gsl_tmp_357 - 0.02 * gsl_tmp_369), noiseTex), texture_mask(gsl_tmp_47, 0.25, 0.25, noiseTex), noiseTex), noiseTex), gsl_simpleDetail);
    }
    let patinaColor = gsl_simpleValue_4;
    return vec4<f32>(patinaColor, 1.0);
  }
  if (gsl_colorIdx == 5u) {
    var gsl_simpleValue_5 = vec3<f32>(0.0032853691, 0.003285369, 0.003285368);
    if (gsl_simpleDetail > 0.0) {
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let gsl_tmp_49 = p / 1.0;
      let gsl_tmp_50 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_49, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      gsl_simpleValue_5 = mix(gsl_simpleValue_5, color_mixOklab(color_black, color_white, gsl_tmp_50, noiseTex), gsl_simpleDetail);
    }
    let glitterColor = gsl_simpleValue_5;
    return vec4<f32>(glitterColor, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    var gsl_simpleValue_6 = vec3<f32>(0.38857287, 0.38857284, 0.3885728);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_58 = texture_squareWave(p.x, 20.0, noiseTex);
      let gsl_tmp_59 = texture_squareWave(p.y, 20.0, noiseTex);
      let gsl_tmp_61 = gsl_tmp_58 + gsl_tmp_59 - 2.0 * gsl_tmp_58 * gsl_tmp_59;
      let gsl_tmp_60 = texture_squareWave(p.z, 20.0, noiseTex);
      gsl_simpleValue_6 = mix(gsl_simpleValue_6, color_mixOklab(color_white, color_black, gsl_tmp_61 + gsl_tmp_60 - 2.0 * gsl_tmp_61 * gsl_tmp_60, noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_6, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    var gsl_simpleValue_7 = vec3<f32>(0.8056361, 0.8831174, 0.8809763);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_teal = vec3<f32>(0.0, 0.5, 0.5);
      gsl_simpleValue_7 = mix(gsl_simpleValue_7, color_mixOklab(color_white, color_teal, texture_fadeTo(texture_coverageLine(noise_cellularEdgesZBlend(p, 5.0, 2.5, noiseTex), 1.0, noiseTex), 0.228, 5.0, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let skinColor = gsl_simpleValue_7;
    return vec4<f32>(skinColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.15, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 2u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.6, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
  if (gsl_materialIdx == 3u) {
    var gsl_simpleValue_8 = mat4x4<f32>(vec4<f32>(0.71344036, 0.43948862, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_162 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_163 = vec2<f32>(gsl_tmp_162.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_162.y + 0.5);
      let gsl_tmp_164 = vec2<f32>(gsl_tmp_163.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_163.y + 0.25);
      let gsl_tmp_165 = vec2<f32>(gsl_tmp_164.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_164.y + 0.0);
      let gsl_tmp_166 = vec2<f32>(gsl_tmp_165.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_165.y + 0.0);
      let gsl_tmp_167 = vec2<f32>(gsl_tmp_166.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_166.y + 0.0);
      let gsl_tmp_37 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_167.x / max(gsl_tmp_167.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_38 = 0.35 * texture_mask(gsl_tmp_37, -0.3, 0.3, noiseTex);
      let gsl_simpleFull = texture_paintMaterial(gsl_tmp_37, 0.0, 0.08, material_blend(mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_38, noiseTex), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
      gsl_simpleValue_8 = mat4x4<f32>(mix(gsl_simpleValue_8[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_8[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_8[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_8[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let rustedMat = gsl_simpleValue_8;
    let gsl_m = rustedMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    var gsl_simpleValue_9 = mat4x4<f32>(vec4<f32>(0.66949147, 0.43948862, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_259 = vec2<f32>(0.0 + 1.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 22.5) / 7.5 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 60.0, noiseTex), 1.0);
      let gsl_tmp_260 = vec2<f32>(gsl_tmp_259.x + 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 11.25) / 3.75 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 30.0, noiseTex), gsl_tmp_259.y + 0.5);
      let gsl_tmp_261 = vec2<f32>(gsl_tmp_260.x + 0.25 * (1.0 - ease_quadratic(clamp((gsl_pw - 5.625) / 1.875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 15.0, noiseTex), gsl_tmp_260.y + 0.25);
      let gsl_tmp_262 = vec2<f32>(gsl_tmp_261.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 2.8125) / 0.9375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 7.5, noiseTex), gsl_tmp_261.y + 0.0);
      let gsl_tmp_263 = vec2<f32>(gsl_tmp_262.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 1.40625) / 0.46875 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 3.75, noiseTex), gsl_tmp_262.y + 0.0);
      let gsl_tmp_264 = vec2<f32>(gsl_tmp_263.x + 0.0 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.703125) / 0.234375 * 0.5 + 0.5, 0.0, 1.0), noiseTex)) * noise_gradientLattice(p / 1.875, noiseTex), gsl_tmp_263.y + 0.0);
      let gsl_tmp_47 = 0.5 * (noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex)) + 0.67 * 2.0 * gsl_tmp_264.x / max(gsl_tmp_264.y, 0.000001) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_tmp_48 = 0.35 * texture_mask(gsl_tmp_47, -0.3, 0.3, noiseTex);
      let gsl_simpleFull_2 = texture_paintMaterial(gsl_tmp_47, 0.0, 0.08, material_blend(mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_48, noiseTex), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
      gsl_simpleValue_9 = mat4x4<f32>(mix(gsl_simpleValue_9[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_9[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_9[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_9[3], gsl_simpleFull_2[3], gsl_simpleDetail));
    }
    let patinaMat = gsl_simpleValue_9;
    let gsl_m = patinaMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    var gsl_simpleValue_10 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, material_polishedMetal[1] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[1]) * 0.06335398, material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_49 = p / 1.0;
      let gsl_tmp_50 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_49, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      let gsl_simpleFull_3 = material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_50, noiseTex);
      gsl_simpleValue_10 = mat4x4<f32>(mix(gsl_simpleValue_10[0], gsl_simpleFull_3[0], gsl_simpleDetail), mix(gsl_simpleValue_10[1], gsl_simpleFull_3[1], gsl_simpleDetail), mix(gsl_simpleValue_10[2], gsl_simpleFull_3[2], gsl_simpleDetail), mix(gsl_simpleValue_10[3], gsl_simpleFull_3[3], gsl_simpleDetail));
    }
    let glitterMat = gsl_simpleValue_10;
    let gsl_m = glitterMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let material_matte = mat4x4<f32>(vec4<f32>(1.0, 0.0, 1.45, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_matte; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 7u) {
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

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

const gsl_foldK_2 = array<vec3<f32>, 27>(vec3<f32>(-1.0, -1.0, -1.0), vec3<f32>(0.0, -1.0, -1.0), vec3<f32>(1.0, -1.0, -1.0), vec3<f32>(-1.0, 0.0, -1.0), vec3<f32>(0.0, 0.0, -1.0), vec3<f32>(1.0, 0.0, -1.0), vec3<f32>(-1.0, 1.0, -1.0), vec3<f32>(0.0, 1.0, -1.0), vec3<f32>(1.0, 1.0, -1.0), vec3<f32>(-1.0, -1.0, 0.0), vec3<f32>(0.0, -1.0, 0.0), vec3<f32>(1.0, -1.0, 0.0), vec3<f32>(-1.0, 0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(-1.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(-1.0, -1.0, 1.0), vec3<f32>(0.0, -1.0, 1.0), vec3<f32>(1.0, -1.0, 1.0), vec3<f32>(-1.0, 0.0, 1.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(-1.0, 1.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

const gsl_foldK_3 = array<vec2<f32>, 9>(vec2<f32>(-1.0, -1.0), vec2<f32>(0.0, -1.0), vec2<f32>(1.0, -1.0), vec2<f32>(-1.0, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(1.0, 0.0), vec2<f32>(-1.0, 1.0), vec2<f32>(0.0, 1.0), vec2<f32>(1.0, 1.0));

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
