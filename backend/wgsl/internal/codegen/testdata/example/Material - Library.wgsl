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

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
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

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn texture_paintMaterial(field: f32, edge: f32, width: f32, matA: mat4x4<f32>, matB: mat4x4<f32>, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return material_blend(matA, matB, texture_mask(field, edge, width, noiseTex), noiseTex);
}

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let la = color_rgb2oklab(a, noiseTex);
  let lb = color_rgb2oklab(b, noiseTex);
  return color_oklab2rgb(la + (lb - la) * t, noiseTex);
}

fn texture_paint(field: f32, edge: f32, width: f32, colorA: vec3<f32>, colorB: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return color_mixOklab(colorA, colorB, texture_mask(field, edge, width, noiseTex), noiseTex);
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

fn noise_cellularHash(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
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
    gsl_fold = noise_cellularEdgesZBlendLattice_gsl_anon(pi, zBand, zBlendT, pf, gsl_fold, gsl_foldK_2[gsl_i], noiseTex);
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

fn material_hiddenShare__spec1(fade: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - fade * gsl_dispDrawn;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
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
  let gsl_tmp_133 = poreBase + vec2<f32>(0.0, 0.0);
  let gsl_tmp_134 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_133.x, gsl_tmp_133.y, gsl_tmp_133.x - gsl_tmp_133.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_133.x, gsl_tmp_133.y, gsl_tmp_133.x - gsl_tmp_133.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_135 = floor(gsl_tmp_134);
  let gsl_tmp_136 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_133.x, gsl_tmp_133.y, gsl_tmp_135))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_133.x, gsl_tmp_133.y, gsl_tmp_135)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let sizeShare = wood_poreSizeShare(progress, ringPorosity, poreZone, noiseTex);
  let size = sizeShare.x;
  let poreShare = sizeShare.y;
  let gsl_tmp_137 = select(10.0, min(10.0, length(vec2<f32>(length(poreQ - (gsl_tmp_133 + gsl_tmp_136.xy)), max(abs(gsl_tmp_134 - gsl_tmp_135 - 0.5) - 0.2 - 0.25 * gsl_tmp_136.x, 0.0) * 8.0)) / size), gsl_tmp_136.z < poreShare);
  let gsl_tmp_138 = poreBase + vec2<f32>(1.0, 0.0);
  let gsl_tmp_139 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_138.x, gsl_tmp_138.y, gsl_tmp_138.x - gsl_tmp_138.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_138.x, gsl_tmp_138.y, gsl_tmp_138.x - gsl_tmp_138.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_140 = floor(gsl_tmp_139);
  let gsl_tmp_141 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_138.x, gsl_tmp_138.y, gsl_tmp_140))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_138.x, gsl_tmp_138.y, gsl_tmp_140)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_142 = select(gsl_tmp_137, min(gsl_tmp_137, length(vec2<f32>(length(poreQ - (gsl_tmp_138 + gsl_tmp_141.xy)), max(abs(gsl_tmp_139 - gsl_tmp_140 - 0.5) - 0.2 - 0.25 * gsl_tmp_141.x, 0.0) * 8.0)) / size), gsl_tmp_141.z < poreShare);
  let gsl_tmp_143 = poreBase + vec2<f32>(0.0, 1.0);
  let gsl_tmp_144 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_143.x, gsl_tmp_143.y, gsl_tmp_143.x - gsl_tmp_143.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_143.x, gsl_tmp_143.y, gsl_tmp_143.x - gsl_tmp_143.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_145 = floor(gsl_tmp_144);
  let gsl_tmp_146 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_143.x, gsl_tmp_143.y, gsl_tmp_145))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_143.x, gsl_tmp_143.y, gsl_tmp_145)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let gsl_tmp_147 = select(gsl_tmp_142, min(gsl_tmp_142, length(vec2<f32>(length(poreQ - (gsl_tmp_143 + gsl_tmp_146.xy)), max(abs(gsl_tmp_144 - gsl_tmp_145 - 0.5) - 0.2 - 0.25 * gsl_tmp_146.x, 0.0) * 8.0)) / size), gsl_tmp_146.z < poreShare);
  let gsl_tmp_148 = poreBase + vec2<f32>(1.0, 1.0);
  let gsl_tmp_149 = local.z / (poreCellSize * 8.0) + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_148.x, gsl_tmp_148.y, gsl_tmp_148.x - gsl_tmp_148.y + 6007.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_148.x, gsl_tmp_148.y, gsl_tmp_148.x - gsl_tmp_148.y + 6007.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
  let gsl_tmp_150 = floor(gsl_tmp_149);
  let gsl_tmp_151 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_148.x, gsl_tmp_148.y, gsl_tmp_150))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_148.x, gsl_tmp_148.y, gsl_tmp_150)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let nearestPore = select(gsl_tmp_147, min(gsl_tmp_147, length(vec2<f32>(length(poreQ - (gsl_tmp_148 + gsl_tmp_151.xy)), max(abs(gsl_tmp_149 - gsl_tmp_150 - 0.5) - 0.2 - 0.25 * gsl_tmp_151.x, 0.0) * 8.0)) / size), gsl_tmp_151.z < poreShare);
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
  let gsl_tmp_278 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010937121 * sin(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.017413743 * sin(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020270431 * sin(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.02272972 * sin(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023828013 * sin(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022254365 * sin(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.01769435 * sin(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.0116398325 * sin(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
  let gsl_tmp_279 = 0.0 + 0.018228535 * cos(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_278) - 0.64525276 * time + 0.0) + 0.029022904 * cos(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_278) - 0.73128647 * time + 2.1) + 0.03378405 * cos(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_278) - 0.81732017 * time + 4.4) + 0.03788287 * cos(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_278) - 0.90335387 * time + 1.3) + 0.039713353 * cos(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_278) - 1.0324044 * time + 5.2) + 0.03709061 * cos(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_278) - 1.2044718 * time + 3.3) + 0.029490585 * cos(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_278) - 1.4625729 * time + 0.7) + 0.019399721 * cos(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_278) - 1.8067077 * time + 2.9);
  let waterHeight = gsl_tmp_279;
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
    var gsl_detailValue = vec3<f32>(0.4217793, 0.42321503, 0.42755702);
    if (max(gsl_pw, gsl_detailPw) < 1.3) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_1 = p / 4.0;
      let gsl_tmp_3 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).nearCell + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_cellularScan(gsl_tmp_1, 1, noiseTex).nearCell + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_4 = (gsl_tmp_3.y - 0.5) * 0.14;
      let gsl_tmp_2 = gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1 * 4.0;
      let gsl_tmp_5 = 1.0 - texture_mask(gsl_tmp_2 / 4.0, 0.35, 0.25, noiseTex);
      gsl_detailValue = color_oklab2rgb(vec3<f32>(0.53104144, 0.00004615242, -0.0017715977) + (select(select(vec3<f32>(0.5981807, 0.000000000048424487, 0.000000022296533) + vec3<f32>(gsl_tmp_4, 0.0, 0.0), vec3<f32>(0.78843343, 0.0009858413, -0.0034093664) + vec3<f32>(gsl_tmp_4 * 0.5 - 0.12 * gsl_tmp_5 - 0.15 * select(0.0, 1.0, fract(textureLoad(noiseTex, vec3<i32>(floor(floor(p / 0.24) + vec3<f32>(313.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((floor(p / 0.24) + vec3<f32>(313.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x > 0.96) * (1.0 - texture_mask(gsl_pw, 0.09, 0.03, noiseTex)), -0.004 * gsl_tmp_5, -0.012 * gsl_tmp_5), gsl_tmp_3.x > 0.55), vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(gsl_tmp_4 * 0.4, 0.0, 0.0), gsl_tmp_3.x > 0.85) - vec3<f32>(0.53104144, 0.00004615242, -0.0017715977)) * (1.0 - texture_mask(gsl_pw, 0.975, 0.325, noiseTex)) - vec3<f32>(0.14 * (1.0 - texture_mask(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f2 * 4.0 - gsl_tmp_2, 0.2, 0.12, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.15, 0.05, noiseTex)), 0.0, 0.0), noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue_2 = vec3<f32>(0.4217793, 0.42321503, 0.42755702);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, gsl_detailValue, gsl_simpleDetail);
    }
    let graniteColor = gsl_simpleValue_2;
    return vec4<f32>(graniteColor, 1.0);
  }
  if (gsl_colorIdx == 3u) {
    let gsl_tmp_329 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
    let gsl_tmp_330 = length(gsl_tmp_329.xy);
    let gsl_tmp_331 = gsl_tmp_329.xy / max(gsl_tmp_330, 0.001) * 40.0;
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    let color_gray = vec3<f32>(0.5, 0.5, 0.5);
    let gsl_tmp_302 = vec2<f32>(gsl_tmp_330 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_331.x, gsl_tmp_331.y, gsl_tmp_329.z * 0.25), color_white, color_gray, noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_329.x, gsl_tmp_329.y, gsl_tmp_329.z * 0.2), color_white, color_gray, noiseTex), 0.0);
    let gsl_tmp_303 = gsl_tmp_302.x;
    let gsl_tmp_304 = floor(gsl_tmp_303);
    let gsl_tmp_305 = wood_ringBoundary(gsl_tmp_304, 3.0, noiseTex);
    let gsl_tmp_306 = wood_ringBoundary(gsl_tmp_304 + 1.0, 3.0, noiseTex);
    let gsl_tmp_308 = gsl_tmp_303 >= gsl_tmp_306;
    let gsl_tmp_307 = gsl_tmp_303 < gsl_tmp_305;
    let gsl_tmp_309 = select(select(gsl_tmp_305, gsl_tmp_306, gsl_tmp_308), wood_ringBoundary(gsl_tmp_304 - 1.0, 3.0, noiseTex), gsl_tmp_307);
    let gsl_tmp_311 = (gsl_tmp_303 - gsl_tmp_309) / (select(select(gsl_tmp_306, wood_ringBoundary(gsl_tmp_304 + 2.0, 3.0, noiseTex), gsl_tmp_308), gsl_tmp_305, gsl_tmp_307) - gsl_tmp_309);
    let gsl_tmp_310 = select(select(gsl_tmp_304, gsl_tmp_304 + 1.0, gsl_tmp_308), gsl_tmp_304 - 1.0, gsl_tmp_307);
    let gsl_tmp_312 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_310, gsl_tmp_310 + 11.0, 3.0 * gsl_tmp_310 + 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_310, gsl_tmp_310 + 11.0, 3.0 * gsl_tmp_310 + 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
    let gsl_tmp_313 = texture_mask(gsl_tmp_311, 0.4 + (gsl_tmp_312.x - 0.5) * 0.027, 0.3, noiseTex) * (1.0 + (gsl_tmp_312.y - 0.5) * 0.18);
    let gsl_tmp_300 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
    let gsl_tmp_301 = length(gsl_tmp_300.xy);
    var gsl_detailValue_2 = 0.0;
    if (max(gsl_pw, gsl_detailPw) < 0.075) { let gsl_detailOuterPw_2 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_314 = (atan2(gsl_tmp_300.y, gsl_tmp_300.x) + 3.1415927) / 6.2831855 * 3351.0 - 0.5;
      let gsl_tmp_315 = floor(gsl_tmp_314);
      let gsl_tmp_317 = wood_fibreFile(gsl_floored(gsl_tmp_315, 3351.0), gsl_tmp_301, gsl_tmp_300.z, noiseTex);
      let gsl_tmp_316 = gsl_tmp_314 - gsl_tmp_315;
      gsl_detailValue_2 = (gsl_tmp_317 + (wood_fibreFile(gsl_floored((gsl_tmp_315 + 1.0), 3351.0), gsl_tmp_301, gsl_tmp_300.z, noiseTex) - gsl_tmp_317) * gsl_tmp_316 * gsl_tmp_316 * (3.0 - 2.0 * gsl_tmp_316) - 0.5) * 0.08 * (1.0 - 0.6 * gsl_tmp_313) * texture_fadeDetail(min(0.15, 6.2831855 * gsl_tmp_301 / 3351.0), noiseTex);
      gsl_pw = gsl_detailOuterPw_2;
    }
    let gsl_tmp_320 = wood_poreSizeShare__spec1(gsl_tmp_311, 0.3, noiseTex).x;
    var gsl_detailValue_3 = 0.05 * gsl_tmp_320 * gsl_tmp_320;
    if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_3 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_318 = wood_poreDistanceAt(gsl_tmp_300, gsl_tmp_311, 0.5, 1.0, 0.3, noiseTex);
      let gsl_tmp_319 = texture_fadeDetail(0.28 * gsl_tmp_318.y, noiseTex);
      gsl_detailValue_3 = (1.0 - texture_mask(gsl_tmp_318.x, 0.28, 0.06, noiseTex)) * gsl_tmp_319 + 0.05 * gsl_tmp_318.y * gsl_tmp_318.y * (1.0 - gsl_tmp_319);
      gsl_pw = gsl_detailOuterPw_3;
    }
    var gsl_simpleValue_3 = vec3<f32>(0.59386784, 0.47988868, 0.3859136);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_321 = (atan2(gsl_tmp_300.y, gsl_tmp_300.x) + 3.1415927) / 6.2831855 * 250.0;
      let gsl_tmp_322 = gsl_floored(floor(gsl_tmp_321), 250.0);
      let gsl_tmp_323 = gsl_tmp_300.z / 12.0 + fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_322, gsl_tmp_322 + 17.0, 3.0 * gsl_tmp_322 + 5.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_322, gsl_tmp_322 + 17.0, 3.0 * gsl_tmp_322 + 5.0)) / 64.0)) & vec3<i32>(63), 0).rgb).y;
      let gsl_tmp_324 = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_322, floor(gsl_tmp_323), gsl_tmp_322 + 401.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_322, floor(gsl_tmp_323), gsl_tmp_322 + 401.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_325 = vec3<f32>(0.6289975, 0.025104912, 0.04613428) + vec3<f32>(-0.044705804, -0.0014734669, -0.0036211603) * gsl_tmp_313 + vec3<f32>((fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(gsl_tmp_310, gsl_tmp_310 + 5.0, 2.0 * gsl_tmp_310 + 29.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(gsl_tmp_310, gsl_tmp_310 + 5.0, 2.0 * gsl_tmp_310 + 29.0)) / 64.0)) & vec3<i32>(63), 0).rgb).x - 0.5) * 0.08 + gsl_detailValue_2 + 0.1 * 1.0 * select(0.0, 1.0, gsl_tmp_324.z < 0.5) * (1.0 - texture_mask(abs(gsl_tmp_321 - floor(gsl_tmp_321) - 0.2 - 0.6 * gsl_tmp_324.x) * 6.2831855 * gsl_tmp_301 / 250.0, 0.15, 0.05, noiseTex)) * (1.0 - texture_mask(abs(gsl_tmp_323 - floor(gsl_tmp_323) - 0.5) * 2.0 / (0.35 + 0.65 * gsl_tmp_324.y), 1.0, 0.1, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.1125, 0.0375, noiseTex)) + 0.0, 0.0, 0.0) + vec3<f32>(0.042500425, 0.00060789636, -0.003918465) * texture_mask(gsl_tmp_302.x, 999979.3, 0.3, noiseTex);
      let gsl_tmp_326 = gsl_tmp_325 + (vec3<f32>(0.0, 0.0, 0.0) - gsl_tmp_325) * 0.0 * (1.0 - texture_mask(abs(gsl_tmp_302.x - 999979.3 + 0.3), 0.3, 0.1, noiseTex));
      let gsl_tmp_327 = gsl_tmp_326 + (vec3<f32>(0.0, 0.0, 0.0) - gsl_tmp_326) * 0.0;
      let gsl_tmp_328 = gsl_tmp_327 + (vec3<f32>(0.45476237, 0.037886485, 0.05228309) - gsl_tmp_327) * 0.85 * gsl_detailValue_3 - vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_261 = color_oklab2rgb(gsl_tmp_328 + (vec3<f32>(0.41366172, 0.0160504, 0.030504888) - gsl_tmp_328) * 0.8 * gsl_tmp_302.y, noiseTex);
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, gsl_tmp_261, gsl_simpleDetail);
    }
    let oakColor = gsl_simpleValue_3;
    return vec4<f32>(oakColor, 1.0);
  }
  if (gsl_colorIdx == 4u) {
    var gsl_simpleValue_4 = vec3<f32>(0.68587554, 0.542306, 0.49029985);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_7 = noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      gsl_simpleValue_4 = mix(gsl_simpleValue_4, texture_paint(gsl_tmp_7, 0.0, 0.5, vec3<f32>(0.77, 0.78, 0.78), vec3<f32>(0.55, 0.248875, 0.1375), noiseTex), gsl_simpleDetail);
    }
    let rustedColor = gsl_simpleValue_4;
    return vec4<f32>(rustedColor, 1.0);
  }
  if (gsl_colorIdx == 5u) {
    var gsl_simpleValue_5 = vec3<f32>(0.75119495, 0.75640696, 0.72030866);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_9 = noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      gsl_simpleValue_5 = mix(gsl_simpleValue_5, texture_paint(gsl_tmp_9, 0.0, 0.5, vec3<f32>(0.98, 0.82, 0.76), vec3<f32>(0.44527867, 0.6713162, 0.66820365), noiseTex), gsl_simpleDetail);
    }
    let patinaColor = gsl_simpleValue_5;
    return vec4<f32>(patinaColor, 1.0);
  }
  if (gsl_colorIdx == 6u) {
    var gsl_simpleValue_6 = vec3<f32>(0.0032853691, 0.003285369, 0.003285368);
    if (gsl_simpleDetail > 0.0) {
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let gsl_tmp_10 = p / 1.0;
      let gsl_tmp_11 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_10, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      gsl_simpleValue_6 = mix(gsl_simpleValue_6, color_mixOklab(color_black, color_white, gsl_tmp_11, noiseTex), gsl_simpleDetail);
    }
    let glitterColor = gsl_simpleValue_6;
    return vec4<f32>(glitterColor, 1.0);
  }
  if (gsl_colorIdx == 7u) {
    var gsl_simpleValue_7 = vec3<f32>(0.38857287, 0.38857284, 0.3885728);
    if (gsl_simpleDetail > 0.0) {
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let gsl_tmp_43 = texture_squareWave(p.x, 20.0, noiseTex);
      let gsl_tmp_44 = texture_squareWave(p.y, 20.0, noiseTex);
      let gsl_tmp_46 = gsl_tmp_43 + gsl_tmp_44 - 2.0 * gsl_tmp_43 * gsl_tmp_44;
      let gsl_tmp_45 = texture_squareWave(p.z, 20.0, noiseTex);
      gsl_simpleValue_7 = mix(gsl_simpleValue_7, color_mixOklab(color_white, color_black, gsl_tmp_46 + gsl_tmp_45 - 2.0 * gsl_tmp_46 * gsl_tmp_45, noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_7, 1.0);
  }
  if (gsl_colorIdx == 8u) {
    var gsl_simpleValue_8 = vec3<f32>(0.048666414, 0.1840356, 0.30572936);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_278 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010937121 * sin(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.017413743 * sin(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020270431 * sin(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.02272972 * sin(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023828013 * sin(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022254365 * sin(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.01769435 * sin(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.0116398325 * sin(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
      let gsl_tmp_279 = 0.0 + 0.018228535 * cos(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_278) - 0.64525276 * time + 0.0) + 0.029022904 * cos(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_278) - 0.73128647 * time + 2.1) + 0.03378405 * cos(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_278) - 0.81732017 * time + 4.4) + 0.03788287 * cos(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_278) - 0.90335387 * time + 1.3) + 0.039713353 * cos(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_278) - 1.0324044 * time + 5.2) + 0.03709061 * cos(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_278) - 1.2044718 * time + 3.3) + 0.029490585 * cos(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_278) - 1.4625729 * time + 0.7) + 0.019399721 * cos(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_278) - 1.8067077 * time + 2.9);
      let gsl_tmp_280 = texture_mask(gsl_tmp_279 + 0.037828535 * texture_fadeTo(noise_gradientLattice((p - vec3<f32>(0.95782626, 0.28734788, 0.0) * 1.368301 * time) / 0.49964657 + vec3<f32>(13.0, 71.0, 5.0), noiseTex), 0.0, 0.49964657, noiseTex), 0.14700663, 0.018914267, noiseTex);
      gsl_simpleValue_8 = mix(gsl_simpleValue_8, color_mixOklab(color_mixOklab(vec3<f32>(0.04, 0.16, 0.29), vec3<f32>(0.1, 0.42, 0.44), 0.5 * texture_mask(gsl_tmp_279, 0.06304756, 0.06304756, noiseTex), noiseTex), vec3<f32>(0.92, 0.94, 0.94), gsl_tmp_280, noiseTex), gsl_simpleDetail);
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
      let gsl_tmp_297 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
      let gsl_tmp_298 = length(gsl_tmp_297.xy);
      let gsl_tmp_299 = gsl_tmp_297.xy / max(gsl_tmp_298, 0.001) * 40.0;
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let color_gray = vec3<f32>(0.5, 0.5, 0.5);
      let gsl_tmp_281 = gsl_tmp_298 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_299.x, gsl_tmp_299.y, gsl_tmp_297.z * 0.25), color_white, color_gray, noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_297.x, gsl_tmp_297.y, gsl_tmp_297.z * 0.2), color_white, color_gray, noiseTex);
      let gsl_tmp_282 = floor(gsl_tmp_281);
      let gsl_tmp_283 = wood_ringBoundary(gsl_tmp_282, 3.0, noiseTex);
      let gsl_tmp_284 = wood_ringBoundary(gsl_tmp_282 + 1.0, 3.0, noiseTex);
      let gsl_tmp_286 = gsl_tmp_281 >= gsl_tmp_284;
      let gsl_tmp_285 = gsl_tmp_281 < gsl_tmp_283;
      let gsl_tmp_287 = select(select(gsl_tmp_283, gsl_tmp_284, gsl_tmp_286), wood_ringBoundary(gsl_tmp_282 - 1.0, 3.0, noiseTex), gsl_tmp_285);
      gsl_detailValue_4 = wood_poreSurfaceAt(map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex), (gsl_tmp_281 - gsl_tmp_287) / (select(select(gsl_tmp_284, wood_ringBoundary(gsl_tmp_282 + 2.0, 3.0, noiseTex), gsl_tmp_286), gsl_tmp_283, gsl_tmp_285) - gsl_tmp_287), 1.5, 0.5, 1.0, 0.0, noiseTex) * 1.0;
      gsl_pw = gsl_detailOuterPw_4;
    }
    var gsl_simpleValue_11 = mat4x4<f32>(vec4<f32>(0.6685386, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_253 = gsl_detailValue_4;
      let gsl_tmp_255 = 0.5851758 + 0.3148242 * gsl_tmp_253.y;
      let gsl_tmp_256 = vec2<f32>(gsl_tmp_255, gsl_tmp_253.x);
      let gsl_simpleFull = material_withDisplacement(wood_woodFinish(sqrt(sqrt(gsl_tmp_256.x * gsl_tmp_256.x * gsl_tmp_256.x * gsl_tmp_256.x + 0.0825 * material_hiddenShare__spec1(1.0 - texture_mask(gsl_pw, 0.105, 0.035, noiseTex), noiseTex))), noiseTex), gsl_tmp_256.y, noiseTex);
      gsl_simpleValue_11 = mat4x4<f32>(mix(gsl_simpleValue_11[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_11[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_11[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_11[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let oakMat = gsl_simpleValue_11;
    let gsl_m = oakMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 4u) {
    var gsl_simpleValue_12 = mat4x4<f32>(vec4<f32>(0.66031295, 0.55252564, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_7 = noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_simpleFull_2 = texture_paintMaterial(gsl_tmp_7, 0.0, 0.5, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
      gsl_simpleValue_12 = mat4x4<f32>(mix(gsl_simpleValue_12[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_12[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_12[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_12[3], gsl_simpleFull_2[3], gsl_simpleDetail));
    }
    let rustedMat = gsl_simpleValue_12;
    let gsl_m = rustedMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 5u) {
    var gsl_simpleValue_13 = mat4x4<f32>(vec4<f32>(0.6050604, 0.55252564, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_9 = noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
      let gsl_simpleFull_3 = texture_paintMaterial(gsl_tmp_9, 0.0, 0.5, mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
      gsl_simpleValue_13 = mat4x4<f32>(mix(gsl_simpleValue_13[0], gsl_simpleFull_3[0], gsl_simpleDetail), mix(gsl_simpleValue_13[1], gsl_simpleFull_3[1], gsl_simpleDetail), mix(gsl_simpleValue_13[2], gsl_simpleFull_3[2], gsl_simpleDetail), mix(gsl_simpleValue_13[3], gsl_simpleFull_3[3], gsl_simpleDetail));
    }
    let patinaMat = gsl_simpleValue_13;
    let gsl_m = patinaMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 6u) {
    let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    var gsl_simpleValue_14 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, material_polishedMetal[1] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[1]) * 0.06335398, material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_10 = p / 1.0;
      let gsl_tmp_11 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_10, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      let gsl_simpleFull_4 = material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_11, noiseTex);
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
      let gsl_tmp_278 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010937121 * sin(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.017413743 * sin(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020270431 * sin(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.02272972 * sin(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023828013 * sin(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022254365 * sin(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.01769435 * sin(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.0116398325 * sin(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
      let gsl_tmp_279 = 0.0 + 0.018228535 * cos(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_278) - 0.64525276 * time + 0.0) + 0.029022904 * cos(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_278) - 0.73128647 * time + 2.1) + 0.03378405 * cos(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_278) - 0.81732017 * time + 4.4) + 0.03788287 * cos(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_278) - 0.90335387 * time + 1.3) + 0.039713353 * cos(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_278) - 1.0324044 * time + 5.2) + 0.03709061 * cos(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_278) - 1.2044718 * time + 3.3) + 0.029490585 * cos(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_278) - 1.4625729 * time + 0.7) + 0.019399721 * cos(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_278) - 1.8067077 * time + 2.9);
      let gsl_tmp_280 = texture_mask(gsl_tmp_279 + 0.037828535 * texture_fadeTo(noise_gradientLattice((p - vec3<f32>(0.95782626, 0.28734788, 0.0) * 1.368301 * time) / 0.49964657 + vec3<f32>(13.0, 71.0, 5.0), noiseTex), 0.0, 0.49964657, noiseTex), 0.14700663, 0.018914267, noiseTex);
      let gsl_simpleFull_5 = material_blend(mat4x4<f32>(vec4<f32>(0.23280893, 0.0, 1.34, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.7, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_280, noiseTex);
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

const gsl_foldK_2 = array<vec2<f32>, 9>(vec2<f32>(-1.0, -1.0), vec2<f32>(0.0, -1.0), vec2<f32>(1.0, -1.0), vec2<f32>(-1.0, 0.0), vec2<f32>(0.0, 0.0), vec2<f32>(1.0, 0.0), vec2<f32>(-1.0, 1.0), vec2<f32>(0.0, 1.0), vec2<f32>(1.0, 1.0));

fn gsl_heights(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> array<f32, 10> {
  var gsl_detailValue_5 = vec2<f32>(0.0, 0.0);
  let gsl_tmp_297 = map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex);
  let gsl_tmp_298 = length(gsl_tmp_297.xy);
  let gsl_tmp_299 = gsl_tmp_297.xy / max(gsl_tmp_298, 0.001) * 40.0;
  let color_white = vec3<f32>(1.0, 1.0, 1.0);
  let color_gray = vec3<f32>(0.5, 0.5, 0.5);
  let gsl_tmp_281 = gsl_tmp_298 * (1.0 + 0.08 * gsl_remap(vec3<f32>(gsl_tmp_299.x, gsl_tmp_299.y, gsl_tmp_297.z * 0.25), color_white, color_gray, noiseTex)) / 1.5 + 0.3 * gsl_remap_2(vec3<f32>(gsl_tmp_297.x, gsl_tmp_297.y, gsl_tmp_297.z * 0.2), color_white, color_gray, noiseTex);
  let gsl_tmp_282 = floor(gsl_tmp_281);
  let gsl_tmp_283 = wood_ringBoundary(gsl_tmp_282, 3.0, noiseTex);
  let gsl_tmp_284 = wood_ringBoundary(gsl_tmp_282 + 1.0, 3.0, noiseTex);
  let gsl_tmp_286 = gsl_tmp_281 >= gsl_tmp_284;
  let gsl_tmp_285 = gsl_tmp_281 < gsl_tmp_283;
  let gsl_tmp_287 = select(select(gsl_tmp_283, gsl_tmp_284, gsl_tmp_286), wood_ringBoundary(gsl_tmp_282 - 1.0, 3.0, noiseTex), gsl_tmp_285);
  if (max(gsl_pw, gsl_detailPw) < 0.14) { let gsl_detailOuterPw_5 = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
    gsl_detailValue_5 = wood_poreSurfaceAt(map_axisAt(p, vec3<f32>(-187.5, 0.0, -25.0), vec3<f32>(0.0, 0.0, 1.0), noiseTex), (gsl_tmp_281 - gsl_tmp_287) / (select(select(gsl_tmp_284, wood_ringBoundary(gsl_tmp_282 + 2.0, 3.0, noiseTex), gsl_tmp_286), gsl_tmp_283, gsl_tmp_285) - gsl_tmp_287), 1.5, 0.5, 1.0, 0.0, noiseTex) * 1.0;
    gsl_pw = gsl_detailOuterPw_5;
  }
  let gsl_tmp_253 = gsl_detailValue_5;
  let gsl_tmp_255 = 0.5851758 + 0.3148242 * gsl_tmp_253.y;
  let gsl_tmp_256 = vec2<f32>(gsl_tmp_255, gsl_tmp_253.x);
  var gsl_simpleValue_17 = mat4x4<f32>(vec4<f32>(0.6685386, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_7 = material_withDisplacement(wood_woodFinish(sqrt(sqrt(gsl_tmp_256.x * gsl_tmp_256.x * gsl_tmp_256.x * gsl_tmp_256.x + 0.0825 * material_hiddenShare__spec1(1.0 - texture_mask(gsl_pw, 0.105, 0.035, noiseTex), noiseTex))), noiseTex), gsl_tmp_256.y, noiseTex);
    gsl_simpleValue_17 = mat4x4<f32>(mix(gsl_simpleValue_17[0], gsl_simpleFull_7[0], gsl_simpleDetail), mix(gsl_simpleValue_17[1], gsl_simpleFull_7[1], gsl_simpleDetail), mix(gsl_simpleValue_17[2], gsl_simpleFull_7[2], gsl_simpleDetail), mix(gsl_simpleValue_17[3], gsl_simpleFull_7[3], gsl_simpleDetail));
  }
  let oakMat = gsl_simpleValue_17;
  let gsl_tmp_7 = noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
  var gsl_simpleValue_18 = mat4x4<f32>(vec4<f32>(0.66031295, 0.55252564, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_8 = texture_paintMaterial(gsl_tmp_7, 0.0, 0.5, mat4x4<f32>(vec4<f32>(0.45, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
    gsl_simpleValue_18 = mat4x4<f32>(mix(gsl_simpleValue_18[0], gsl_simpleFull_8[0], gsl_simpleDetail), mix(gsl_simpleValue_18[1], gsl_simpleFull_8[1], gsl_simpleDetail), mix(gsl_simpleValue_18[2], gsl_simpleFull_8[2], gsl_simpleDetail), mix(gsl_simpleValue_18[3], gsl_simpleFull_8[3], gsl_simpleDetail));
  }
  let rustedMat = gsl_simpleValue_18;
  let gsl_tmp_9 = noise_gradientLattice(p / 15.0, noiseTex) + 0.5 * noise_gradientLattice(p / 7.5, noiseTex) + 0.25 * noise_gradientLattice(p / 3.75, noiseTex) - texture_fadeTo(noise_cellularEdges(p, 1.5, noiseTex) / 1.5, 0.199, 1.5, noiseTex) * 0.24;
  var gsl_simpleValue_19 = mat4x4<f32>(vec4<f32>(0.6050604, 0.55252564, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_9 = texture_paintMaterial(gsl_tmp_9, 0.0, 0.5, mat4x4<f32>(vec4<f32>(0.35, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.92, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), noiseTex);
    gsl_simpleValue_19 = mat4x4<f32>(mix(gsl_simpleValue_19[0], gsl_simpleFull_9[0], gsl_simpleDetail), mix(gsl_simpleValue_19[1], gsl_simpleFull_9[1], gsl_simpleDetail), mix(gsl_simpleValue_19[2], gsl_simpleFull_9[2], gsl_simpleDetail), mix(gsl_simpleValue_19[3], gsl_simpleFull_9[3], gsl_simpleDetail));
  }
  let patinaMat = gsl_simpleValue_19;
  let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  let gsl_tmp_10 = p / 1.0;
  let gsl_tmp_11 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_10, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
  var gsl_simpleValue_20 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, material_polishedMetal[1] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[1]) * 0.06335398, material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_10 = material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_11, noiseTex);
    gsl_simpleValue_20 = mat4x4<f32>(mix(gsl_simpleValue_20[0], gsl_simpleFull_10[0], gsl_simpleDetail), mix(gsl_simpleValue_20[1], gsl_simpleFull_10[1], gsl_simpleDetail), mix(gsl_simpleValue_20[2], gsl_simpleFull_10[2], gsl_simpleDetail), mix(gsl_simpleValue_20[3], gsl_simpleFull_10[3], gsl_simpleDetail));
  }
  let glitterMat = gsl_simpleValue_20;
  let gsl_tmp_278 = vec3<f32>(0.0, 0.0, 0.0) + vec3<f32>(0.94226784, 0.33486015, 0.0) * 0.010937121 * sin(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p) - 0.64525276 * time + 0.0) + vec3<f32>(0.9991408, 0.04144492, 0.0) * 0.017413743 * sin(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p) - 0.73128647 * time + 2.1) + vec3<f32>(0.83012927, 0.55757093, 0.0) * 0.020270431 * sin(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p) - 0.81732017 * time + 4.4) + vec3<f32>(0.9817281, 0.19028927, 0.0) * 0.02272972 * sin(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p) - 0.90335387 * time + 1.3) + vec3<f32>(0.73748547, 0.675363, 0.0) * 0.023828013 * sin(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p) - 1.0324044 * time + 5.2) + vec3<f32>(0.96676344, -0.25567245, 0.0) * 0.022254365 * sin(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p) - 1.2044718 * time + 3.3) + vec3<f32>(0.54747134, 0.8368244, 0.0) * 0.01769435 * sin(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p) - 1.4625729 * time + 0.7) + vec3<f32>(0.84802806, -0.52995133, 0.0) * 0.0116398325 * sin(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p) - 1.8067077 * time + 2.9);
  let gsl_tmp_279 = 0.0 + 0.018228535 * cos(0.35367915 * dot(vec3<f32>(0.94226784, 0.33486015, 0.0), p + gsl_tmp_278) - 0.64525276 * time + 0.0) + 0.029022904 * cos(0.45428124 * dot(vec3<f32>(0.9991408, 0.04144492, 0.0), p + gsl_tmp_278) - 0.73128647 * time + 2.1) + 0.03378405 * cos(0.56745857 * dot(vec3<f32>(0.83012927, 0.55757093, 0.0), p + gsl_tmp_278) - 0.81732017 * time + 4.4) + 0.03788287 * cos(0.69321114 * dot(vec3<f32>(0.9817281, 0.19028927, 0.0), p + gsl_tmp_278) - 0.90335387 * time + 1.3) + 0.039713353 * cos(0.90541863 * dot(vec3<f32>(0.73748547, 0.675363, 0.0), p + gsl_tmp_278) - 1.0324044 * time + 5.2) + 0.03709061 * cos(1.2323754 * dot(vec3<f32>(0.96676344, -0.25567245, 0.0), p + gsl_tmp_278) - 1.2044718 * time + 3.3) + 0.029490585 * cos(1.817125 * dot(vec3<f32>(0.54747134, 0.8368244, 0.0), p + gsl_tmp_278) - 1.4625729 * time + 0.7) + 0.019399721 * cos(2.7728446 * dot(vec3<f32>(0.84802806, -0.52995133, 0.0), p + gsl_tmp_278) - 1.8067077 * time + 2.9);
  let gsl_tmp_280 = texture_mask(gsl_tmp_279 + 0.037828535 * texture_fadeTo(noise_gradientLattice((p - vec3<f32>(0.95782626, 0.28734788, 0.0) * 1.368301 * time) / 0.49964657 + vec3<f32>(13.0, 71.0, 5.0), noiseTex), 0.0, 0.49964657, noiseTex), 0.14700663, 0.018914267, noiseTex);
  var gsl_simpleValue_21 = mat4x4<f32>(vec4<f32>(0.2354835, 0.0, 1.3399427, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_11 = material_blend(mat4x4<f32>(vec4<f32>(0.23280893, 0.0, 1.34, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.7, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_280, noiseTex);
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
